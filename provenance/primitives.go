// Package provenance implements independent Note source-recognition proofs.
package provenance

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strconv"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/invisi-lab/invisi-sdk-go/core"
)

var publicInputNames = [...]string{"chain_id", "protocol_address", "token_address", "policy_id_hi", "policy_id_lo", "chain_generation", "source_count", "sources_digest", "input_count", "operation_digest_hi", "operation_digest_lo", "input_digest", "binding"}
var canonicalInteger = regexp.MustCompile(`^(0|[1-9][0-9]*|0x[0-9a-fA-F]+)$`)
var modulus = fr.Modulus()
var domains = map[string]string{}

func init() {
	for k, v := range map[string]string{"scope": "Scope", "note": "Note", "inputs": "Inputs", "step": "Step", "sources": "Sources", "sourceStep": "SourceStep", "bind": "Bind"} {
		b := sha256.Sum256([]byte("Invisi.Provenance." + v + ".v1"))
		domains[k] = new(big.Int).Mod(new(big.Int).SetBytes(b[:]), modulus).String()
	}
}
func Uint(s string, bits int) (*big.Int, error) {
	if !canonicalInteger.MatchString(s) {
		return nil, errors.New("invalid integer encoding")
	}
	base := 10
	v := s
	if len(v) > 2 && v[:2] == "0x" {
		base = 16
		v = v[2:]
	}
	n, ok := new(big.Int).SetString(v, base)
	if !ok || n.Sign() < 0 || n.BitLen() > bits {
		return nil, fmt.Errorf("integer exceeds u%d", bits)
	}
	return n, nil
}
func Field(s string) (*big.Int, error) {
	n, e := Uint(s, 256)
	if e != nil {
		return nil, e
	}
	if n.Cmp(modulus) >= 0 {
		return nil, errors.New("noncanonical field")
	}
	return n, nil
}
func word(s string) (string, error) {
	n, e := Uint(s, 256)
	if e != nil {
		return "", e
	}
	return fmt.Sprintf("0x%064x", n), nil
}
func hash(v ...string) (string, error) {
	a := make([]fr.Element, len(v))
	for i, s := range v {
		n, e := Field(s)
		if e != nil {
			return "", e
		}
		a[i].SetBigInt(n)
	}
	r, e := core.Poseidon2Hash(a...)
	if e != nil {
		return "", e
	}
	return r.BigInt(new(big.Int)).String(), nil
}
func limbs(s string) (string, string, error) {
	n, e := Uint(s, 256)
	if e != nil {
		return "", "", e
	}
	return new(big.Int).Rsh(new(big.Int).Set(n), 128).String(), new(big.Int).And(n, new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1))).String(), nil
}
func check(s string, bits int, nonzero bool) error {
	n, e := Uint(s, bits)
	if e != nil {
		return e
	}
	if nonzero && n.Sign() == 0 {
		return errors.New("zero value")
	}
	return nil
}

type Scope struct {
	ChainID         string `json:"chainId"`
	ProtocolAddress string `json:"protocolAddress"`
	PolicyID        string `json:"policyId"`
}

func ScopeHash(s Scope) (string, error) {
	if e := check(s.ChainID, 64, false); e != nil {
		return "", e
	}
	if e := check(s.ProtocolAddress, 160, true); e != nil {
		return "", e
	}
	hi, lo, e := limbs(s.PolicyID)
	if e != nil {
		return "", e
	}
	return hash(domains["scope"], s.ChainID, s.ProtocolAddress, hi, lo)
}
func RecognizedLeaf(scope, commitment string) (string, error) {
	n, e := Field(commitment)
	if e != nil {
		return "", e
	}
	if n.Sign() == 0 {
		return "", errors.New("zero commitment")
	}
	return hash(domains["note"], scope, commitment)
}

type Opening struct {
	Owner  string `json:"owner"`
	Nonce  string `json:"nonce"`
	Amount string `json:"amount"`
}

func NoteOpening(s Scope, token string, o Opening) (commitment, nullifier string, err error) {
	for _, v := range []struct {
		s string
		b int
		n bool
	}{{s.ChainID, 64, false}, {s.ProtocolAddress, 160, true}, {token, 160, true}, {o.Owner, 160, true}, {o.Amount, 128, true}} {
		if e := check(v.s, v.b, v.n); e != nil {
			return "", "", e
		}
	}
	id, e := hash("2", o.Owner, o.Nonce)
	if e != nil {
		return "", "", e
	}
	c, e := hash("3", s.ChainID, s.ProtocolAddress, id, token, o.Amount)
	if e != nil {
		return "", "", e
	}
	n, e := hash("4", s.ChainID, s.ProtocolAddress, o.Nonce, c)
	return c, n, e
}
func InputDigest(nullifiers []string) (string, error) {
	d, e := hash(domains["inputs"], strconv.Itoa(len(nullifiers)))
	if e != nil {
		return "", e
	}
	for _, n := range nullifiers {
		d, e = hash(domains["step"], d, n)
		if e != nil {
			return "", e
		}
	}
	return d, nil
}

type Source struct {
	TreeID    string `json:"treeId"`
	Root      string `json:"root"`
	LeafCount string `json:"leafCount"`
}

func SourcesDigest(generation string, sources []Source) (string, error) {
	if e := check(generation, 64, false); e != nil {
		return "", e
	}
	d, e := hash(domains["sources"], generation, strconv.Itoa(len(sources)))
	if e != nil {
		return "", e
	}
	prior := new(big.Int)
	for _, s := range sources {
		id, e := Uint(s.TreeID, 64)
		if e != nil {
			return "", e
		}
		count, e := Uint(s.LeafCount, 64)
		if e != nil {
			return "", e
		}
		if id.Cmp(prior) <= 0 || count.Sign() == 0 || count.Cmp(new(big.Int).Lsh(big.NewInt(1), 32)) > 0 {
			return "", errors.New("invalid ordered source")
		}
		prior = id
		d, e = hash(domains["sourceStep"], d, s.TreeID, s.Root, s.LeafCount)
		if e != nil {
			return "", e
		}
	}
	return d, nil
}
func BindingDigest(scope, token, generation, sourceCount, sourcesDigest, inputCount, operationDigest, inputDigest string) (string, error) {
	if e := check(token, 160, true); e != nil {
		return "", e
	}
	if e := check(generation, 64, false); e != nil {
		return "", e
	}
	hi, lo, e := limbs(operationDigest)
	if e != nil {
		return "", e
	}
	return hash(domains["bind"], scope, token, generation, sourceCount, sourcesDigest, inputCount, hi, lo, inputDigest)
}
func OperationDigest(s Scope, capacity int, inputs []string) (string, error) {
	if e := check(s.ChainID, 64, false); e != nil {
		return "", e
	}
	if e := check(s.ProtocolAddress, 160, true); e != nil {
		return "", e
	}
	if capacity < 1 || capacity > 190 || len(inputs) != 17+3*capacity {
		return "", errors.New("invalid authenticated transact statement")
	}
	words := make([][32]byte, len(inputs))
	for i, v := range inputs {
		n, e := Field(v)
		if e != nil {
			return "", e
		}
		n.FillBytes(words[i][:])
	}
	var args abi.Arguments
	for _, t := range []string{"string", "uint64", "address", "uint8", "uint32", "bytes32[]"} {
		a, e := abi.NewType(t, "", nil)
		if e != nil {
			return "", e
		}
		args = append(args, abi.Argument{Type: a})
	}
	chain, _ := Uint(s.ChainID, 64)
	address, _ := Uint(s.ProtocolAddress, 160)
	encoded, e := args.Pack("Invisi.Provenance.Operation.v1", chain.Uint64(), common.BigToAddress(address), uint8(2), uint32(capacity), words)
	if e != nil {
		return "", e
	}
	return crypto.Keccak256Hash(encoded).Hex(), nil
}
func ProvenanceIntentHash(original, policy string) (string, error) {
	a, e := Uint(original, 256)
	if e != nil {
		return "", e
	}
	b, e := Uint(policy, 256)
	if e != nil {
		return "", e
	}
	var args abi.Arguments
	for _, t := range []string{"string", "bytes32", "bytes32"} {
		v, _ := abi.NewType(t, "", nil)
		args = append(args, abi.Argument{Type: v})
	}
	var x, y [32]byte
	a.FillBytes(x[:])
	b.FillBytes(y[:])
	encoded, e := args.Pack("InvisiBroadcasterProvenanceIntentV1", x, y)
	if e != nil {
		return "", e
	}
	return crypto.Keccak256Hash(encoded).Hex(), nil
}

type RootCertificate struct {
	CertificateVersion     string `json:"certificateVersion"`
	ProviderID             string `json:"providerId"`
	SigningKeyID           string `json:"signingKeyId"`
	PolicyID               string `json:"policyId"`
	ChainID                string `json:"chainId"`
	ProtocolAddress        string `json:"protocolAddress"`
	ChainGeneration        string `json:"chainGeneration"`
	RootSequence           string `json:"rootSequence"`
	TreeID                 string `json:"treeId"`
	LeafCount              string `json:"leafCount"`
	Root                   string `json:"root"`
	ConfirmedBlockNumber   string `json:"confirmedBlockNumber"`
	ConfirmedBlockHash     string `json:"confirmedBlockHash"`
	DecisionLogHash        string `json:"decisionLogHash"`
	PreviousCheckpointHash string `json:"previousCheckpointHash"`
	IssuedAt               string `json:"issuedAt"`
	Signature              string `json:"signature"`
}

func RootCertificatePayload(c RootCertificate) ([]byte, error) {
	if c.CertificateVersion != "1" {
		return nil, errors.New("certificate version")
	}
	if _, e := Field(c.Root); e != nil {
		return nil, e
	}
	if e := check(c.TreeID, 64, true); e != nil {
		return nil, e
	}
	count, e := Uint(c.LeafCount, 64)
	if e != nil || count.Cmp(new(big.Int).Lsh(big.NewInt(1), 32)) > 0 {
		return nil, errors.New("certificate leaf count")
	}
	out := []byte{}
	for _, v := range []struct {
		s string
		w int
	}{{c.CertificateVersion, 1}, {c.ProviderID, 32}, {c.SigningKeyID, 32}, {c.PolicyID, 32}, {c.ChainID, 8}, {c.ProtocolAddress, 20}, {c.ChainGeneration, 8}, {c.RootSequence, 8}, {c.TreeID, 8}, {c.LeafCount, 8}, {c.Root, 32}, {c.ConfirmedBlockNumber, 8}, {c.ConfirmedBlockHash, 32}, {c.DecisionLogHash, 32}, {c.PreviousCheckpointHash, 32}, {c.IssuedAt, 8}} {
		n, e := Uint(v.s, v.w*8)
		if e != nil {
			return nil, e
		}
		out = append(out, n.FillBytes(make([]byte, v.w))...)
	}
	return out, nil
}
func RootCertificateMessage(c RootCertificate) ([]byte, error) {
	b, e := RootCertificatePayload(c)
	if e != nil {
		return nil, e
	}
	h := sha256.Sum256(append([]byte("Invisi.Provenance.Root.v1"), b...))
	return h[:], nil
}
func CheckpointHash(c RootCertificate) (string, error) {
	b, e := RootCertificatePayload(c)
	if e != nil {
		return "", e
	}
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:]), nil
}
func VerifyRootCertificate(c RootCertificate, key ed25519.PublicKey) bool {
	if len(key) != 32 || len(c.Signature) != 130 || c.Signature[:2] != "0x" {
		return false
	}
	sig, e := hex.DecodeString(c.Signature[2:])
	if e != nil || len(sig) != 64 {
		return false
	}
	m, e := RootCertificateMessage(c)
	return e == nil && ed25519.Verify(key, m, sig)
}

func OperationDigestDecimal(s Scope, mainCircuitCapacity string, inputs []string) (string, error) {
	capacity, e := Uint(mainCircuitCapacity, 32)
	if e != nil {
		return "", e
	}
	return OperationDigest(s, int(capacity.Uint64()), inputs)
}

func PublicInputNames() []string { return append([]string(nil), publicInputNames[:]...) }
