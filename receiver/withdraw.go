package receiver

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/payer"
	"math/big"
)

// BuildUnsignedWithdrawalCalldata fully consumes supplied notes to their common bound owner.
func BuildUnsignedWithdrawalCalldata(p BuildUnsignedWithdrawalCalldataParams) (*payer.Calldata, error) {
	r, err := payer.NormalizeTransactRequest(p.Context, p.Params, nil, p.Funding)
	if err != nil {
		return nil, err
	}
	if len(p.Params.OutputCommitments) != 0 || r.Funding.Amount.Sign() != 0 || len(p.SourceNotes) == 0 || len(p.SourceNotes) != len(p.Params.Nullifiers) {
		return nil, fmt.Errorf("unsigned withdrawal requires inputs, no outputs and no funding")
	}
	owner := p.SourceNotes[0].ReceivingAddress
	ownerBig := fieldBig(owner)
	if err := payer.CheckUint(ownerBig, 160, "receiving address"); err != nil {
		return nil, err
	}
	total := new(big.Int)
	for i, n := range p.SourceNotes {
		if !n.ReceivingAddress.Equal(&owner) || fieldBig(n.TokenAddress).Cmp(p.Context.TokenAddress.Big()) != 0 || n.Amount.IsZero() {
			return nil, fmt.Errorf("source note owner/token/amount mismatch")
		}
		if err := payer.CheckUint(fieldBig(n.Amount), 128, "note amount"); err != nil {
			return nil, err
		}
		nullifier, err := core.ComputeNoteNullifier(core.BigIntToElement(p.Context.ChainID), n.NoteSecret, n.ReceivingAddress, n.TokenAddress, n.Amount, n.Nonce)
		if err != nil {
			return nil, err
		}
		if nullifier.Bytes() != p.Params.Nullifiers[i] {
			return nil, fmt.Errorf("source note nullifier mismatch")
		}
		total.Add(total, fieldBig(n.Amount))
	}
	if common.BigToAddress(ownerBig) != p.Params.WithdrawalAddress || p.Params.WithdrawalAmount.Cmp(total) != 0 {
		return nil, fmt.Errorf("unsigned withdrawal must pay full input balance to bound owner")
	}
	return payer.BuildTransactCalldata(payer.BuildTransactCalldataParams{Context: p.Context, Proof: p.Proof, Params: p.Params, Funding: p.Funding, Caller: p.Caller})
}
