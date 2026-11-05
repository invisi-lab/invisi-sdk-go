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
