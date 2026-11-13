package session

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/credit"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
	"github.com/invisi-lab/invisi-sdk-go/rpc/broadcaster"
)

func (o *Operation) quoteBroadcaster(ctx context.Context, client *broadcaster.Client, options credit.BroadcasterQuoteOptions) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	if err := o.checkFresh(ctx); err != nil {
		return err
	}
	before, err := o.prepared.ExportIntent()
	if err != nil {
		return err
	}
	if err = client.Quote(ctx, o.prepared, options); err != nil {
		return err
	}
	quoted, err := o.prepared.ExportIntent()
	if err == nil {
		err = o.update(ctx, func(record *Journal) error {
			if record.Externalized || record.Proof != nil || record.Status != "prepared" {
				return fmt.Errorf("select broadcaster before proving or externalization")
			}
			record.Intent = quoted
			return nil
		})
	}
	if err != nil {
		restored, restoreErr := credit.Restore(ctx, before, nil, nil)
		if restoreErr != nil {
			return restoreErr
		}
		o.prepared = restored
	}
	return err
}

// submitViaBroadcaster records uncertainty and exact calldata before making the submission call.
func (o *Operation) submitViaBroadcaster(ctx context.Context, client *broadcaster.Client) (broadcaster.Status, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	snapshot, err := o.session.store.ReadSnapshot(ctx)
	if err != nil {
		return broadcaster.Status{}, err
	}
	record, ok := snapshot.Operations[o.id]
	if !ok || record.WorkerID != o.worker {
		return broadcaster.Status{}, fmt.Errorf("Credit operation lease changed")
	}
	if record.Externalized {
		binding := record.Intent.Broadcaster
		if client == nil || binding == nil || binding.Configuration != client.Configuration() {
			return broadcaster.Status{}, fmt.Errorf("use the originally selected broadcaster")
		}
		if record.BroadcasterRequestID != "" {
			return client.Status(ctx, record.BroadcasterRequestID)
		}
		request := record.BroadcasterRequest
		if request == nil {
			quote, e := o.prepared.ExportBroadcasterQuote()
			if e != nil {
				return broadcaster.Status{}, e
			}
			value, e := client.PrepareSubmission(record.Calldata, quote, nil, nil)
			if e != nil {
				return broadcaster.Status{}, e
			}
			request = &value
		}
		status, e := client.SubmitPrepared(ctx, *request)
		if e != nil {
			return status, e
		}
		e = o.update(ctx, func(r *Journal) error {
			r.BroadcasterRequestID = status.ID
			if status.TransactionHash != "" {
				r.TransactionHash = common.HexToHash(status.TransactionHash)
			}
			return nil
		})
		return status, e
	}
	if err := o.validateSubmission(ctx); err != nil {
		return broadcaster.Status{}, err
	}
	data, quote, err := o.prepared.ExportBroadcasterSubmission(client)
	if err != nil {
		return broadcaster.Status{}, err
	}
	if _, err = client.SubmissionRequirements(ctx, quote); err != nil {
		return broadcaster.Status{}, err
	}
	request, err := client.PrepareSubmission(data, quote, nil, nil)
	if err != nil {
		return broadcaster.Status{}, err
	}
	if err = o.update(ctx, func(record *Journal) error {
		record.BroadcasterRequest = &request
		if len(record.Calldata) > 0 && string(record.Calldata) != string(data) {
			return fmt.Errorf("Credit calldata changed after externalization")
		}
		record.Calldata = append([]byte(nil), data...)
		record.Externalized = true
		record.Status = "uncertain"
		return nil
	}); err != nil {
		return broadcaster.Status{}, err
	}
	status, err := client.SubmitPrepared(ctx, request)
	if err == nil {
		hash := common.HexToHash(status.TransactionHash)
		persistErr := o.update(ctx, func(record *Journal) error {
			if status.TransactionHash != "" && record.TransactionHash != (common.Hash{}) && record.TransactionHash != hash {
				return fmt.Errorf("conflicting Credit submission hash")
			}
			record.BroadcasterRequestID = status.ID
			if status.TransactionHash != "" {
				record.TransactionHash = hash
			}
			return nil
		})
		if persistErr != nil {
			return status, persistErr
		}
	}
	return status, err
}
func (o *Operation) ExportSigningRequest(ctx context.Context) (primitives.SigningRequest, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if err := o.checkFresh(ctx); err != nil {
		return primitives.SigningRequest{}, err
	}
	return o.prepared.ExportSigningRequest()
}
func (o *Operation) WithWalletSignature(ctx context.Context, signature []byte) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	if err := o.checkFresh(ctx); err != nil {
		return err
	}
	return o.prepared.WithWalletSignature(signature)
}
func (o *Operation) Preview(ctx context.Context) (credit.Preview, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	state, err := o.session.store.ReadSnapshot(ctx)
	if err != nil {
		return credit.Preview{}, err
	}
	record, ok := state.Operations[o.id]
	if !ok || record.WorkerID != o.worker {
		return credit.Preview{}, fmt.Errorf("Credit operation lease changed")
	}
	return o.prepared.Preview(), nil
}
func (o *Operation) ExportBroadcasterQuote(ctx context.Context) (broadcaster.SignedQuote, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if err := o.check(ctx); err != nil {
		return broadcaster.SignedQuote{}, err
	}
	return o.prepared.ExportBroadcasterQuote()
}

func (o *Operation) bindBroadcasterQuote() *Operation {
	broadcaster.BindOperationTarget(&o.OperationTarget, func(ctx context.Context, c *broadcaster.Client, opt broadcaster.QuoteOptions) error {
		return o.quoteBroadcaster(ctx, c, opt)
	}, o.submitViaBroadcaster, o.cancelViaBroadcaster)
	return o
}

func (o *Operation) cancelViaBroadcaster(ctx context.Context, client *broadcaster.Client) (broadcaster.Status, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	snapshot, err := o.session.store.ReadSnapshot(ctx)
	if err != nil {
		return broadcaster.Status{}, err
	}
	record, ok := snapshot.Operations[o.id]
	if !ok || record.WorkerID != o.worker {
		return broadcaster.Status{}, fmt.Errorf("Credit operation lease changed")
	}
	b := record.Intent.Broadcaster
	if client == nil || b == nil || b.Configuration != client.Configuration() || b.Quote.Quote.Schedule == nil || b.CancelToken == "" {
		return broadcaster.Status{}, fmt.Errorf("scoped scheduled cancellation handle required")
	}
	if record.BroadcasterRequestID == "" {
		return broadcaster.Status{}, fmt.Errorf("submission ID unknown; retry the original submission")
	}
	result := broadcaster.Status{ID: record.BroadcasterRequestID, State: "cancelled"}
	if record.TransactionHash != (common.Hash{}) || record.Status == "confirmed" || record.Status == "reverted" {
		return result, fmt.Errorf("Credit submission already has transaction hash")
	}
	if record.Status == "cancelled" && record.CancellationConfirmed {
		return result, nil
	}
	if !record.CancellationConfirmed {
		result, err = client.CancelTransaction(ctx, record.BroadcasterRequestID, b.CancelToken)
		if err != nil {
			return result, err
		}
		if result.State != "cancelled" {
			return result, fmt.Errorf("cancellation was not acknowledged")
		}
		if result.TransactionHash != "" {
			return result, fmt.Errorf("cancelled submission has transaction hash")
		}
		if err = o.update(context.WithoutCancel(ctx), func(r *Journal) error {
			if r.TransactionHash != (common.Hash{}) {
				return fmt.Errorf("Credit submission already has transaction hash")
			}
			r.CancellationConfirmed = true
			return nil
		}); err != nil {
			return result, err
		}
	}
	if o.session.checkCancellation == nil {
		return result, fmt.Errorf("canonical Credit cancellation checker required")
	}
	if err = o.session.checkCancellation(ctx, record.Intent); err != nil {
		return result, err
	}
	err = o.session.store.Update(ctx, func(state *State) error {
		r, ok := state.Operations[o.id]
		if !ok || r.WorkerID != o.worker || !r.CancellationConfirmed || r.TransactionHash != (common.Hash{}) {
			return fmt.Errorf("Credit cancellation journal changed")
		}
		for _, input := range r.Intent.Inputs {
			commitment, e := input.State.Commitment()
			if e != nil {
				return e
			}
			role := credit.Spender
			if r.Intent.Update != nil {
				role = credit.Owner
			}
			g, ok := state.Grants[GrantKey(common.Hash(commitment.Bytes()), role)]
			if !ok || g.ReservedBy != o.id || g.Consumed || g.Revoked {
				return fmt.Errorf("cancelled Credit input unavailable")
			}
		}
		for key, g := range state.Grants {
			if g.ReservedBy == o.id {
				g.ReservedBy = ""
				state.Grants[key] = g
			}
		}
		r.Status = "cancelled"
		state.Operations[o.id] = r
		return nil
	})
	return result, err
}
