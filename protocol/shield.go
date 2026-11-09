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

func ValidateShieldRequest(p BuildShieldCalldataParams) (FundingParams, error) {
	if err := ValidateContext(p.Context); err != nil {
		return FundingParams{}, err
	}
	if _, err := normalizeFee(p.Params.FeeRecipient, p.Params.FeeAmount); err != nil {
		return FundingParams{}, err
	}
	if err := CheckDeadline(p.Context, p.Params.Deadline); err != nil {
		return FundingParams{}, err
	}
	n := len(p.Outputs)
	if n < 1 || n > 10 || len(p.Params.NoteCommitments) != n || len(p.Deliveries) != n {
		return FundingParams{}, fmt.Errorf("shield requires 1..10 outputs and matching deliveries")
	}
	if err := ValidateNotes(p.Params.NoteCommitments); err != nil {
		return FundingParams{}, err
	}
	f, err := NormalizeFunding(p.Funding)
	if err != nil {
		return f, err
	}
	total := new(big.Int)
	for i, o := range p.Outputs {
		if err := CheckUint(o.Amount, 128, "amount"); err != nil {
			return f, err
		}
		if o.Amount.Sign() == 0 {
			return f, fmt.Errorf("amount must be positive")
		}
		if err := Field(o.Identity); err != nil {
			return f, err
		}
		if err := ValidateDelivery(p.Deliveries[i], false, false); err != nil {
			return f, err
		}
		commitment, err := core.ComputeNoteCommitment(core.BigIntToElement(p.Context.ChainID), core.BigIntToElement(p.Context.ProtocolAddress.Big()), core.BigIntToElement(number(o.Identity)), core.BigIntToElement(p.Context.TokenAddress.Big()), core.BigIntToElement(o.Amount))
		if err != nil {
			return f, err
		}
		if commitment.Bytes() != p.Params.NoteCommitments[i] {
			return f, fmt.Errorf("output commitment mismatch")
		}
		total.Add(total, o.Amount)
	}
	if new(big.Int).Add(total, feeAmount(p.Params.FeeAmount)).Cmp(f.Amount) != 0 {
		return f, fmt.Errorf("funding must equal output total plus fee")
	}
	return f, nil
}
func ShieldFundingRequestHash(p ShieldParams, outputs []ShieldOutput, cs []Delivery) (common.Hash, error) {
	var feeErr error
	p.FeeAmount, feeErr = normalizeFee(p.FeeRecipient, p.FeeAmount)
	if feeErr != nil {
		return common.Hash{}, feeErr
	}
	if err := CheckUint(p.Deadline, 64, "deadline"); err != nil {
		return common.Hash{}, err
	}
	for _, output := range outputs {
		if err := CheckUint(output.Amount, 128, "amount"); err != nil {
			return common.Hash{}, err
		}
	}

	a, err := ProtocolABI()
	if err != nil {
		return common.Hash{}, err
	}
	method := a.Methods["shield"]
	outputHash, err := HashABI(abi.Arguments{method.Inputs[1]}, outputs)
	if err != nil {
		return common.Hash{}, err
	}
	deliveryHash, err := DeliveriesHash(cs)
	if err != nil {
		return common.Hash{}, err
	}
	return HashABI(abi.Arguments{{Type: mustType("bytes32")}, method.Inputs[0], {Type: mustType("bytes32")}, {Type: mustType("bytes32")}},
		crypto.Keccak256Hash([]byte("InvisiShield")), p, outputHash, deliveryHash)
}
func EncodeShieldFundingTypedData(p BuildShieldCalldataParams) (apitypes.TypedData, error) {
	f, err := ValidateShieldRequest(p)
	if err != nil {
		return apitypes.TypedData{}, err
	}
	hash, err := ShieldFundingRequestHash(p.Params, p.Outputs, p.Deliveries)
	if err != nil {
		return apitypes.TypedData{}, err
	}
	return EncodeFundingTypedData(p.Context, "FundShield", hash, f)
}
func BuildShieldCalldata(p BuildShieldCalldataParams) (*Calldata, error) {
	f, err := ValidateShieldRequest(p)
	if err != nil {
		return nil, err
	}
	if err := ValidateFundingSubmission(f, p.Caller, p.Context.Now); err != nil {
		return nil, err
	}
	p.Params.FeeAmount = feeAmount(p.Params.FeeAmount)
	data, err := PackProtocolCalldata("shield", p.Params, p.Outputs, p.Deliveries, f)
	if err != nil {
		return nil, err
	}
	return &Calldata{Data: data, ApproveAmount: new(big.Int).Set(f.Amount)}, nil
}
