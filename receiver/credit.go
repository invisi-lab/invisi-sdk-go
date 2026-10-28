package receiver

import (
	"fmt"
	"math/big"
	"math/bits"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/invisi-lab/invisi-sdk-go/payer"
)

// CreditDrawCircuit describes the smallest current circuit that fits a draw.
type CreditDrawCircuit struct {
	Name                                 string
	MaxSize, PublicInputCount, TotalBits int
}

func SelectCreditDrawCircuit(count int) (CreditDrawCircuit, error) {
	if count < 1 || count > 120 {
		return CreditDrawCircuit{}, fmt.Errorf("draw requires 1..120 credits")
	}
	for _, n := range []int{1, 12, 50, 120} {
		if count <= n {
			return CreditDrawCircuit{Name: fmt.Sprintf("credit_draw_slots%d", n), MaxSize: n, PublicInputCount: 6*n + 10, TotalBits: 126 + bits.Len(uint(n-1))}, nil
		}
	}
	return CreditDrawCircuit{}, fmt.Errorf("unsupported Credit draw shape")
}

type NormalizedCreditDraw struct {
	Params  payer.CreditDrawParams
	Profile CreditDrawCircuit
}

// ValidateCreditDrawRequest validates public request structure; it cannot establish
// ownership or conservation without the private witness and proof verification.
func ValidateCreditDrawRequest(c payer.TransactionContext, p payer.CreditDrawParams, cs [][]byte) (NormalizedCreditDraw, error) {
	r := NormalizedCreditDraw{Params: p}
	if err := payer.ValidateContext(c); err != nil {
		return r, err
	}
	profile, err := SelectCreditDrawCircuit(len(p.Nullifiers))
	if err != nil {
		return r, err
	}
	r.Profile = profile
	if err := payer.ValidateNotes(p.Nullifiers); err != nil {
		return r, err
	}
	r.Params.Sources, err = payer.NormalizeSources(p.Sources, len(p.Nullifiers), profile.MaxSize)
	if err != nil {
		return r, err
	}
	n := len(p.Nullifiers)
	if len(p.RemainderCommitments) != n || len(p.Expiries) != n || len(p.RecoveryAmounts) != n {
		return r, fmt.Errorf("Credit draw array count mismatch")
	}
	if len(cs) != 2*n {
		return r, fmt.Errorf("ciphertexts count mismatch")
	}
	seen := make(map[[32]byte]bool, n)
	for i, remainder := range p.RemainderCommitments {
		if err := payer.Field(remainder); err != nil {
			return r, err
		}
		if err := payer.Field(p.RecoveryAmounts[i]); err != nil {
			return r, err
		}
		if remainder != ([32]byte{}) {
			if seen[remainder] {
				return r, fmt.Errorf("duplicate remainder commitment")
			}
			seen[remainder] = true
		} else if p.RecoveryAmounts[i] != ([32]byte{}) {
			return r, fmt.Errorf("full draw requires zero recovery amount")
		}
		if err := payer.CheckUint(p.Expiries[i], 64, "expiry"); err != nil {
			return r, err
		}
		if p.Expiries[i].Sign() == 0 || (c.Now != nil && p.Expiries[i].Cmp(c.Now) <= 0) {
			return r, fmt.Errorf("expiry must be future")
		}
		length, version := 254, 2
		if remainder == ([32]byte{}) {
			length, version = 0, -1
		}
		for j := 0; j < 2; j++ {
			if err := payer.ValidateCiphertext(cs[2*i+j], length, version); err != nil {
				return r, err
			}
		}
	}
	if err := payer.CheckUint(p.TotalAmount, profile.TotalBits, "totalAmount"); err != nil {
		return r, err
	}
	if p.Spender == (common.Address{}) || p.TotalAmount.Sign() == 0 {
		return r, fmt.Errorf("spender and total amount must be positive")
	}
	if err := payer.CheckDeadline(c, p.Deadline); err != nil {
		return r, err
	}
	return r, nil
}

func CreditDrawPublicInputs(c payer.TransactionContext, p payer.CreditDrawParams, cs [][]byte) ([][32]byte, error) {
	r, err := ValidateCreditDrawRequest(c, p, cs)
	if err != nil {
		return nil, err
	}
	n := r.Profile.MaxSize
	ids, roots := make([]*big.Int, len(r.Params.Sources)), make([]*big.Int, len(r.Params.Sources))
	for i, s := range r.Params.Sources {
		ids[i] = s.TreeId
		roots[i] = creditNumber(s.MerkleRoot)
	}
	h := payer.CiphertextsHash(cs)
	v := []*big.Int{c.ChainID, c.ProtocolAddress.Big(), c.TokenAddress.Big(), p.Spender.Big(), big.NewInt(int64(len(p.Nullifiers))), big.NewInt(int64(len(ids)))}
	v = append(v, creditPadded(ids, n)...)
	v = append(v, creditPadded(roots, n)...)
	v = append(v, creditPadded(creditNumbers(p.Nullifiers), n)...)
	v = append(v, creditPadded(creditNumbers(p.RemainderCommitments), n)...)
	v = append(v, creditPadded(p.Expiries, n)...)
	v = append(v, p.TotalAmount, p.Deadline, new(big.Int).SetBytes(h[:16]), new(big.Int).SetBytes(h[16:]))
	v = append(v, creditPadded(creditNumbers(p.RecoveryAmounts), n)...)
	return creditWords(v), nil
}

type NormalizedCancelCredit struct{ Params payer.CancelCreditParams }

// Cancellation has no Credit expiry condition: the owner may cancel anytime.
func ValidateCancelCreditRequest(c payer.TransactionContext, p payer.CancelCreditParams) (NormalizedCancelCredit, error) {
	r := NormalizedCancelCredit{Params: p}
	if err := payer.ValidateContext(c); err != nil {
		return r, err
	}
	sources, err := payer.NormalizeSources([]payer.InputSource{p.Source}, 1, 1)
	if err != nil {
		return r, err
	}
	r.Params.Source = sources[0]
	if err := payer.ValidateNotes([][32]byte{p.Nullifier}); err != nil {
		return r, err
	}
	if err := payer.CheckUint(p.Amount, 126, "amount"); err != nil {
		return r, err
	}
	if p.Owner == (common.Address{}) || p.Amount.Sign() == 0 {
		return r, fmt.Errorf("owner and amount must be positive")
	}
	if err := payer.CheckDeadline(c, p.Deadline); err != nil {
		return r, err
	}
	return r, nil
}
func CancelCreditPublicInputs(c payer.TransactionContext, p payer.CancelCreditParams) ([][32]byte, error) {
	r, err := ValidateCancelCreditRequest(c, p)
	if err != nil {
		return nil, err
	}
	p = r.Params
	return creditWords([]*big.Int{c.ChainID, c.ProtocolAddress.Big(), c.TokenAddress.Big(), p.Owner.Big(), p.Source.TreeId, creditNumber(p.Source.MerkleRoot), creditNumber(p.Nullifier), p.Amount, p.Deadline}), nil
}

type NormalizedTopUpCredit struct {
	Params  payer.TopUpCreditParams
	Funding payer.FundingParams
}

func ValidateTopUpCreditRequest(c payer.TransactionContext, p payer.TopUpCreditParams, cs [][]byte, f payer.FundingParams) (NormalizedTopUpCredit, error) {
	r := NormalizedTopUpCredit{Params: p}
	creation, err := payer.ValidateCreateCreditRequest(c, p.Creation, cs, f)
	if err != nil {
		return r, err
	}
	r.Params.Creation = creation.Params
	r.Funding = creation.Funding
	sources, err := payer.NormalizeSources([]payer.InputSource{p.CreditSource}, 1, 1)
	if err != nil {
		return r, err
	}
	r.Params.CreditSource = sources[0]
	if err := payer.ValidateNotes([][32]byte{p.CreditNullifier}); err != nil {
		return r, err
	}
	if err := payer.CheckUint(p.PreviousExpiry, 64, "previousExpiry"); err != nil {
		return r, err
	}
	if p.PreviousExpiry.Sign() == 0 || (c.Now != nil && p.PreviousExpiry.Cmp(c.Now) <= 0) {
		return r, fmt.Errorf("previous Credit has expired")
	}
	if p.Creation.Expiry.Cmp(p.PreviousExpiry) < 0 {
		return r, fmt.Errorf("expiry must not decrease")
	}
	for _, v := range p.Creation.Nullifiers {
		if v == p.CreditNullifier {
			return r, fmt.Errorf("duplicate credit and ordinary nullifier")
		}
	}
	return r, nil
}
func TopUpCreditPublicInputs(c payer.TransactionContext, p payer.TopUpCreditParams, cs [][]byte, f payer.FundingParams) ([][32]byte, error) {
	r, err := ValidateTopUpCreditRequest(c, p, cs, f)
	if err != nil {
		return nil, err
	}
	v, err := payer.CreateCreditPublicInputs(c, r.Params.Creation, cs, r.Funding)
	if err != nil {
		return nil, err
	}
	return append(v, payer.Word(r.Params.CreditSource.TreeId), r.Params.CreditSource.MerkleRoot, p.CreditNullifier, payer.Word(p.PreviousExpiry)), nil
}
func CancelCreditRequestHash(p payer.CancelCreditParams) (common.Hash, error) {
	r, err := ValidateCancelCreditRequest(payer.TransactionContext{ChainID: new(big.Int)}, p)
	if err != nil {
		return common.Hash{}, err
	}
	tuple, err := creditMethodTuple("cancelCredit", 1)
	if err != nil {
		return common.Hash{}, err
	}
	return payer.HashABI(abi.Arguments{{Type: creditABIType("bytes32")}, {Type: tuple}}, crypto.Keccak256Hash([]byte("InvisiCancelCredit")), r.Params)
}
func TopUpCreditRequestHash(p payer.TopUpCreditParams, cs [][]byte, f payer.FundingParams) (common.Hash, error) {
	r, err := ValidateTopUpCreditRequest(payer.TransactionContext{ChainID: new(big.Int)}, p, cs, f)
	if err != nil {
		return common.Hash{}, err
	}
	tuple, err := creditMethodTuple("topUpCredit", 1)
	if err != nil {
		return common.Hash{}, err
	}
	return payer.HashABI(abi.Arguments{{Type: creditABIType("bytes32")}, {Type: tuple}, {Type: creditABIType("bytes32")}, {Type: creditABIType("address")}, {Type: creditABIType("uint256")}}, crypto.Keccak256Hash([]byte("InvisiTopUpCredit")), r.Params, payer.CiphertextsHash(cs), r.Funding.Payer, r.Funding.Amount)
}
func EncodeCancelCreditTypedData(c payer.TransactionContext, p payer.CancelCreditParams, a payer.CreditAuthorization) (apitypes.TypedData, error) {
	if _, err := ValidateCancelCreditRequest(c, p); err != nil {
		return apitypes.TypedData{}, err
	}
	h, err := CancelCreditRequestHash(p)
	if err != nil {
		return apitypes.TypedData{}, err
	}
	return payer.EncodeOwnerTypedData(c, "CancelCredit", h, p.Owner, a.SigNonce, p.Deadline)
}
func EncodeTopUpCreditTypedData(c payer.TransactionContext, p payer.TopUpCreditParams, cs [][]byte, a payer.CreditAuthorization, f payer.FundingParams) (apitypes.TypedData, error) {
	if _, err := ValidateTopUpCreditRequest(c, p, cs, f); err != nil {
		return apitypes.TypedData{}, err
	}
	h, err := TopUpCreditRequestHash(p, cs, f)
	if err != nil {
		return apitypes.TypedData{}, err
	}
	return payer.EncodeOwnerTypedData(c, "TopUpCredit", h, p.Creation.Owner, a.SigNonce, p.Creation.Deadline)
}
func EncodeTopUpCreditFundingTypedData(c payer.TransactionContext, p payer.TopUpCreditParams, cs [][]byte, f payer.FundingParams) (apitypes.TypedData, error) {
	r, err := ValidateTopUpCreditRequest(c, p, cs, f)
	if err != nil {
		return apitypes.TypedData{}, err
	}
	h, err := TopUpCreditRequestHash(p, cs, f)
	if err != nil {
		return apitypes.TypedData{}, err
	}
	return payer.EncodeFundingTypedData(c, "FundTransact", h, r.Funding)
}
func TopUpCreditNonceGuide(p payer.CreditNonceGuideParams) (payer.CreditNonceGuide, error) {
	return payer.CreateCreditNonceGuide(p)
}

type CancelCreditNonceGuideResult struct{ AuthorizationNonce, OwnerNextNonce *big.Int }

func CancelCreditNonceGuide(ownerNonce *big.Int) (CancelCreditNonceGuideResult, error) {
	var r CancelCreditNonceGuideResult
	if err := payer.CheckUint(ownerNonce, 256, "owner nonce"); err != nil {
		return r, err
	}
	next := new(big.Int).Add(ownerNonce, big.NewInt(1))
	if err := payer.CheckUint(next, 256, "next owner nonce"); err != nil {
		return r, err
	}
	return CancelCreditNonceGuideResult{AuthorizationNonce: new(big.Int).Set(ownerNonce), OwnerNextNonce: next}, nil
}
func BuildCreditDrawCalldata(p BuildCreditDrawCalldataParams) (*payer.Calldata, error) {
	r, err := ValidateCreditDrawRequest(p.Context, p.Params, p.Ciphertexts)
	if err != nil {
		return nil, err
	}
	proof, err := payer.EncodeCreditProof(p.Proof)
	if err != nil {
		return nil, err
	}
	data, err := payer.PackProtocolCalldata("creditDraw", proof, r.Params, p.Ciphertexts)
	if err != nil {
		return nil, err
	}
	return &payer.Calldata{Data: data, ApproveAmount: new(big.Int), Capacity: r.Profile.MaxSize}, nil
}
func BuildCancelCreditCalldata(p BuildCancelCreditCalldataParams) (*payer.Calldata, error) {
	r, err := ValidateCancelCreditRequest(p.Context, p.Params)
	if err != nil {
		return nil, err
	}
	if err := validateCreditAuthorization(p.Authorization); err != nil {
		return nil, err
	}
	proof, err := payer.EncodeCreditProof(p.Proof)
	if err != nil {
		return nil, err
	}
	data, err := payer.PackProtocolCalldata("cancelCredit", proof, r.Params, p.Authorization)
	if err != nil {
		return nil, err
	}
	return &payer.Calldata{Data: data, ApproveAmount: new(big.Int)}, nil
}
func BuildTopUpCreditCalldata(p BuildTopUpCreditCalldataParams) (*payer.Calldata, error) {
	r, err := ValidateTopUpCreditRequest(p.Context, p.Params, p.Ciphertexts, p.Funding)
	if err != nil {
		return nil, err
	}
	if err := payer.ValidateFundingSubmission(r.Funding, p.Caller, p.Context.Now); err != nil {
		return nil, err
	}
	if err := validateCreditAuthorization(p.Authorization); err != nil {
		return nil, err
	}
	proof, err := payer.EncodeCreditProof(p.Proof)
	if err != nil {
		return nil, err
	}
	data, err := payer.PackProtocolCalldata("topUpCredit", proof, r.Params, p.Ciphertexts, r.Funding, p.Authorization)
	if err != nil {
		return nil, err
	}
	return &payer.Calldata{Data: data, ApproveAmount: new(big.Int).Set(r.Funding.Amount)}, nil
}
func validateCreditAuthorization(a payer.CreditAuthorization) error {
	if err := payer.CheckUint(a.SigNonce, 256, "authorization nonce"); err != nil {
		return err
	}
	if len(a.Signature) == 0 {
		return fmt.Errorf("Credit owner authorization required")
	}
	return nil
}
func creditNumber(v [32]byte) *big.Int { return new(big.Int).SetBytes(v[:]) }
func creditNumbers(values [][32]byte) []*big.Int {
	v := make([]*big.Int, len(values))
	for i, w := range values {
		v[i] = creditNumber(w)
	}
	return v
}
func creditPadded(v []*big.Int, n int) []*big.Int {
	out := make([]*big.Int, n)
	for i := range out {
		out[i] = new(big.Int)
		if i < len(v) {
			out[i].Set(v[i])
		}
	}
	return out
}
func creditWords(v []*big.Int) [][32]byte {
	out := make([][32]byte, len(v))
	for i, n := range v {
		out[i] = payer.Word(n)
	}
	return out
}
func creditABIType(name string) abi.Type {
	v, err := abi.NewType(name, "", nil)
	if err != nil {
		panic(err)
	}
	return v
}
func creditMethodTuple(name string, index int) (abi.Type, error) {
	a, err := payer.ProtocolABI()
	if err != nil {
		return abi.Type{}, err
	}
	m, ok := a.Methods[name]
	if !ok || index >= len(m.Inputs) {
		return abi.Type{}, fmt.Errorf("current ABI missing %s tuple", name)
	}
	return m.Inputs[index].Type, nil
}
