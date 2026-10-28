package payer

import (
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"math/big"
)

type NormalizedCreditCreation struct {
	Params  CreateCreditParams
	Funding FundingParams
}

func ValidateCreateCreditRequest(c TransactionContext, p CreateCreditParams, cs [][]byte, f FundingParams) (NormalizedCreditCreation, error) {
	r := NormalizedCreditCreation{Params: p}
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
	want := 2
	if nonzero(p.ChangeCommitment) {
		want++
	}
	if len(cs) != want {
		return r, fmt.Errorf("ciphertexts count mismatch")
	}
	for i, v := range cs {
		length, version := 254, 2
		if i == 2 {
			length, version = 178, 1
		}
		if err := ValidateCiphertext(v, length, version); err != nil {
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
func CreateCreditPublicInputs(c TransactionContext, p CreateCreditParams, cs [][]byte, f FundingParams) ([][32]byte, error) {
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
	high, low := splitHash(CiphertextsHash(cs))
	v := []*big.Int{c.ChainID, c.ProtocolAddress.Big(), c.TokenAddress.Big(), p.Owner.Big(), big.NewInt(int64(len(ns))), big.NewInt(int64(len(ids)))}
	v = append(v, padded(ids, 10)...)
	v = append(v, padded(roots, 10)...)
	v = append(v, padded(ns, 10)...)
	v = append(v, number(p.CreditCommitment), number(p.ChangeCommitment), high, low, p.Expiry, p.Deadline, r.Funding.Amount, r.Funding.Payer.Big())
	return words(v), nil
}
func CreateCreditRequestHash(p CreateCreditParams, cs [][]byte, f FundingParams) (common.Hash, error) {
	r, err := ValidateCreateCreditRequest(TransactionContext{ChainID: new(big.Int)}, p, cs, f)
	if err != nil {
		return common.Hash{}, err
	}
	tuple, err := methodTuple("createCredit", 1)
	if err != nil {
		return common.Hash{}, err
	}
	return HashABI(abi.Arguments{{Type: mustType("bytes32")}, {Type: tuple}, {Type: mustType("bytes32")}, {Type: mustType("address")}, {Type: mustType("uint256")}},
		crypto.Keccak256Hash([]byte("InvisiCreateCredit")), r.Params, CiphertextsHash(cs), r.Funding.Payer, r.Funding.Amount)
}
func EncodeCreateCreditTypedData(c TransactionContext, p CreateCreditParams, cs [][]byte, a CreditAuthorization, f FundingParams) (apitypes.TypedData, error) {
	if _, err := ValidateCreateCreditRequest(c, p, cs, f); err != nil {
		return apitypes.TypedData{}, err
	}
	hash, err := CreateCreditRequestHash(p, cs, f)
	if err != nil {
		return apitypes.TypedData{}, err
	}
	return EncodeOwnerTypedData(c, "CreateCredit", hash, p.Owner, a.SigNonce, p.Deadline)
}
func EncodeCreateCreditFundingTypedData(c TransactionContext, p CreateCreditParams, cs [][]byte, f FundingParams) (apitypes.TypedData, error) {
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
	r, err := ValidateCreateCreditRequest(p.Context, p.Params, p.Ciphertexts, p.Funding)
	if err != nil {
		return nil, err
	}
	if err := ValidateFundingSubmission(r.Funding, p.Caller, p.Context.Now); err != nil {
		return nil, err
	}
	if err := CheckUint(p.Authorization.SigNonce, 256, "authorization nonce"); err != nil {
		return nil, err
	}
	if len(p.Authorization.Signature) == 0 {
		return nil, fmt.Errorf("Credit owner authorization required")
	}
	proof, err := EncodeCreditProof(p.Proof)
	if err != nil {
		return nil, err
	}
	data, err := PackProtocolCalldata("createCredit", proof, r.Params, p.Ciphertexts, r.Funding, p.Authorization)
	if err != nil {
		return nil, err
	}
	return &Calldata{Data: data, ApproveAmount: new(big.Int).Set(r.Funding.Amount)}, nil
}

type CreditNonceGuideParams struct {
	Owner, Payer, Caller                  common.Address
	OwnerNonce, PayerNonce, FundingAmount *big.Int
}
type CreditNonceGuide struct {
	AuthorizationNonce, FundingNonce, OwnerNextNonce, PayerNextNonce *big.Int
	FundingSignatureRequired                                         bool
}

func CreateCreditNonceGuide(p CreditNonceGuideParams) (CreditNonceGuide, error) {
	var result CreditNonceGuide
	if p.Owner == (common.Address{}) {
		return result, fmt.Errorf("owner must be nonzero")
	}
	if err := CheckUint(p.OwnerNonce, 256, "owner nonce"); err != nil {
		return result, err
	}
	if p.PayerNonce != nil {
		if err := CheckUint(p.PayerNonce, 256, "payer nonce"); err != nil {
			return result, err
		}
	}
	f, err := NormalizeFunding(FundingParams{Payer: p.Payer, Amount: p.FundingAmount})
	if err != nil {
		return result, err
	}
	signed, same := f.Amount.Sign() > 0 && p.Payer != p.Caller, p.Payer == p.Owner
	if same && p.PayerNonce != nil && p.PayerNonce.Cmp(p.OwnerNonce) != 0 {
		return result, fmt.Errorf("same owner and payer must have same current nonce")
	}
	result.AuthorizationNonce = new(big.Int).Set(p.OwnerNonce)
	result.FundingNonce = new(big.Int)
	result.FundingSignatureRequired = signed
	result.OwnerNextNonce = new(big.Int).Add(p.OwnerNonce, big.NewInt(1))
	if signed {
		if same {
			result.FundingNonce.Set(result.OwnerNextNonce)
			result.OwnerNextNonce.Add(result.OwnerNextNonce, big.NewInt(1))
		} else {
			if err := CheckUint(p.PayerNonce, 256, "payer nonce"); err != nil {
				return result, err
			}
			result.FundingNonce.Set(p.PayerNonce)
		}
	}
	if same {
		result.PayerNextNonce = new(big.Int).Set(result.OwnerNextNonce)
	} else if signed {
		result.PayerNextNonce = new(big.Int).Add(result.FundingNonce, big.NewInt(1))
	} else {
		result.PayerNextNonce = zeroIfNil(p.PayerNonce)
	}
	if err := CheckUint(result.OwnerNextNonce, 256, "next owner nonce"); err != nil {
		return result, err
	}
	if err := CheckUint(result.PayerNextNonce, 256, "next payer nonce"); err != nil {
		return result, err
	}
	return result, nil
}
