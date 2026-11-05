package protocol

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"

	"github.com/invisi-lab/invisi-sdk-go/core"
)

type NormalizedCreditCreation struct {
	Params  CreateCreditParams
	Funding FundingParams
}

func ValidateCreateCreditRequest(c TransactionContext, p CreateCreditParams, cs []Delivery, f FundingParams) (NormalizedCreditCreation, error) {
	r := NormalizedCreditCreation{Params: p}
	var feeErr error
	r.Params.FeeAmount, feeErr = normalizeFee(p.FeeRecipient, p.FeeAmount)
	if feeErr != nil {
		return r, feeErr
	}
	if err := ValidateContext(c); err != nil {
		return r, err
	}
	if err := ValidateNotes(p.Nullifiers); err != nil {
		return r, err
	}
	sources, err := NormalizeSources(p.Sources, len(p.Nullifiers), 10)
	if err != nil {
		return r, err
	}
	r.Params.Sources = sources
	for _, v := range [][32]byte{p.CreditCommitment, p.ChangeCommitment} {
		if err := Field(v); err != nil {
			return r, err
		}
	}
	if !nonzero(p.CreditCommitment) || p.CreditCommitment == p.ChangeCommitment {
		return r, fmt.Errorf("zero or duplicate commitment")
	}
	if err := CheckUint(p.Expiry, 64, "expiry"); err != nil {
		return r, err
	}
	if p.Owner == (common.Address{}) || p.Expiry.Sign() == 0 || (c.Now != nil && p.Expiry.Cmp(c.Now) <= 0) {
		return r, fmt.Errorf("owner must be nonzero and expiry future")
	}
	if err := CheckDeadline(c, p.Deadline); err != nil {
		return r, err
	}
	if len(cs) != 3 {
		return r, fmt.Errorf("Credit requires three delivery slots")
	}
	for i, v := range cs {
		if err := ValidateDelivery(v, i < 2, i == 2 && !nonzero(p.ChangeCommitment)); err != nil {
			return r, err
		}
	}
	r.Funding, err = NormalizeFunding(f)
	if err != nil {
		return r, err
	}
	if len(p.Nullifiers) == 0 && r.Funding.Amount.Sign() == 0 {
		return r, fmt.Errorf("empty inputs require positive funding")
	}
	return r, nil
}
func CreateCreditBaseInputs(c TransactionContext, p CreateCreditParams, cs []Delivery, f FundingParams) ([][32]byte, error) {
	r, err := ValidateCreateCreditRequest(c, p, cs, f)
	if err != nil {
		return nil, err
	}
	ids, roots := make([]*big.Int, len(p.Sources)), make([]*big.Int, len(p.Sources))
	for i, s := range r.Params.Sources {
		ids[i] = s.TreeId
		roots[i] = number(s.MerkleRoot)
	}
	ns := make([]*big.Int, len(p.Nullifiers))
	for i, v := range p.Nullifiers {
		ns[i] = number(v)
	}
	hash, err := DeliveriesHash(cs)
	if err != nil {
		return nil, err
	}
	high, low := splitHash(hash)
	v := []*big.Int{c.ChainID, c.ProtocolAddress.Big(), c.TokenAddress.Big(), p.Owner.Big(), big.NewInt(int64(len(ns))), big.NewInt(int64(len(ids)))}
	v = append(v, padded(ids, 10)...)
	v = append(v, padded(roots, 10)...)
	v = append(v, padded(ns, 10)...)
	v = append(v, number(p.CreditCommitment), number(p.ChangeCommitment), high, low, p.Expiry, p.Deadline, r.Funding.Amount, r.Funding.Payer.Big(), r.Params.FeeRecipient.Big(), r.Params.FeeAmount)
	return words(v), nil
}
func CreateCreditPublicInputs(c TransactionContext, p CreateCreditParams, cs []Delivery, f FundingParams) ([][32]byte, error) {
	base, err := CreateCreditBaseInputs(c, p, cs, f)
	if err != nil {
		return nil, err
	}
	return AppendWalletAuthorization(c, core.WalletCreateCredit, 10, base)
}
func CreateCreditRequestHash(p CreateCreditParams, cs []Delivery, f FundingParams) (common.Hash, error) {
	r, err := ValidateCreateCreditRequest(TransactionContext{ChainID: new(big.Int)}, p, cs, f)
	if err != nil {
		return common.Hash{}, err
	}
	tuple, err := methodTuple("createCredit", 1)
	if err != nil {
		return common.Hash{}, err
	}
	deliveryHash, err := DeliveriesHash(cs)
	if err != nil {
		return common.Hash{}, err
	}
	return HashABI(abi.Arguments{{Type: mustType("bytes32")}, {Type: tuple}, {Type: mustType("bytes32")}, {Type: mustType("address")}, {Type: mustType("uint256")}},
		crypto.Keccak256Hash([]byte("InvisiCreateCredit")), r.Params, deliveryHash, r.Funding.Payer, r.Funding.Amount)
}
func EncodeCreateCreditTypedData(c TransactionContext, p CreateCreditParams, cs []Delivery, f FundingParams) (apitypes.TypedData, error) {
	base, err := CreateCreditBaseInputs(c, p, cs, f)
	if err != nil {
		return apitypes.TypedData{}, err
	}
	return EncodeWalletAuthorizationTypedData(c, core.WalletCreateCredit, 10, base)
}
func EncodeCreateCreditFundingTypedData(c TransactionContext, p CreateCreditParams, cs []Delivery, f FundingParams) (apitypes.TypedData, error) {
	r, err := ValidateCreateCreditRequest(c, p, cs, f)
	if err != nil {
		return apitypes.TypedData{}, err
	}
	hash, err := CreateCreditRequestHash(p, cs, f)
	if err != nil {
		return apitypes.TypedData{}, err
	}
	return EncodeFundingTypedData(c, "FundTransact", hash, r.Funding)
}
func BuildCreateCreditCalldata(p BuildCreateCreditCalldataParams) (*Calldata, error) {
	r, err := ValidateCreateCreditRequest(p.Context, p.Params, p.Deliveries, p.Funding)
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
	data, err := PackProtocolCalldata("createCredit", proof, r.Params, p.Deliveries, r.Funding)
	if err != nil {
		return nil, err
	}
	return &Calldata{Data: data, ApproveAmount: new(big.Int).Set(r.Funding.Amount)}, nil
}
