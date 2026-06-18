package receiver

import (
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-labs/invisi-sdk-go/core"
	"github.com/invisi-labs/invisi-sdk-go/ipp"
)

// VerifyVoucher confirms the voucher binds funds to the receiver's own address
// and is internally consistent. MUST pass before withdraw.
func VerifyVoucher(voucher ipp.IPPVoucher, receivingAddress common.Address, chainID *big.Int) (bool, error) {
	p := voucher.Payload

	nk, err := core.HexToElement(p.NullifierKey)
	if err != nil {
		return false, err
	}
	ic, err := core.ComputeIdentityCommitment(nk, core.BigIntToElement(receivingAddress.Big()))
	if err != nil {
		return false, err
	}
	if !strings.EqualFold(core.ElementToHex(ic), p.IdentityCommitment) {
		return false, nil
	}

	token, err := core.HexToElement(p.TokenAddress)
	if err != nil {
		return false, err
	}
	amount, ok := new(big.Int).SetString(p.Amount, 10)
	if !ok {
		return false, nil
	}
	nonce, err := core.HexToElement(p.Nonce)
	if err != nil {
		return false, err
	}
	pc, err := core.ComputePaymentCommitment(
		core.BigIntToElement(chainID), ic, token,
		core.BigIntToElement(amount), nonce,
	)
	if err != nil {
		return false, err
	}
	return strings.EqualFold(core.ElementToHex(pc), p.PaymentCommitment), nil
}

// VoucherToWithdrawRequest maps voucher payload to the withdraw request body.
// nullifierKey stays client-side (local proof generation), never in the request.
func VoucherToWithdrawRequest(voucher ipp.IPPVoucher) ipp.WithdrawRequest {
	p := voucher.Payload
	return ipp.WithdrawRequest{
		IdentityCommitment: p.IdentityCommitment,
		TokenAddress:       p.TokenAddress,
		TreeID:             p.TreeID,
		PaymentCommitments: []string{p.PaymentCommitment},
	}
}
