package receiver

import (
	"fmt"
	"github.com/invisi-lab/invisi-sdk-go/payer"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/ipp"
)

// VerifyVoucher confirms the voucher binds funds to the receiver's own address
// and is internally consistent. MUST pass before withdraw.
// The chain ID is always taken from voucher.Network (CAIP-2), so recomputing
// NoteCommitment against it also verifies the envelope's network field is
// unaltered — a tampered Network makes the recomputed commitment mismatch.
// The voucher carries ReceivingAddress directly; identityCommitment is recomputed
// locally from ReceivingAddress + NoteSecret (one-way hash — not reversible).
func VerifyVoucher(voucher ipp.IPPVoucher, receivingAddress common.Address) (bool, error) {
	if err := ipp.ValidateIPPVoucher(voucher); err != nil {
		return false, err
	}
	p := voucher.Payload
	if !common.IsHexAddress(p.ReceivingAddress) || !common.IsHexAddress(p.TokenAddress) {
		return false, fmt.Errorf("invalid voucher address")
	}

	derivedChainID, err := ipp.CAIP2ToChainID(voucher.Network)
	if err != nil {
		return false, err
	}

	if err := payer.CheckUint(derivedChainID, 64, "chainID"); err != nil {
		return false, err
	}
	// Check 1: the voucher's receiving address must be this receiver's own address —
	// the settle circuit sends funds there, so a mismatch means funds go elsewhere.
	if common.HexToAddress(p.ReceivingAddress) != receivingAddress {
		return false, nil
	}

	nk, err := core.HexToElement(p.NoteSecret)
	if err != nil {
		return false, err
	}
	// Recompute identityCommitment locally (the voucher no longer carries it).
	ic, err := core.ComputeIdentityCommitment(nk, core.BigIntToElement(common.HexToAddress(p.ReceivingAddress).Big()))
	if err != nil {
		return false, err
	}

	token, err := core.HexToElement(p.TokenAddress)
	if err != nil {
		return false, err
	}
	amount, ok := new(big.Int).SetString(p.Amount, 10)
	if !ok {
		return false, nil
	}
	if err := payer.CheckUint(amount, 128, "amount"); err != nil {
		return false, err
	}
	if amount.Sign() == 0 {
		return false, fmt.Errorf("amount must be positive")
	}
	nonce, err := core.HexToElement(p.Nonce)
	if err != nil {
		return false, err
	}
	pc, err := core.ComputeNoteCommitment(
		core.BigIntToElement(derivedChainID), ic, token,
		core.BigIntToElement(amount), nonce,
	)
	if err != nil {
		return false, err
	}
	return strings.EqualFold(core.ElementToHex(pc), p.NoteCommitment), nil
}

// VoucherToWithdrawRequest maps voucher payload to the withdraw request body.
// Network, Operator, and TokenAddress form the (chain, operator, token) instance
// triple the server needs to route the request. Network comes from the voucher
// envelope (CAIP-2 chain id) — IPPVoucherV1 carries no network field of its own.
// IdentityCommitment is omitted (server-advisory — the server locates leaves by
// note commitments). ReceivingAddress and noteSecret stay client-side (local
// proof generation), never in the request.
func VoucherToWithdrawRequest(voucher ipp.IPPVoucher) ipp.WithdrawRequest {
	p := voucher.Payload
	return ipp.WithdrawRequest{
		Network:         voucher.Network,
		TokenAddress:    p.TokenAddress,
		Operator:        p.Operator,
		TreeID:          p.TreeID,
		NoteCommitments: []string{p.NoteCommitment},
	}
}
