package session

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"

	"github.com/invisi-lab/invisi-sdk-go/rpc/broadcaster"
	sdk "github.com/invisi-lab/invisi-sdk-go/sdk"
)

func (o *Operation) WithBroadcaster(ctx context.Context, c *broadcaster.Client, opt sdk.WithBroadcasterOptions) (*Operation, error) {
	if opt.MaxFeeAmount != nil {
		opt.MaxFeeAmount = new(big.Int).Set(opt.MaxFeeAmount)
	}
	return o.stage(ctx, func(next *Operation) error { return next.withBroadcaster(ctx, c, opt) }, true)
}

func (o *Operation) commitBroadcasterPlan(ctx context.Context) error {

	encoded, e := json.Marshal(o.plan)
	if e != nil {
		return e
	}
	return o.session.options.Store.Update(ctx, o.session.scope, func(state *WalletState) error {
		record, ok := state.Operations[o.plan.ID]
		if !ok || record.WorkerID != o.workerID || state.Invalidated {
			return errors.New("stale operation worker")
		}
		for _, n := range o.plan.Inputs {
			current, ok := state.Notes[n.Commitment]
			if !ok || current.Spent || (current.ReservedBy != "" && current.ReservedBy != o.plan.ID) || !current.Note.Equal(n.Note) || current.TreeID != n.TreeID || current.LeafIndex != n.LeafIndex {
				return errors.New("fee input changed during preparation")
			}
		}
		record.Inputs = nil
		for _, n := range o.plan.Inputs {
			current := state.Notes[n.Commitment]
			current.ReservedBy = o.plan.ID
			state.Notes[n.Commitment] = current
			record.Inputs = append(record.Inputs, n.Commitment)
		}
		record.Plan = encoded
		state.Operations[o.plan.ID] = record
		return nil
	})
}
