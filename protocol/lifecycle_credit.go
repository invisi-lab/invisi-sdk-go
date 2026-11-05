package protocol

import (
	"fmt"
	"math/big"
	"math/bits"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/invisi-lab/invisi-sdk-go/core"
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
			return CreditDrawCircuit{Name: fmt.Sprintf("credit_draw_slots%d", n), MaxSize: n, PublicInputCount: 6*n + 14, TotalBits: 126 + bits.Len(uint(n-1))}, nil
		}
	}
	return CreditDrawCircuit{}, fmt.Errorf("unsupported Credit draw shape")
}

type NormalizedCreditDraw struct {
	Params  CreditDrawParams
	Profile CreditDrawCircuit
}

// ValidateCreditDrawRequest validates public request structure; it cannot establish
// ownership or conservation without the private witness and proof verification.
func ValidateCreditDrawRequest(c TransactionContext, p CreditDrawParams, cs []Delivery) (NormalizedCreditDraw, error) {
	r := NormalizedCreditDraw{Params: p}
	if err := ValidateContext(c); err != nil {
		return r, err
	}
	profile, err := SelectCreditDrawCircuit(len(p.Nullifiers))
	if err != nil {
		return r, err
	}
	r.Profile = profile
	if err := CheckUint(c.Now, 64, "chain timestamp"); err != nil {
		return r, err
	}
	if err := ValidateNotes(p.Nullifiers); err != nil {
		return r, err
	}
	r.Params.Sources, err = NormalizeSources(p.Sources, len(p.Nullifiers), profile.MaxSize)
	if err != nil {
		return r, err
	}
	n := len(p.Nullifiers)
	if len(p.RemainderCommitments) != n || len(p.Expiries) != n || len(p.RecoveryAmounts) != n {
		return r, fmt.Errorf("Credit draw array count mismatch")
	}
	if len(cs) != 2*n {
		return r, fmt.Errorf("deliveries count mismatch")
	}
	seen := make(map[[32]byte]bool, n)
	for i, remainder := range p.RemainderCommitments {
		if err := Field(remainder); err != nil {
			return r, err
		}
		if err := Field(p.RecoveryAmounts[i]); err != nil {
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
		if err := CheckUint(p.Expiries[i], 64, "expiry"); err != nil {
			return r, err
		}
		if p.Expiries[i].Sign() == 0 || (c.Now != nil && p.Expiries[i].Cmp(c.Now) <= 0) {
			return r, fmt.Errorf("expiry must be future")
		}
		for j := 0; j < 2; j++ {
			if err := ValidateDelivery(cs[2*i+j], true, remainder == ([32]byte{})); err != nil {
				return r, err
			}
		}

	}
	if err := CheckUint(p.TotalAmount, profile.TotalBits, "totalAmount"); err != nil {
		return r, err
	}
	if p.Spender == (common.Address{}) || p.TotalAmount.Sign() == 0 {
		return r, fmt.Errorf("spender and total amount must be positive")
	}
	if err := CheckDeadline(c, p.Deadline); err != nil {
		return r, err
	}
	return r, nil
}

func CreditDrawBaseInputs(c TransactionContext, p CreditDrawParams, cs []Delivery) ([][32]byte, error) {
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
	h, err := DeliveriesHash(cs)
	if err != nil {
		return nil, err
	}
	v := []*big.Int{c.ChainID, c.ProtocolAddress.Big(), c.TokenAddress.Big(), p.Spender.Big(), big.NewInt(int64(len(p.Nullifiers))), big.NewInt(int64(len(ids)))}
	v = append(v, creditPadded(ids, n)...)
	v = append(v, creditPadded(roots, n)...)
	v = append(v, creditPadded(creditNumbers(p.Nullifiers), n)...)
	v = append(v, creditPadded(creditNumbers(p.RemainderCommitments), n)...)
	v = append(v, creditPadded(p.Expiries, n)...)
	v = append(v, p.TotalAmount, p.Deadline, new(big.Int).SetBytes(h[:16]), new(big.Int).SetBytes(h[16:]))
	v = append(v, creditPadded(creditNumbers(p.RecoveryAmounts), n)...)
	periods := core.CreditPeriods(c.Now.Uint64())
	v = append(v, new(big.Int).SetUint64(periods.CurrentDay), new(big.Int).SetUint64(periods.CurrentMonth))
	return creditWords(v), nil
}

func CreditDrawPublicInputs(c TransactionContext, p CreditDrawParams, cs []Delivery) ([][32]byte, error) {
	base, err := CreditDrawBaseInputs(c, p, cs)
	if err != nil {
		return nil, err
	}
	profile, err := SelectCreditDrawCircuit(len(p.Nullifiers))
	if err != nil {
		return nil, err
	}
	return AppendWalletAuthorization(c, core.WalletCreditDraw, profile.MaxSize, base)
}
func EncodeCreditDrawTypedData(c TransactionContext, p CreditDrawParams, cs []Delivery) (apitypes.TypedData, error) {
	base, err := CreditDrawBaseInputs(c, p, cs)
	if err != nil {
		return apitypes.TypedData{}, err
	}
	profile, err := SelectCreditDrawCircuit(len(p.Nullifiers))
	if err != nil {
		return apitypes.TypedData{}, err
	}
	return EncodeWalletAuthorizationTypedData(c, core.WalletCreditDraw, profile.MaxSize, base)
}

type NormalizedCancelCredit struct{ Params CancelCreditParams }

// Cancellation has no Credit expiry condition: the owner may cancel anytime.
func ValidateCancelCreditRequest(c TransactionContext, p CancelCreditParams) (NormalizedCancelCredit, error) {
	r := NormalizedCancelCredit{Params: p}
	var feeErr error
	r.Params.FeeAmount, feeErr = normalizeFee(p.FeeRecipient, p.FeeAmount)
	if feeErr != nil {
		return r, feeErr
	}
	if p.Amount != nil && r.Params.FeeAmount.Cmp(p.Amount) >= 0 {
		return r, fmt.Errorf("fee must leave positive owner refund")
	}
	if err := ValidateContext(c); err != nil {
		return r, err
	}
	sources, err := NormalizeSources([]InputSource{p.Source}, 1, 1)
	if err != nil {
		return r, err
	}
	r.Params.Source = sources[0]
	if err := ValidateNotes([][32]byte{p.Nullifier}); err != nil {
		return r, err
	}
	if err := CheckUint(p.Amount, 126, "amount"); err != nil {
		return r, err
	}
	if p.Owner == (common.Address{}) || p.Amount.Sign() == 0 {
		return r, fmt.Errorf("owner and amount must be positive")
	}
	if err := CheckDeadline(c, p.Deadline); err != nil {
		return r, err
	}
	return r, nil
}
func CancelCreditBaseInputs(c TransactionContext, p CancelCreditParams) ([][32]byte, error) {
	r, err := ValidateCancelCreditRequest(c, p)
	if err != nil {
		return nil, err
	}
	p = r.Params
	return creditWords([]*big.Int{c.ChainID, c.ProtocolAddress.Big(), c.TokenAddress.Big(), p.Owner.Big(), p.Source.TreeId, creditNumber(p.Source.MerkleRoot), creditNumber(p.Nullifier), p.Amount, p.Deadline, p.FeeRecipient.Big(), p.FeeAmount}), nil
}

func CancelCreditPublicInputs(c TransactionContext, p CancelCreditParams) ([][32]byte, error) {
	base, err := CancelCreditBaseInputs(c, p)
	if err != nil {
		return nil, err
	}
	return AppendWalletAuthorization(c, core.WalletCancelCredit, 1, base)
}

type NormalizedTopUpCredit struct {
	Params  TopUpCreditParams
	Funding FundingParams
}

func ValidateTopUpCreditRequest(c TransactionContext, p TopUpCreditParams, cs []Delivery, f FundingParams) (NormalizedTopUpCredit, error) {
	r := NormalizedTopUpCredit{Params: p}
	creation, err := ValidateCreateCreditRequest(c, p.Creation, cs, f)
	if err != nil {
		return r, err
	}
	r.Params.Creation = creation.Params
	r.Funding = creation.Funding
	sources, err := NormalizeSources([]InputSource{p.CreditSource}, 1, 1)
	if err != nil {
		return r, err
	}
	r.Params.CreditSource = sources[0]
	if err := ValidateNotes([][32]byte{p.CreditNullifier}); err != nil {
		return r, err
	}
	if err := CheckUint(p.PreviousExpiry, 64, "previousExpiry"); err != nil {
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
func TopUpCreditBaseInputs(c TransactionContext, p TopUpCreditParams, cs []Delivery, f FundingParams) ([][32]byte, error) {
	r, err := ValidateTopUpCreditRequest(c, p, cs, f)
	if err != nil {
		return nil, err
	}
	v, err := CreateCreditBaseInputs(c, r.Params.Creation, cs, r.Funding)
	if err != nil {
		return nil, err
	}
	return append(v, Word(r.Params.CreditSource.TreeId), r.Params.CreditSource.MerkleRoot, p.CreditNullifier, Word(p.PreviousExpiry)), nil
}
func TopUpCreditPublicInputs(c TransactionContext, p TopUpCreditParams, cs []Delivery, f FundingParams) ([][32]byte, error) {
	base, err := TopUpCreditBaseInputs(c, p, cs, f)
	if err != nil {
		return nil, err
	}
	return AppendWalletAuthorization(c, core.WalletTopUpCredit, 10, base)
}
func CancelCreditRequestHash(p CancelCreditParams) (common.Hash, error) {
	r, err := ValidateCancelCreditRequest(TransactionContext{ChainID: new(big.Int)}, p)
	if err != nil {
		return common.Hash{}, err
	}
	tuple, err := creditMethodTuple("cancelCredit", 1)
	if err != nil {
		return common.Hash{}, err
	}
	return HashABI(abi.Arguments{{Type: creditABIType("bytes32")}, {Type: tuple}}, crypto.Keccak256Hash([]byte("InvisiCancelCredit")), r.Params)
}
func TopUpCreditRequestHash(p TopUpCreditParams, cs []Delivery, f FundingParams) (common.Hash, error) {
	r, err := ValidateTopUpCreditRequest(TransactionContext{ChainID: new(big.Int)}, p, cs, f)
	if err != nil {
		return common.Hash{}, err
	}
	tuple, err := creditMethodTuple("topUpCredit", 1)
	if err != nil {
		return common.Hash{}, err
	}
	deliveryHash, err := DeliveriesHash(cs)
	if err != nil {
		return common.Hash{}, err
	}
	return HashABI(abi.Arguments{{Type: creditABIType("bytes32")}, {Type: tuple}, {Type: creditABIType("bytes32")}, {Type: creditABIType("address")}, {Type: creditABIType("uint256")}}, crypto.Keccak256Hash([]byte("InvisiTopUpCredit")), r.Params, deliveryHash, r.Funding.Payer, r.Funding.Amount)
}
func EncodeCancelCreditTypedData(c TransactionContext, p CancelCreditParams) (apitypes.TypedData, error) {
	base, err := CancelCreditBaseInputs(c, p)
	if err != nil {
		return apitypes.TypedData{}, err
	}
	return EncodeWalletAuthorizationTypedData(c, core.WalletCancelCredit, 1, base)
}
func EncodeTopUpCreditTypedData(c TransactionContext, p TopUpCreditParams, cs []Delivery, f FundingParams) (apitypes.TypedData, error) {
	base, err := TopUpCreditBaseInputs(c, p, cs, f)
	if err != nil {
		return apitypes.TypedData{}, err
	}
	return EncodeWalletAuthorizationTypedData(c, core.WalletTopUpCredit, 10, base)
}
func EncodeTopUpCreditFundingTypedData(c TransactionContext, p TopUpCreditParams, cs []Delivery, f FundingParams) (apitypes.TypedData, error) {
	r, err := ValidateTopUpCreditRequest(c, p, cs, f)
	if err != nil {
		return apitypes.TypedData{}, err
	}
	h, err := TopUpCreditRequestHash(p, cs, f)
	if err != nil {
		return apitypes.TypedData{}, err
	}
	return EncodeFundingTypedData(c, "FundTransact", h, r.Funding)
}
func BuildCreditDrawCalldata(p BuildCreditDrawCalldataParams) (*Calldata, error) {
	r, err := ValidateCreditDrawRequest(p.Context, p.Params, p.Deliveries)
	if err != nil {
		return nil, err
	}
	proof, err := EncodeCreditProof(p.Proof)
	if err != nil {
		return nil, err
	}
	data, err := PackProtocolCalldata("creditDraw", proof, r.Params, p.Deliveries)
	if err != nil {
		return nil, err
	}
	return &Calldata{Data: data, ApproveAmount: new(big.Int), Capacity: r.Profile.MaxSize}, nil
}
func BuildCancelCreditCalldata(p BuildCancelCreditCalldataParams) (*Calldata, error) {
	r, err := ValidateCancelCreditRequest(p.Context, p.Params)
	if err != nil {
		return nil, err
	}
	proof, err := EncodeCreditProof(p.Proof)
	if err != nil {
		return nil, err
	}
	data, err := PackProtocolCalldata("cancelCredit", proof, r.Params)
	if err != nil {
		return nil, err
	}
	return &Calldata{Data: data, ApproveAmount: new(big.Int)}, nil
}
func BuildTopUpCreditCalldata(p BuildTopUpCreditCalldataParams) (*Calldata, error) {
	r, err := ValidateTopUpCreditRequest(p.Context, p.Params, p.Deliveries, p.Funding)
	if err != nil {
		return nil, err
	}
	if err := ValidateFundingSubmission(r.Funding, p.Caller, p.Context.Now); err != nil {
		return nil, err
	}
	proof, err := EncodeCreditProof(p.Proof)
	if err != nil {
		return nil, err
	}
	data, err := PackProtocolCalldata("topUpCredit", proof, r.Params, p.Deliveries, r.Funding)
	if err != nil {
		return nil, err
	}
	return &Calldata{Data: data, ApproveAmount: new(big.Int).Set(r.Funding.Amount)}, nil
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
		out[i] = Word(n)
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
	a, err := ProtocolABI()
	if err != nil {
		return abi.Type{}, err
	}
	m, ok := a.Methods[name]
	if !ok || index >= len(m.Inputs) {
		return abi.Type{}, fmt.Errorf("current ABI missing %s tuple", name)
	}
	return m.Inputs[index].Type, nil
}
