package transaction

import (
	"bytes"
	"context"
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/broadcaster"
)

func (o *Execution) ExportBroadcasterQuoteInternal() (broadcaster.SignedQuote, error) {
	if o.Plan.Broadcaster == nil {
		return broadcaster.SignedQuote{}, errors.New("broadcaster quote unavailable")
	}
	saved, e := Frozen(*o.Plan.Broadcaster)
	if e != nil {
		return broadcaster.SignedQuote{}, e
	}
	saved.Signed.Context = &broadcaster.Trust{ChainID: saved.ChainID, Protocol: saved.Protocol, Signer: saved.Configuration.QuoteSigner}
	return saved.Signed, nil
}

func (o *Execution) ExportBroadcasterQuote(ctx context.Context) (broadcaster.SignedQuote, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if e := o.CheckLease(ctx); e != nil {
		return broadcaster.SignedQuote{}, e
	}
	return o.ExportBroadcasterQuoteInternal()
}

func (p *PreparedTransaction) ExportBroadcasterQuote(ctx context.Context) (broadcaster.SignedQuote, error) {
	p.lifecycle.mu.Lock()
	defer p.lifecycle.mu.Unlock()
	if e := p.current(ctx); e != nil {
		return broadcaster.SignedQuote{}, e
	}
	return p.operation.ExportBroadcasterQuoteInternal()
}

func (o *Execution) ExportBroadcasterIntent(ctx context.Context) ([]byte, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if e := o.CheckLease(ctx); e != nil {
		return nil, e
	}
	if o.Plan.Broadcaster == nil {
		return nil, errors.New("broadcaster intent unavailable")
	}
	return append([]byte{}, o.Plan.Broadcaster.Intent...), nil
}

func (p *PreparedTransaction) ExportBroadcasterIntent(ctx context.Context) ([]byte, error) {
	p.lifecycle.mu.Lock()
	defer p.lifecycle.mu.Unlock()
	if e := p.current(ctx); e != nil {
		return nil, e
	}
	if p.operation.Plan.Broadcaster == nil {
		return nil, errors.New("broadcaster intent unavailable")
	}
	return append([]byte{}, p.operation.Plan.Broadcaster.Intent...), nil
}

func (o *Execution) SubmitBroadcaster(ctx context.Context, c *broadcaster.Client) (broadcaster.Status, error) {
	var empty broadcaster.Status
	if c == nil || o.Plan.Broadcaster == nil {
		return empty, errors.New("selected broadcaster required")
	}
	if c.Configuration() != o.Plan.Broadcaster.Configuration {
		return empty, errors.New("broadcaster differs from quoted registration")
	}
	if o.Status != "ready" || len(o.Data) == 0 {
		return empty, errors.New("ready transaction required")
	}
	if e := o.ValidateRecovered(); e != nil {
		return empty, e
	}
	if e := o.ValidateExport(ctx); e != nil {
		return empty, e
	}
	signed, e := o.ExportBroadcasterQuoteInternal()
	if e != nil {
		return empty, e
	}
	if e = broadcaster.VerifyQuote(signed, *signed.Context, o.Data, 0); e != nil {
		return empty, e
	}
	if o.Runtime.OnExternalize != nil {
		if e = o.Runtime.OnExternalize(ctx, o); e != nil {
			return empty, e
		}
	}

	o.Status = "uncertain"
	if e = o.Persist(context.WithoutCancel(ctx), o.Status); e != nil {
		return empty, e
	}
	result, sendErr := c.Submit(ctx, o.Data, signed, o.Plan.ID)
	if result.TransactionHash != "" {
		o.TransactionHash = common.HexToHash(result.TransactionHash)
	}
	if sendErr == nil {
		switch result.State {
		case "submitted", "confirmed", "reverted":
			o.Status = "submitted"
		default:
			o.Status = "uncertain"
		}
	}
	persistErr := o.Persist(context.WithoutCancel(ctx), o.Status)
	return result, errors.Join(sendErr, persistErr)
}

func (p *PreparedTransaction) SubmitBroadcaster(ctx context.Context, c *broadcaster.Client) (broadcaster.Status, error) {
	p.lifecycle.mu.Lock()
	defer p.lifecycle.mu.Unlock()
	if e := p.current(ctx); e != nil {
		return broadcaster.Status{}, e
	}
	result, e := p.operation.SubmitBroadcaster(ctx, c)
	return result, e
}

func (o *Execution) ValidateBroadcasterPlan() error {
	saved := o.Plan.Broadcaster
	if saved == nil {
		return nil
	}
	if saved.ChainID == nil || saved.ChainID.Cmp(o.Plan.Context.ChainID) != 0 || saved.Protocol != o.Plan.Context.ProtocolAddress || saved.Signed.Quote.FeeRecipient != saved.Configuration.FeeRecipient {
		return errors.New("broadcaster quote context mismatch")
	}
	if _, e := broadcaster.NewClient(saved.Configuration); e != nil {
		return e
	}
	data, e := o.BroadcasterIntent()
	if e != nil {
		return e
	}
	if !bytes.Equal(data, saved.Intent) {
		return errors.New("broadcaster intent differs from Frozen operation")
	}
	return broadcaster.VerifyQuote(saved.Signed, broadcaster.Trust{ChainID: saved.ChainID, Protocol: saved.Protocol, Signer: saved.Configuration.QuoteSigner}, data, 0)
}
