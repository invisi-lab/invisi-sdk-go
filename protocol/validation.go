package protocol

import (
	"fmt"
	"math/big"
	"reflect"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/invisi-lab/invisi-sdk-go/protocol/contracts"
)

func CheckUint(v *big.Int, bits int, name string) error {
	if v == nil || v.Sign() < 0 || v.BitLen() > bits {
		return fmt.Errorf("%s must be uint%d", name, bits)
	}
	return nil
}
func Field(v [32]byte) error {
	if new(big.Int).SetBytes(v[:]).Cmp(fr.Modulus()) >= 0 {
		return fmt.Errorf("noncanonical BN254 Field")
	}
	return nil
}
func ValidateContext(c TransactionContext) error {
	if err := CheckUint(c.ChainID, 64, "chainID"); err != nil {
		return err
	}
	if c.Now != nil {
		return CheckUint(c.Now, 64, "now")
	}
	return nil
}
func CheckDeadline(c TransactionContext, d *big.Int) error {
	if err := CheckUint(d, 64, "deadline"); err != nil {
		return err
	}
	if c.Now != nil && d.Cmp(c.Now) < 0 {
		return fmt.Errorf("deadline has expired")
	}
	return nil
}
func nonzero(v [32]byte) bool { return v != [32]byte{} }
func ValidateNotes(values [][32]byte) error {
	seen := map[[32]byte]bool{}
	for _, v := range values {
		if err := Field(v); err != nil {
			return err
		}
		if !nonzero(v) || seen[v] {
			return fmt.Errorf("duplicate or zero note/nullifier")
		}
		seen[v] = true
	}
	return nil
}
func NormalizeSources(s []InputSource, n, capacity int) ([]InputSource, error) {
	if n < 0 || n > capacity || len(s) > n || (n == 0 && len(s) != 0) || (n > 0 && len(s) == 0) {
		return nil, fmt.Errorf("source count must be 1..input count, or zero for empty inputs")
	}
	result := make([]InputSource, len(s))
	previous := new(big.Int)
	for i, v := range s {
		if err := CheckUint(v.TreeId, 64, "treeId"); err != nil {
			return nil, err
		}
		if err := Field(v.MerkleRoot); err != nil {
			return nil, err
		}
		if v.TreeId.Sign() == 0 || v.TreeId.Cmp(previous) <= 0 || !nonzero(v.MerkleRoot) {
			return nil, fmt.Errorf("sources require positive roots and strictly ascending tree IDs")
		}
		result[i] = InputSource{TreeId: new(big.Int).Set(v.TreeId), MerkleRoot: v.MerkleRoot}
		previous = result[i].TreeId
	}
	return result, nil
}
func zeroIfNil(v *big.Int) *big.Int {
	if v == nil {
		return new(big.Int)
	}
	return new(big.Int).Set(v)
}

// NormalizeFunding expands the zero Go value to the canonical empty funding tuple.
func NormalizeFunding(f FundingParams) (FundingParams, error) {
	f.Amount = zeroIfNil(f.Amount)
	f.SigNonce = zeroIfNil(f.SigNonce)
	f.Deadline = zeroIfNil(f.Deadline)
	f.Permit.Deadline = zeroIfNil(f.Permit.Deadline)
	for _, v := range []struct {
		n    *big.Int
		bits int
		name string
	}{{f.Amount, 128, "deposit amount"}, {f.SigNonce, 256, "funding nonce"}, {f.Deadline, 256, "funding deadline"}, {f.Permit.Deadline, 256, "permit deadline"}} {
		if err := CheckUint(v.n, v.bits, v.name); err != nil {
			return f, err
		}
	}
	if (f.Amount.Sign() == 0) != (f.Payer == common.Address{}) {
		return f, fmt.Errorf("funding amount and payer must both be zero or positive")
	}
	if f.Amount.Sign() == 0 && (f.SigNonce.Sign() != 0 || f.Deadline.Sign() != 0 || len(f.Signature) != 0 || f.Permit.Deadline.Sign() != 0 || f.Permit.V != 0 || f.Permit.R != [32]byte{} || f.Permit.S != [32]byte{}) {
		return f, fmt.Errorf("zero funding must be entirely empty")
	}
	f.Signature = common.CopyBytes(f.Signature)
	return f, nil
}
func ValidateFundingSubmission(f FundingParams, caller *common.Address, now *big.Int) error {
	f, err := NormalizeFunding(f)
	if err != nil {
		return err
	}
	if len(f.Signature) == 0 {
		if f.SigNonce.Sign() != 0 || f.Deadline.Sign() != 0 || (caller != nil && f.Amount.Sign() > 0 && *caller != f.Payer) {
			return fmt.Errorf("direct funding requires payer caller and zero nonce/deadline")
		}
	} else if now != nil && now.Cmp(f.Deadline) > 0 {
		return fmt.Errorf("funding authorization has expired")
	}
	return nil
}
func ValidateCiphertext(v []byte, length int, version int) error {
	if len(v) != length {
		return fmt.Errorf("ciphertext must be %d bytes", length)
	}
	if version >= 0 && (len(v) == 0 || int(v[0]) != version) {
		return fmt.Errorf("unsupported ciphertext envelope version")
	}
	return nil
}
func mustType(name string) abi.Type {
	v, err := abi.NewType(name, "", nil)
	if err != nil {
		panic(err)
	}
	return v
}
func args(names ...string) abi.Arguments {
	out := make(abi.Arguments, len(names))
	for i, n := range names {
		out[i] = abi.Argument{Type: mustType(n)}
	}
	return out
}

// ValidateDelivery validates the current envelope shape and fixed 24-bit detection clue.
// Empty output slots must have both fields empty.
func ValidateDelivery(delivery Delivery, credit, empty bool) error {
	if empty {
		if len(delivery.Ciphertext) != 0 || len(delivery.Clue) != 0 {
			return fmt.Errorf("empty delivery must have no ciphertext or clue")
		}
		return nil
	}
	length, version := 160, 2
	if credit {
		length, version = 402, 4
	}
	if len(delivery.Ciphertext) < length || len(delivery.Ciphertext) > length+256 {
		return fmt.Errorf("ciphertext must be %d..%d bytes", length, length+256)
	}
	if int(delivery.Ciphertext[0]) != version {
		return fmt.Errorf("unsupported ciphertext envelope version")
	}
	if len(delivery.Clue) != 68 || delivery.Clue[64] != 24 {
		return fmt.Errorf("delivery clue must be 68 bytes with precision 24")
	}
	return nil
}

// DeliveriesHash commits to the ordered complete ABI tuple array, including clues.
func DeliveriesHash(deliveries []Delivery) (common.Hash, error) {
	tuple, err := abi.NewType("tuple[]", "", []abi.ArgumentMarshaling{{Name: "ciphertext", Type: "bytes"}, {Name: "clue", Type: "bytes"}})
	if err != nil {
		return common.Hash{}, err
	}
	return HashABI(abi.Arguments{{Type: tuple}}, deliveries)
}
func EncodeCreditProof(proof []byte) ([]byte, error) {
	if len(proof) == 0 {
		return nil, fmt.Errorf("empty Credit proof")
	}
	return args("uint8", "bytes").Pack(uint8(0), proof)
}
func Word(v *big.Int) [32]byte { return common.BigToHash(v) }
func words(values []*big.Int) [][32]byte {
	out := make([][32]byte, len(values))
	for i, v := range values {
		out[i] = Word(v)
	}
	return out
}
func number(v [32]byte) *big.Int           { return new(big.Int).SetBytes(v[:]) }
func addrNumber(a common.Address) *big.Int { return a.Big() }
func splitHash(h common.Hash) (*big.Int, *big.Int) {
	return new(big.Int).SetBytes(h[:16]), new(big.Int).SetBytes(h[16:])
}
func zeroWords(n int) []*big.Int {
	out := make([]*big.Int, n)
	for i := range out {
		out[i] = new(big.Int)
	}
	return out
}
func padded(v []*big.Int, n int) []*big.Int {
	out := zeroWords(n)
	for i, x := range v {
		out[i] = new(big.Int).Set(x)
	}
	return out
}
func hashWords(v []*big.Int) common.Hash {
	data := make([]byte, 0, 32*len(v))
	for _, x := range v {
		w := Word(x)
		data = append(data, w[:]...)
	}
	return crypto.Keccak256Hash(data)
}
func HashABI(arguments abi.Arguments, values ...any) (common.Hash, error) {
	for _, v := range values {
		if err := validateABIValue(reflect.ValueOf(v)); err != nil {
			return common.Hash{}, err
		}
	}
	v, err := arguments.Pack(values...)
	if err != nil {
		return common.Hash{}, err
	}
	return crypto.Keccak256Hash(v), nil
}
func ProtocolABI() (*abi.ABI, error) { return contracts.InvisiProtocolMetaData.GetAbi() }
func PackProtocolCalldata(name string, values ...any) ([]byte, error) {
	a, err := ProtocolABI()
	if err != nil {
		return nil, err
	}
	for _, v := range values {
		if err := validateABIValue(reflect.ValueOf(v)); err != nil {
			return nil, err
		}
	}
	return a.Pack(name, values...)
}
func methodTuple(method string, index int) (abi.Type, error) {
	a, err := ProtocolABI()
	if err != nil {
		return abi.Type{}, err
	}
	m, ok := a.Methods[method]
	if !ok {
		return abi.Type{}, fmt.Errorf("unknown current ABI method")
	}
	return m.Inputs[index].Type, nil
}
func packCalldata(a abi.ABI, method string, values ...any) ([]byte, error) {
	return a.Pack(method, values...)
}

var erc20ABI = func() abi.ABI {
	t := mustType("address")
	u := mustType("uint256")
	return abi.ABI{Methods: map[string]abi.Method{"approve": abi.NewMethod("approve", "approve", abi.Function, "nonpayable", false, false, abi.Arguments{{Name: "spender", Type: t}, {Name: "amount", Type: u}}, nil)}}
}()

// A nil integer in a tuple makes go-ethereum's reflection encoder panic.
// Validate before entering that encoder so public builders report ordinary errors.
func validateABIValue(v reflect.Value) error {
	if !v.IsValid() {
		return fmt.Errorf("nil ABI value")
	}
	switch v.Kind() {
	case reflect.Pointer, reflect.Interface:
		if v.IsNil() {
			return fmt.Errorf("nil ABI value")
		}
		return validateABIValue(v.Elem())
	case reflect.Struct:
		if v.Type() == reflect.TypeOf(big.Int{}) {
			return nil
		}
		for i := 0; i < v.NumField(); i++ {
			if v.Type().Field(i).IsExported() {
				if err := validateABIValue(v.Field(i)); err != nil {
					return err
				}
			}
		}
	case reflect.Slice, reflect.Array:
		if v.Type().Elem().Kind() == reflect.Uint8 {
			return nil
		}
		for i := 0; i < v.Len(); i++ {
			if err := validateABIValue(v.Index(i)); err != nil {
				return err
			}
		}
	}
	return nil
}
