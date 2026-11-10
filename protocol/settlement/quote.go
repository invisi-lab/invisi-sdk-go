package settlement

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// Quote uses decimal strings on the JSON wire, matching the TypeScript SDK.
type Quote struct {
	Token        string `json:"token"`
	FeeRecipient string `json:"feeRecipient"`
	FeeAmount    string `json:"feeAmount"`
	Deadline     string `json:"deadline"`
	MaxGas       string `json:"maxGas"`
	RequestHash  string `json:"requestHash"`
}
type ProvenanceTrackingMatch struct {
	TxHash         string `json:"txHash"`
	BlockHash      string `json:"blockHash"`
	OperationIndex string `json:"operationIndex"`
}
type ProvenanceTracking struct {
	OperationDigest    common.Hash              `json:"operationDigest"`
	ID                 string                   `json:"id"`
	Status             string                   `json:"status"`
	ExpiresAt          string                   `json:"expiresAt"`
	Matched            *ProvenanceTrackingMatch `json:"matched,omitempty"`
	ReasonCode         string                   `json:"reasonCode,omitempty"`
	CheckpointSequence string                   `json:"checkpointSequence,omitempty"`
}
type SignedQuote struct {
	ProvenanceRequired bool   `json:"provenanceRequired,omitempty"`
	RequestHashFormat  string `json:"requestHashFormat,omitempty"`
	Quote              Quote  `json:"quote"`
	Signature          string `json:"signature"`
}
type QuoteDomain struct {
	ChainID  *big.Int
	Protocol common.Address
}

var ErrQuote = errors.New("invalid broadcaster quote")
var quoteTypeHash = crypto.Keccak256Hash([]byte("Quote(address token,address feeRecipient,uint256 feeAmount,uint256 deadline,uint256 maxGas,bytes32 requestHash)"))
var domainTypeHash = crypto.Keccak256Hash([]byte("EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)"))

// Uint256 parses a decimal unsigned integer without truncation.
func Uint256(s string) (*big.Int, error) {
	n, ok := new(big.Int).SetString(s, 10)
	if !ok || n.Sign() < 0 || n.BitLen() > 256 {
		return nil, fmt.Errorf("%w: uint256", ErrQuote)
	}
	return n, nil
}

// Match ethers address parsing: mixed case carries an EIP-55 checksum.
func validAddress(s string) bool {
	if !common.IsHexAddress(s) || strings.HasPrefix(s, "0X") {
		return false
	}
	raw := strings.TrimPrefix(s, "0x")
	return raw == strings.ToLower(raw) || raw == strings.ToUpper(raw) || common.HexToAddress(s).Hex()[2:] == raw
}

func QuoteDigest(d QuoteDomain, q Quote) (common.Hash, error) {
	if d.ChainID == nil || d.ChainID.Sign() < 0 || d.ChainID.BitLen() > 256 || !validAddress(q.Token) || !validAddress(q.FeeRecipient) {
		return common.Hash{}, ErrQuote
	}
	fee, e := Uint256(q.FeeAmount)
	if e != nil {
		return common.Hash{}, e
	}
	deadline, e := Uint256(q.Deadline)
	if e != nil {
		return common.Hash{}, e
	}
	gas, e := Uint256(q.MaxGas)
	if e != nil {
		return common.Hash{}, e
	}
	hashBytes := common.FromHex(q.RequestHash)
	if len(hashBytes) != 32 || len(q.RequestHash) != 66 {
		return common.Hash{}, ErrQuote
	}
	makeType := func(t string) abi.Type {
		v, e := abi.NewType(t, "", nil)
		if e != nil {
			panic(e)
		}
		return v
	}
	word, address, integer := makeType("bytes32"), makeType("address"), makeType("uint256")
	domainArgs := abi.Arguments{{Type: word}, {Type: word}, {Type: word}, {Type: integer}, {Type: address}}
	encoded, e := domainArgs.Pack(domainTypeHash, crypto.Keccak256Hash([]byte("InvisiBroadcaster")), crypto.Keccak256Hash([]byte("1")), d.ChainID, d.Protocol)
	if e != nil {
		return common.Hash{}, e
	}
	domainHash := crypto.Keccak256(encoded)
	quoteArgs := abi.Arguments{{Type: word}, {Type: address}, {Type: address}, {Type: integer}, {Type: integer}, {Type: integer}, {Type: word}}
	encoded, e = quoteArgs.Pack(quoteTypeHash, common.HexToAddress(q.Token), common.HexToAddress(q.FeeRecipient), fee, deadline, gas, common.BytesToHash(hashBytes))
	if e != nil {
		return common.Hash{}, e
	}
	return crypto.Keccak256Hash([]byte{0x19, 0x01}, domainHash, crypto.Keccak256(encoded)), nil
}
func SignQuote(d QuoteDomain, q Quote, key *ecdsa.PrivateKey) ([]byte, error) {
	if key == nil {
		return nil, ErrQuote
	}
	h, e := QuoteDigest(d, q)
	if e != nil {
		return nil, e
	}
	sig, e := crypto.Sign(h[:], key)
	if e != nil {
		return nil, e
	}
	sig[64] += 27
	return sig, nil
}
func RecoverQuoteSigner(d QuoteDomain, q Quote, signature []byte) (common.Address, error) {
	h, e := QuoteDigest(d, q)
	if e != nil {
		return common.Address{}, e
	}
	sig := append([]byte(nil), signature...)
	if len(sig) == 64 {
		sig = append(sig, 0)
		sig[64] = sig[32] >> 7
		sig[32] &= 0x7f
	} else if len(sig) == 65 {
		if sig[64] >= 27 {
			sig[64] -= 27
		}
	} else {
		return common.Address{}, ErrQuote
	}
	if !crypto.ValidateSignatureValues(sig[64], new(big.Int).SetBytes(sig[:32]), new(big.Int).SetBytes(sig[32:64]), true) {
		return common.Address{}, ErrQuote
	}
	pub, e := crypto.SigToPub(h[:], sig)
	if e != nil {
		return common.Address{}, ErrQuote
	}
	return crypto.PubkeyToAddress(*pub), nil
}
func VerifyQuote(d QuoteDomain, q Quote, sig []byte, signer common.Address) error {
	actual, e := RecoverQuoteSigner(d, q, sig)
	if e != nil || actual != signer {
		return ErrQuote
	}
	return nil
}
