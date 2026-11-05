package sdk

import (
	"bytes"
	"context"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/broadcaster"
)

func (o *Operation) exportBroadcasterQuote() (broadcaster.SignedQuote, error) {
	if o.plan.Broadcaster == nil {
		return broadcaster.SignedQuote{}, errors.New("broadcaster quote unavailable")
	}
	saved, e := frozen(*o.plan.Broadcaster)
	if e != nil {
		return broadcaster.SignedQuote{}, e
	}
	saved.Signed.Context = &broadcaster.Trust{ChainID: saved.ChainID, Protocol: saved.Protocol, Signer: saved.Configuration.QuoteSigner}
	return saved.Signed, nil
}
func (o *Operation) ExportBroadcasterQuote(ctx context.Context) (broadcaster.SignedQuote, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if e := o.checkLease(ctx); e != nil {
		return broadcaster.SignedQuote{}, e
	}
	return o.exportBroadcasterQuote()
}
func (p *PreparedTransaction) ExportBroadcasterQuote(ctx context.Context) (broadcaster.SignedQuote, error) {
	p.lifecycle.mu.Lock()
	defer p.lifecycle.mu.Unlock()
	if e := p.current(ctx); e != nil {
		return broadcaster.SignedQuote{}, e
	}
	return p.operation.exportBroadcasterQuote()
}
func (o *Operation) ExportBroadcasterIntent(ctx context.Context) ([]byte, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if e := o.checkLease(ctx); e != nil {
		return nil, e
	}
	if o.plan.Broadcaster == nil {
		return nil, errors.New("broadcaster intent unavailable")
	}
	return append([]byte{}, o.plan.Broadcaster.Intent...), nil
}
func (p *PreparedTransaction) ExportBroadcasterIntent(ctx context.Context) ([]byte, error) {
	p.lifecycle.mu.Lock()
	defer p.lifecycle.mu.Unlock()
	if e := p.current(ctx); e != nil {
		return nil, e
	}
	if p.operation.plan.Broadcaster == nil {
		return nil, errors.New("broadcaster intent unavailable")
	}
	return append([]byte{}, p.operation.plan.Broadcaster.Intent...), nil
}
func (o *Operation) submitBroadcaster(ctx context.Context, c *broadcaster.Client) (broadcaster.Status, error) {
	var empty broadcaster.Status
	if c == nil || o.plan.Broadcaster == nil {
		return empty, errors.New("selected broadcaster required")
	}
	if c.Configuration() != o.plan.Broadcaster.Configuration {
		return empty, errors.New("broadcaster differs from quoted registration")
	}
	if o.status != "ready" || len(o.data) == 0 {
		return empty, errors.New("ready transaction required")
	}
	if e := o.validateRecovered(); e != nil {
		return empty, e
	}
	if e := o.validateExport(ctx); e != nil {
		return empty, e
	}
	signed, e := o.exportBroadcasterQuote()
	if e != nil {
		return empty, e
	}
	if e = broadcaster.VerifyQuote(signed, *signed.Context, o.data, 0); e != nil {
		return empty, e
	}
	if o.session.explicitInventory == nil {
		if e = o.session.options.Store.Update(ctx, o.session.scope, func(state *WalletState) error {
			r, ok := state.Operations[o.plan.ID]
			if !ok || r.WorkerID != o.workerID || state.Invalidated {
				return errors.New("stale operation worker")
			}
			r.Externalized = true
			r.Status = "uncertain"
			state.Operations[o.plan.ID] = r
			return nil
		}); e != nil {
			return empty, e
		}
	}
	o.status = "uncertain"
	if e = o.persist(context.WithoutCancel(ctx), o.status); e != nil {
		return empty, e
	}
	result, sendErr := c.Submit(ctx, o.data, signed, o.plan.ID)
	if result.TransactionHash != "" {
		o.transactionHash = common.HexToHash(result.TransactionHash)
	}
	if sendErr == nil {
		switch result.State {
		case "submitted", "confirmed", "reverted":
			o.status = "submitted"
		default:
			o.status = "uncertain"
		}
	}
	persistErr := o.persist(context.WithoutCancel(ctx), o.status)
	return result, errors.Join(sendErr, persistErr)
}

// SubmitBroadcaster preserves reservations after any potentially externalized request.
func (o *Operation) SubmitBroadcaster(ctx context.Context, c *broadcaster.Client) (broadcaster.Status, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if e := o.checkLease(ctx); e != nil {
		return broadcaster.Status{}, e
	}
	return o.submitBroadcaster(ctx, c)
}
func (p *PreparedTransaction) SubmitBroadcaster(ctx context.Context, c *broadcaster.Client) (broadcaster.Status, error) {
	p.lifecycle.mu.Lock()
	defer p.lifecycle.mu.Unlock()
	if e := p.current(ctx); e != nil {
		return broadcaster.Status{}, e
	}
	result, e := p.operation.submitBroadcaster(ctx, c)
	return result, e
}

func (o *Operation) validateBroadcasterPlan() error {
	saved := o.plan.Broadcaster
	if saved == nil {
		return nil
	}
	if saved.ChainID == nil || saved.ChainID.Cmp(o.plan.Context.ChainID) != 0 || saved.Protocol != o.plan.Context.ProtocolAddress || saved.Signed.Quote.FeeRecipient != saved.Configuration.FeeRecipient {
		return errors.New("broadcaster quote context mismatch")
	}
	if _, e := broadcaster.NewClient(saved.Configuration); e != nil {
		return e
	}
	data, e := o.broadcasterIntent()
	if e != nil {
		return e
	}
	if !bytes.Equal(data, saved.Intent) {
		return errors.New("broadcaster intent differs from frozen operation")
	}
	return broadcaster.VerifyQuote(saved.Signed, broadcaster.Trust{ChainID: saved.ChainID, Protocol: saved.Protocol, Signer: saved.Configuration.QuoteSigner}, data, 0)
}
