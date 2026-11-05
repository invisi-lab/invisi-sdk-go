package transaction

import (
	"context"
	"math/big"

	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
)

func (s *Runtime) PrepareCancelCredit(ctx context.Context, p CancelCreditPreparation) (result *Execution, failure error) {
	defer func() { failure = WrapPreparation("prepareCancelCredit", failure) }()
	c, block, deadline, e := s.Frame(ctx, p.TokenAddress, p.Deadline)
	if e != nil {
		return nil, e
	}
	old, m, e := s.CreditInput(ctx, p.TokenAddress, p.Credit, "owner", block)
	if e != nil {
		return nil, e
	}
	params := protocol.CancelCreditParams{Token: p.TokenAddress, Source: protocol.InputSource{TreeId: new(big.Int).SetUint64(old.TreeID), MerkleRoot: m.Root}, Nullifier: old.Nullifier, Owner: s.Wallet.identity, Amount: core.ElementToBigInt(old.State.Note.Amount), Deadline: new(big.Int).SetUint64(deadline)}
	funding, _ := protocol.NormalizeFunding(protocol.FundingParams{})
	return s.PersistPlan(ctx, Plan{Kind: "cancel-credit", Target: "evm", Context: c, Block: block, Deadline: deadline, Funding: funding, FundingMode: "none", Capacity: 1, NeedsAuthorization: true, CreditInputs: []CreditRecord{old}, CreditMemberships: []protocolrpc.Membership{m}, CancelCredit: &params})
}
