package session

import (
	"context"
	"errors"

	"github.com/invisi-lab/invisi-sdk-go/rpc/broadcaster"
)

// submitViaBroadcaster preserves reservations after any potentially externalized request.
func (o *Operation) submitViaBroadcaster(ctx context.Context, c *broadcaster.Client) (broadcaster.Status, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if e := o.checkLease(ctx); e != nil {
		return broadcaster.Status{}, e
	}
	return o.submitBroadcaster(ctx, c)
}

// cancelViaBroadcaster releases reservations only after an explicit cancellation and canonical state refresh.
func (o *Operation) cancelViaBroadcaster(ctx context.Context, c *broadcaster.Client) (broadcaster.Status, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	saved := o.plan.Broadcaster
	if o.status == "cancelled" && saved != nil && saved.CancellationConfirmed && c != nil && c.Configuration() == saved.Configuration {
		state, err := o.session.fencedSnapshot(ctx)
		if err != nil {
			return broadcaster.Status{}, err
		}
		record, ok := state.Operations[o.plan.ID]
		if state.Invalidated || !ok || record.WorkerID != o.workerID || record.Status != "cancelled" || record.TransactionHash != ([32]byte{}) {
			return broadcaster.Status{}, errors.New("cancelled operation changed")
		}
		return broadcaster.Status{ID: saved.RequestID, State: "cancelled"}, nil
	}
	if e := o.checkLease(ctx); e != nil {
		return broadcaster.Status{}, e
	}
	if c == nil || saved == nil || saved.Signed.Quote.Schedule == nil || saved.CancelToken == "" || c.Configuration() != saved.Configuration {
		return broadcaster.Status{}, errors.New("scoped scheduled cancellation handle required")
	}
	if saved.RequestID == "" {
		if saved.Request == nil {
			return broadcaster.Status{}, errors.New("broadcaster submission must be attempted before cancellation")
		}
		return broadcaster.Status{}, errors.New("broadcaster submission ID unknown; retry the original frozen client.Submit call to recover it before cancellation")
	}
	result := broadcaster.Status{ID: saved.RequestID, State: "cancelled"}
	var e error
	if !saved.CancellationConfirmed {
		result, e = c.CancelTransaction(ctx, saved.RequestID, saved.CancelToken)
		if e != nil {
			return result, e
		}
		if result.State != "cancelled" {
			return result, errors.New("cancellation was not acknowledged")
		}
		if result.TransactionHash != "" {
			return result, errors.New("cancelled submission has transaction hash")
		}
		saved.CancellationConfirmed = true
		if e = o.persist(context.WithoutCancel(ctx), o.status); e != nil {
			saved.CancellationConfirmed = false
			return result, e
		}
	}
	if e = o.session.RefreshState(ctx); e != nil {
		return result, e
	}
	e = o.session.options.Store.Update(ctx, o.session.scope, func(state *WalletState) error {
		record, ok := state.Operations[o.plan.ID]
		if !ok || record.WorkerID != o.workerID || state.Invalidated || state.CheckedThrough == nil {
			return errors.New("stale cancellation refresh")
		}
		for _, input := range o.plan.Inputs {
			n, ok := state.Notes[input.Commitment]
			if !ok || n.Spent || n.CheckedThrough == nil || *n.CheckedThrough != *state.CheckedThrough || n.ReservedBy != o.plan.ID {
				return errors.New("cancelled input unavailable after refresh")
			}
		}
		for _, input := range o.plan.Inputs {
			n := state.Notes[input.Commitment]
			n.ReservedBy = ""
			state.Notes[input.Commitment] = n
		}
		record.Status = "cancelled"
		record.ResolvedAt = state.CheckedThrough
		state.Operations[o.plan.ID] = record
		return nil
	})
	if e == nil {
		o.status = "cancelled"
	}
	return result, e
}
