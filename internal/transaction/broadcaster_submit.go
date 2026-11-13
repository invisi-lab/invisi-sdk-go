package transaction

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/invisi-lab/invisi-sdk-go/provenance"
	"github.com/invisi-lab/invisi-sdk-go/rpc/broadcaster"
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

func (o *Execution) SubmitViaBroadcaster(ctx context.Context, c *broadcaster.Client) (broadcaster.Status, error) {
	var empty broadcaster.Status
	if c == nil || o.Plan.Broadcaster == nil {
		return empty, errors.New("selected broadcaster required")
	}
	if c.Configuration() != o.Plan.Broadcaster.Configuration {
		return empty, errors.New("broadcaster differs from quoted registration")
	}
	if o.Plan.Broadcaster.CancellationConfirmed {
		return empty, errors.New("broadcaster cancellation already acknowledged")
	}
	retry := o.Status == "uncertain" || o.Status == "submitted"
	if !retry {
		if o.Runtime.Options.Provider == nil {
			return empty, &MissingCapabilityError{Capability: "Chain RPC provider for first broadcaster submission"}
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
	}
	signed, e := o.ExportBroadcasterQuoteInternal()
	if e != nil {
		return empty, e
	}
	required := signed.ProvenanceRequired
	if !retry {
		if _, e = c.SubmissionRequirements(ctx, signed); e != nil {
			return empty, e
		}
		if required && o.Plan.Kind == "transact" && len(o.Plan.Inputs) > 0 && o.Plan.Provenance == nil {
			if e = o.ProveProvenance(ctx); e != nil {
				return empty, e
			}
		}
		if required {
			if e = o.validateProvenance(ctx, o.Plan.Provenance, true); e != nil {
				return empty, e
			}
		}
	}
	source := o.Plan.Provenance

	member, e := o.broadcasterMember()
	if e != nil {
		return empty, e
	}
	var sidecars []provenance.ProofSidecar
	if required && source != nil {
		sidecars = source.Sidecars
	}
	request, e := c.PrepareSubmission(o.Data, signed, sidecars, member)
	if e != nil {
		return empty, e
	}
	if saved := o.Plan.Broadcaster.Request; saved != nil {
		actual, _ := json.Marshal(request)
		expected, _ := json.Marshal(saved)
		if !bytes.Equal(actual, expected) {
			return empty, errors.New("prepared broadcaster submission changed")
		}
	} else if retry {
		return empty, errors.New("frozen broadcaster request unavailable")
	}
	if retry && o.Plan.Broadcaster.RequestID != "" {
		return c.Status(ctx, o.Plan.Broadcaster.RequestID)
	}

	o.Plan.Broadcaster.Request = &request
	if o.Runtime.OnExternalize != nil {
		if e = o.Runtime.OnExternalize(ctx, o); e != nil {
			return empty, e
		}
	}

	o.Status = "uncertain"
	if e = o.Persist(context.WithoutCancel(ctx), o.Status); e != nil {
		return empty, e
	}
	result, sendErr := c.SubmitPrepared(ctx, *o.Plan.Broadcaster.Request)
	if sendErr == nil {
		if o.Plan.Broadcaster.RequestID != "" && o.Plan.Broadcaster.RequestID != result.ID {
			return result, errors.New("broadcaster changed submission ID")
		}
		o.Plan.Broadcaster.RequestID = result.ID
	}
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

func (p *PreparedTransaction) submitViaBroadcaster(ctx context.Context, c *broadcaster.Client) (broadcaster.Status, error) {
	p.lifecycle.mu.Lock()
	defer p.lifecycle.mu.Unlock()
	if e := p.current(ctx); e != nil {
		return broadcaster.Status{}, e
	}
	result, e := p.operation.SubmitViaBroadcaster(ctx, c)
	return result, e
}

func (o *Execution) ValidateBroadcasterPlan() error {
	saved := o.Plan.Broadcaster
	if saved == nil {
		return nil
	}
	if len(saved.RequestID) > 256 || saved.CancellationConfirmed && saved.RequestID == "" {
		return errors.New("invalid retained broadcaster submission ID")
	}
	if schedule := saved.Signed.Quote.Schedule; schedule != nil {
		token := common.FromHex(saved.CancelToken)
		if len(token) != 32 || saved.CancelToken != "0x"+common.Bytes2Hex(token) || crypto.Keccak256Hash(token).Hex() != schedule.CancelTokenHash {
			return errors.New("invalid local cancellation handle")
		}
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

	if e = o.validateProvenance(context.Background(), o.Plan.Provenance, false); e != nil {
		return e
	}
	return broadcaster.VerifyQuote(saved.Signed, broadcaster.Trust{ChainID: saved.ChainID, Protocol: saved.Protocol, Signer: saved.Configuration.QuoteSigner}, data, 0)
}

func (o *Execution) broadcasterMember() (*broadcaster.MemberProof, error) {
	recursive := o.Plan.Target == "noir-recursive"
	if (o.Plan.Broadcaster.Signed.Quote.Execution == "aggregate") != recursive {
		return nil, errors.New("broadcaster execution does not match proof target")
	}
	if !recursive {
		return nil, nil
	}
	if o.Bundle == nil || o.Plan.Kind != "transact" || o.Plan.FundingMode != "none" || len(o.Plan.Inputs) == 0 {
		return nil, errors.New("unfunded private recursive member required")
	}
	b := o.Bundle
	return &broadcaster.MemberProof{FormatVersion: b.FormatVersion, CircuitName: b.CircuitName, Target: b.Target, BBVersion: b.BBVersion, ArtifactSHA256: b.ArtifactSHA256, VKSHA256: b.VKSHA256, ProofHex: "0x" + common.Bytes2Hex(b.Proof), ProofSHA256: b.ProofSHA256, PublicInputs: append([]common.Hash(nil), b.PublicInputs...)}, nil
}

func (o *Execution) validateFrozenBroadcasterRequest() error {
	saved := o.Plan.Broadcaster
	if saved == nil || saved.Request == nil {
		return nil
	}
	client, err := broadcaster.NewClient(saved.Configuration)
	if err != nil {
		return err
	}
	signed, err := o.ExportBroadcasterQuoteInternal()
	if err != nil {
		return err
	}
	member, err := o.broadcasterMember()
	if err != nil {
		return err
	}
	var sidecars []provenance.ProofSidecar
	if signed.ProvenanceRequired && o.Plan.Provenance != nil {
		sidecars = o.Plan.Provenance.Sidecars
	}
	request, err := client.PrepareSubmission(o.Data, signed, sidecars, member)
	if err != nil {
		return err
	}
	actual, err := json.Marshal(request)
	if err != nil {
		return err
	}
	expected, err := json.Marshal(saved.Request)
	if err != nil {
		return err
	}
	if !bytes.Equal(actual, expected) {
		return errors.New("prepared broadcaster submission changed")
	}
	return nil
}

func (p *PreparedTransaction) cancelViaBroadcaster(ctx context.Context, c *broadcaster.Client) (broadcaster.Status, error) {
	p.lifecycle.mu.Lock()
	defer p.lifecycle.mu.Unlock()
	if e := p.current(ctx); e != nil {
		return broadcaster.Status{}, e
	}
	saved := p.operation.Plan.Broadcaster
	if c == nil || saved == nil || saved.Signed.Quote.Schedule == nil || saved.CancelToken == "" || c.Configuration() != saved.Configuration {
		return broadcaster.Status{}, errors.New("scoped scheduled cancellation handle required")
	}
	if saved.RequestID == "" {
		if saved.Request == nil {
			return broadcaster.Status{}, errors.New("broadcaster submission must be attempted before cancellation")
		}
		return broadcaster.Status{}, errors.New("broadcaster submission ID unknown; retry the original frozen client.Submit call to recover it before cancellation")
	}
	if saved.CancellationConfirmed {
		return broadcaster.Status{ID: saved.RequestID, State: "cancelled"}, nil
	}
	result, e := c.CancelTransaction(ctx, saved.RequestID, saved.CancelToken)
	if e == nil && result.State != "cancelled" {
		return result, errors.New("cancellation was not acknowledged")
	}
	if e == nil && result.State == "cancelled" {
		if result.TransactionHash != "" {
			return result, errors.New("cancelled submission has transaction hash")
		}
		previous := p.operation.Status
		saved.CancellationConfirmed = true
		p.operation.Status = "cancelled"
		e = p.operation.Persist(context.WithoutCancel(ctx), "cancelled")
		if e != nil {
			saved.CancellationConfirmed = false
			p.operation.Status = previous
		}
	}
	return result, e
}
