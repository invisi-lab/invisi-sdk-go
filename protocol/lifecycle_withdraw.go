package protocol

import (
	"fmt"
	"math/big"

	"github.com/invisi-lab/invisi-sdk-go/core"
)

// BuildWithdrawalCalldata validates a complete local source balance. The supplied
// proof privately authorizes its selected public payout; this helper does not prove it.
func BuildWithdrawalCalldata(p BuildWithdrawalCalldataParams) (*Calldata, error) {
	r, err := NormalizeTransactRequest(p.Context, p.Params, nil, p.Funding)
	if err != nil {
		return nil, err
	}
	if len(p.Params.OutputCommitments) != 0 || r.Funding.Amount.Sign() != 0 || len(p.SourceNotes) == 0 || len(p.SourceNotes) != len(p.Params.Nullifiers) {
		return nil, fmt.Errorf("full withdrawal requires inputs, no outputs and no funding")
	}
	chain := core.BigIntToElement(p.Context.ChainID)
	protocol := core.BigIntToElement(p.Context.ProtocolAddress.Big())
	identity := p.SourceNotes[0].WalletIdentity
	total := new(big.Int)
	for i, n := range p.SourceNotes {
		if n.ChainID != chain || n.ProtocolAddress != protocol {
			return nil, fmt.Errorf("source note context mismatch")
		}
		if n.WalletIdentity != identity || core.ElementToBigInt(n.TokenAddress).Cmp(p.Context.TokenAddress.Big()) != 0 || n.Amount.IsZero() {
			return nil, fmt.Errorf("source note wallet/token/amount mismatch")
		}
		randomized, err := core.ComputeNoteIdentity(n.WalletIdentity, n.Nonce)
		if err != nil {
			return nil, err
		}
		leaf, err := core.ComputeNoteCommitment(n.ChainID, randomized, n.TokenAddress, n.Amount)
		if err != nil {
			return nil, err
		}
		nullifier, err := core.ComputeNoteNullifier(n.ChainID, n.Nonce, leaf)
		if err != nil {
			return nil, err
		}
		if nullifier.Bytes() != p.Params.Nullifiers[i] {
			return nil, fmt.Errorf("source note nullifier mismatch")
		}
		total.Add(total, core.ElementToBigInt(n.Amount))
	}
	if p.Params.WithdrawalAmount.Cmp(total) != 0 {
		return nil, fmt.Errorf("withdrawal must equal full input balance")
	}
	return BuildTransactCalldata(BuildTransactCalldataParams{Context: p.Context, Proof: p.Proof, Params: p.Params, Funding: p.Funding, Caller: p.Caller})
}
