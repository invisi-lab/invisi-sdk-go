package transaction

import (
	"context"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
)

func (s *Runtime) PrepareDrawCredit(ctx context.Context, p DrawCreditPreparation) (result *Execution, failure error) {
	defer func() { failure = WrapPreparation("prepareDrawCredit", failure) }()
	var err error
	p, err = Frozen(p)
	if err != nil {
		return nil, err
	}
	c, block, deadline, e := s.Frame(ctx, p.TokenAddress, p.Deadline)
	if e != nil {
		return nil, e
	}
	profile, e := protocol.SelectCreditDrawCircuit(len(p.Credits))
	if e != nil {
		return nil, e
	}
	params := protocol.CreditDrawParams{Token: p.TokenAddress, Spender: s.Wallet.identity, TotalAmount: new(big.Int), Deadline: new(big.Int).SetUint64(deadline)}
	inputs := []CreditRecord{}
	memberships := []protocolrpc.Membership{}
	deliveries := []protocol.Delivery{}
	amounts := []*big.Int{}
	seen := map[common.Hash]bool{}
	for _, entry := range p.Credits {
		if err := ValidatePayload(entry.Payload); err != nil {
			return nil, err
		}
		if seen[entry.Credit] {
			return nil, errors.New("duplicate Credit")
		}
		seen[entry.Credit] = true
		if e = Positive(entry.Amount, 126); e != nil {
			return nil, e
		}
		old, m, e := s.CreditInput(ctx, p.TokenAddress, entry.Credit, "spender", block)
		if e != nil {
			return nil, e
		}
		note := old.State.Note
		if core.ElementToBigInt(note.Expiry).Uint64() <= block.Timestamp || entry.Amount.Cmp(core.ElementToBigInt(note.Amount)) > 0 {
			return nil, errors.New("draw exceeds an unexpired Credit")
		}
		updated, e := core.ApplyCreditDrawUsage(note, core.BigIntToElement(entry.Amount), block.Timestamp)
		if e != nil {
			return nil, e
		}
		remaining := new(big.Int).Sub(core.ElementToBigInt(note.Amount), entry.Amount)
		remainder := common.Hash{}
		recovery := common.Hash{}
		if remaining.Sign() > 0 {
			nonce, e := core.DeriveRemainderNonce(note.ChainID, note.ProtocolAddress, note.TokenAddress, core.BigIntToElement(old.Nullifier.Big()), note.CreditSharedKey)
			if e != nil {
				return nil, e
			}
			child := old.State
			child.Note = updated
			child.Note.Payload = ClonePayload(entry.Payload)
			child.Note.Amount = core.BigIntToElement(remaining)
			child.Note.Nonce = nonce
			remainder, e = CreditCommitment(child.Note)
			if e != nil {
				return nil, e
			}
			pad, e := core.ComputeCreditRecoveryPad(note.ChainID, note.ProtocolAddress, note.TokenAddress, core.BigIntToElement(old.Nullifier.Big()), note.CreditSharedKey)
			if e != nil {
				return nil, e
			}
			value := core.BigIntToElement(remaining)
			value.Add(&value, &pad)
			recovery = FieldHash(value)
			ds, e := CreditDeliveries(child)
			if e != nil {
				return nil, e
			}
			deliveries = append(deliveries, ds...)
		} else {
			deliveries = append(deliveries, protocol.Delivery{}, protocol.Delivery{})
		}
		inputs = append(inputs, old)
		memberships = append(memberships, m)
		amounts = append(amounts, new(big.Int).Set(entry.Amount))
		params.Nullifiers = append(params.Nullifiers, old.Nullifier)
		params.RemainderCommitments = append(params.RemainderCommitments, remainder)
		params.RecoveryAmounts = append(params.RecoveryAmounts, recovery)
		params.Expiries = append(params.Expiries, core.ElementToBigInt(note.Expiry))
		params.TotalAmount.Add(params.TotalAmount, entry.Amount)
	}
	sources, e := SourcesFor(memberships)
	if e != nil {
		return nil, e
	}
	params.Sources = sources
	funding, _ := protocol.NormalizeFunding(protocol.FundingParams{})
	return s.PersistPlan(ctx, Plan{Kind: "credit-draw", Target: "evm", Context: c, Block: block, Deadline: deadline, Funding: funding, FundingMode: "none", Capacity: profile.MaxSize, NeedsAuthorization: true, CreditInputs: inputs, CreditMemberships: memberships, DrawCredit: &params, Deliveries: deliveries, DrawAmounts: amounts})
}
