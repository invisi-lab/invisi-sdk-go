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

func EncodeCancelCreditTypedData(c TransactionContext, p CancelCreditParams) (apitypes.TypedData, error) {
	base, err := CancelCreditBaseInputs(c, p)
	if err != nil {
		return apitypes.TypedData{}, err
	}
	return EncodeWalletAuthorizationTypedData(c, core.WalletCancelCredit, 1, base)
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
