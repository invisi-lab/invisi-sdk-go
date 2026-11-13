package session

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/invisi-lab/invisi-sdk-go/rpc/broadcaster"
)

func (o *Operation) bindBroadcasterQuote() *Operation {
	broadcaster.BindQuoteTarget(&o.QuoteTarget, func(ctx context.Context, c *broadcaster.Client, opt broadcaster.QuoteOptions) error {
		next, err := o.stage(ctx, func(next *Operation) error { return next.quoteBroadcaster(ctx, c, opt) }, true)
		if err != nil {
			return err
		}
		o.mu.Lock()
		defer o.mu.Unlock()
		o.session = next.session
		o.apply(next.execution())
		return nil
	})
	return o
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
