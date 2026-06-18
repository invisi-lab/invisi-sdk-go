package receiver

import (
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/ipp"
)

// VerifyVoucher confirms the voucher binds funds to the receiver's own address
// and is internally consistent. MUST pass before withdraw.
// chainID is accepted for API compatibility but the authoritative chain ID is
// always derived from voucher.Network via CAIP-2 parsing.
func VerifyVoucher(voucher ipp.IPPVoucher, receivingAddress common.Address, chainID *big.Int) (bool, error) {
	p := voucher.Payload

	derivedChainID, err := ipp.CAIP2ToChainID(voucher.Network)
	if err != nil {
		return false, err
	}

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
		core.BigIntToElement(derivedChainID), ic, token,
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
