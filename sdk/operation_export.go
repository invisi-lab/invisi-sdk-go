package sdk

import (
	"context"
	"errors"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	"github.com/invisi-lab/invisi-sdk-go/protocol/contracts"
)

func (o *Operation) sources() []protocol.InputSource {
	p := o.plan
	switch {
	case p.Kind == "transact":
		return p.Params.Sources
	case p.TopUpCredit != nil:
		return append(append([]protocol.InputSource{}, p.TopUpCredit.Creation.Sources...), p.TopUpCredit.CreditSource)
	case p.CreateCredit != nil:
		return p.CreateCredit.Sources
	case p.DrawCredit != nil:
		return p.DrawCredit.Sources
	case p.CancelCredit != nil:
		return []protocol.InputSource{p.CancelCredit.Source}
	}
	return nil
}
func (o *Operation) validateExport(ctx context.Context) error {
	if err := o.session.protocol.ValidateDeployment(ctx, o.session.options.Provider); err != nil {
		return err
	}
	head, err := o.session.options.Provider.HeaderByNumber(ctx, nil)
	if err != nil {
		return err
	}
	if head == nil || head.Number == nil {
		return errors.New("latest block unavailable")
	}
	if o.plan.Broadcaster != nil {
		deadline, ok := new(big.Int).SetString(o.plan.Broadcaster.Signed.Quote.Deadline, 10)
		if !ok || deadline.Cmp(new(big.Int).SetUint64(head.Time)) < 0 || deadline.Cmp(big.NewInt(time.Now().Unix())) < 0 {
			return errors.New("broadcaster quote expired")
		}
	}
	if operationExpired(o.plan, head.Time) {
		return errors.New("operation has expired")
	}
	caller, err := contracts.NewInvisiProtocolCaller(o.plan.Context.ProtocolAddress, o.session.options.Provider)
	if err != nil {
		return err
	}
	options := &bind.CallOpts{Context: ctx, BlockNumber: head.Number}
	if o.plan.FundingMode == "signature" {
		if o.plan.Funding.Deadline == nil || new(big.Int).SetUint64(head.Time).Cmp(o.plan.Funding.Deadline) > 0 {
			return errors.New("funding authorization expired")
		}
		nonce, err := caller.Nonces(options, o.plan.Funding.Payer)
		if err != nil {
			return err
		}
		if nonce.Cmp(o.plan.Funding.SigNonce) != 0 {
			return errors.New("funding nonce changed; prepare again")
		}
	}
	for _, source := range o.sources() {
		known, err := caller.RootHistory(options, source.TreeId, source.MerkleRoot)
		if err != nil {
			return err
		}
		if !known {
			return errors.New("source root is no longer accepted; prepare again")
		}
	}
	nullifiers := []common.Hash{}
	for _, n := range o.plan.Inputs {
		nullifiers = append(nullifiers, n.Nullifier)
	}
	for _, c := range o.plan.CreditInputs {
		nullifiers = append(nullifiers, c.Nullifier)
	}
	for _, nf := range nullifiers {
		spent, err := caller.IsNullifierSpent(options, nf)
		if err != nil {
			return err
		}
		if spent {
			return errors.New("input already spent")
		}
	}
	return ctx.Err()
}

// ExportTransaction exposes only the frozen public call and keeps managed inputs reserved.
func (o *Operation) ExportTransaction(ctx context.Context) (TransactionCall, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if err := o.checkLease(ctx); err != nil {
		return TransactionCall{}, err
	}
	if o.status != "ready" || o.plan.Target == "noir-recursive" || len(o.data) == 0 {
		return TransactionCall{}, errors.New("ready EVM transaction required")
	}
	if err := o.validateExport(ctx); err != nil {
		return TransactionCall{}, err
	}
	err := o.session.options.Store.Update(ctx, o.session.scope, func(state *WalletState) error {
		record, ok := state.Operations[o.plan.ID]
		if !ok || record.WorkerID != o.workerID || state.Invalidated {
			return errors.New("stale operation worker")
		}
		record.Externalized = true
		state.Operations[o.plan.ID] = record
		return nil
	})
	if err != nil {
		return TransactionCall{}, err
	}
	return TransactionCall{To: o.plan.Context.ProtocolAddress, Data: append([]byte(nil), o.data...), Value: new(big.Int)}, nil
}

// TrackTransaction records an association, not confirmation of successful execution.
func (o *Operation) TrackTransaction(ctx context.Context, hash common.Hash) (*Operation, error) {
	if hash == (common.Hash{}) {
		return nil, errors.New("transaction hash required")
	}
	return o.stage(ctx, func(next *Operation) error {
		if next.status != "ready" && next.status != "submitted" && next.status != "uncertain" {
			return errors.New("ready or uncertain operation required")
		}
		next.transactionHash = hash
		next.status = "uncertain"
		return next.persist(ctx, next.status)
	})
}

func operationExpired(p operationPlan, now uint64) bool {
	if now > p.Deadline {
		return true
	}
	timestamp := new(big.Int).SetUint64(now)
	if p.TopUpCredit != nil && timestamp.Cmp(p.TopUpCredit.PreviousExpiry) >= 0 {
		return true
	}
	if p.CreateCredit != nil && timestamp.Cmp(p.CreateCredit.Expiry) >= 0 {
		return true
	}
	if p.DrawCredit != nil {
		if core.CreditPeriods(now) != core.CreditPeriods(p.Block.Timestamp) {
			return true
		}
		for _, expiry := range p.DrawCredit.Expiries {
			if timestamp.Cmp(expiry) >= 0 {
				return true
			}
		}
	}
	return false
}

func (o *Operation) checkDrawPeriod(ctx context.Context) error {
	if o.plan.DrawCredit == nil {
		return nil
	}
	header, err := o.session.options.Provider.HeaderByNumber(ctx, nil)
	if err != nil {
		return err
	}
	if header == nil || header.Number == nil {
		return errors.New("latest block unavailable")
	}
	if core.CreditPeriods(header.Time) != core.CreditPeriods(o.plan.Block.Timestamp) {
		return errors.New("Credit draw crossed a UTC period; prepare again")
	}
	return nil
}
