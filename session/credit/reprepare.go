package session

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-labs/invisi-sdk-go/credit"
	"github.com/invisi-labs/invisi-sdk-go/primitives"
)

type preparationReservationKey struct{}

// PreparationReservation identifies the existing reservation a managed rebuild
// may read. Only ReprepareDraw can attach this capability to a build context.
func PreparationReservation(ctx context.Context) string {
	id, _ := ctx.Value(preparationReservationKey{}).(string)
	return id
}

// ReprepareDraw replaces a stale, known-unsent draw while retaining its inputs.
// The caller supplies a fresh preparation request, then authorizes and proves
// the new statement. Sent or uncertain operations must be reconciled instead.
func (o *Operation) ReprepareDraw(ctx context.Context, request credit.DrawPreparation) (failure error) {
	request = request.Clone()
	o.mu.Lock()
	defer o.mu.Unlock()
	s := o.session
	defer func() {
		if s.wrapPreparation != nil {
			failure = s.wrapPreparation("reprepareCreditDraw", failure)
		}
	}()
	if err := o.check(ctx); err != nil {
		return err
	}
	snapshot, err := s.store.ReadSnapshot(ctx)
	if err != nil {
		return err
	}
	previous, ok := snapshot.Operations[o.id]
	if !ok || previous.WorkerID != o.worker {
		return fmt.Errorf("Credit operation lease changed")
	}
	if previous.Intent.Draw == nil || previous.Status != "prepared" && previous.Status != "proved" && previous.Status != "ready" {
		return fmt.Errorf("only a locally prepared Credit draw can be rebuilt")
	}
	if previous.Externalized || previous.TransactionHash != (common.Hash{}) || previous.BroadcasterRequestID != "" || previous.BroadcasterRequest != nil {
		return fmt.Errorf("externalized Credit draw requires canonical reconciliation")
	}
	if s.chainTimestamp == nil {
		return fmt.Errorf("current Credit chain timestamp reader required")
	}
	now, err := s.chainTimestamp(ctx)
	if err != nil {
		return err
	}
	if now == 0 {
		return fmt.Errorf("current Credit chain timestamp unavailable")
	}
	oldIntent := previous.Intent
	oldIntent.Broadcaster = nil
	old, err := credit.Restore(ctx, oldIntent, nil, nil)
	if err != nil {
		return err
	}
	stale := old.CheckTimestamp(now)
	quoteExpired := false
	if b := previous.Intent.Broadcaster; b != nil {
		deadline, e := strconv.ParseUint(b.Quote.Quote.Deadline, 10, 64)
		quoteExpired = e == nil && deadline < uint64(time.Now().Unix())
	}
	if previous.SubmissionRefusal == "" && !quoteExpired && !errors.Is(stale, credit.ErrCreditPeriodChanged) && !errors.Is(stale, credit.ErrCreditExecutionExpired) {
		return fmt.Errorf("Credit draw usage period, deadline and quote remain current")
	}
	frozen, err := json.Marshal(previous)
	if err != nil {
		return err
	}
	defer clear(frozen)
	if s.build == nil {
		return fmt.Errorf("Credit request preparation capability required")
	}
	buildctx := context.WithValue(ctx, preparationReservationKey{}, o.id)
	built, err := s.build(buildctx, request)
	if err != nil {
		return err
	}
	if built == nil {
		return fmt.Errorf("Credit draw preparation required")
	}
	intent, err := built.ExportIntent()
	if err != nil {
		return err
	}
	if intent.Draw == nil || intent.Context.ProtocolContext != previous.Intent.Context.ProtocolContext || intent.Context.TokenAddress != previous.Intent.Context.TokenAddress {
		return fmt.Errorf("Credit draw deployment or token changed")
	}
	if s.context != nil && intent.Context.ProtocolContext != *s.context {
		return fmt.Errorf("Credit intent deployment differs from managed wallet")
	}
	original := map[common.Hash]bool{}
	for _, input := range previous.Intent.Inputs {
		c, e := input.State.Commitment()
		if e != nil {
			return e
		}
		original[common.Hash(c.Bytes())] = true
	}
	if len(intent.Inputs) != len(previous.Intent.Inputs) || len(original) != len(intent.Inputs) {
		return fmt.Errorf("Credit draw input set changed")
	}
	for _, input := range intent.Inputs {
		c, e := input.State.Commitment()
		if e != nil {
			return e
		}
		hash := common.Hash(c.Bytes())
		if !original[hash] {
			return fmt.Errorf("Credit draw input set changed")
		}
		delete(original, hash)
		if s.owner != (common.Address{}) && common.BigToAddress(primitives.ElementToBigInt(input.State.SpenderWalletIdentity)) != s.owner {
			return fmt.Errorf("Credit spender differs from managed wallet")
		}
	}
	// Restore only the new statement; process-local authorization, proof and quote
	// capabilities from either preparation never survive the replacement.
	intent.Broadcaster = nil
	fresh, err := credit.Restore(ctx, intent, nil, nil)
	if err != nil {
		return err
	}
	var evidence *PreparationEvidence
	if s.prepare != nil {
		evidence, err = s.prepare(ctx, intent)
	} else if s.validate != nil {
		err = s.validate(ctx, intent)
	}
	if err != nil {
		return err
	}
	validation := ""
	if evidence != nil {
		validation = "required"
		if s.checkPreparation != nil {
			if err = s.checkPreparation(ctx, *evidence); err != nil {
				return err
			}
		}
	}
	worker, err := workerID()
	if err != nil {
		return err
	}
	now, err = s.chainTimestamp(ctx)
	if err != nil {
		return err
	}
	if now == 0 {
		return fmt.Errorf("current Credit chain timestamp unavailable")
	}
	if err = fresh.CheckTimestamp(now); err != nil {
		return err
	}
	if evidence != nil && s.checkPreparation != nil {
		if err = s.checkPreparation(ctx, *evidence); err != nil {
			return err
		}
	}
	err = s.store.Update(ctx, func(state *State) error {
		if state.Version != 2 {
			return fmt.Errorf("legacy Credit store is unsupported")
		}
		current, ok := state.Operations[o.id]
		if !ok || current.WorkerID != o.worker {
			return fmt.Errorf("Credit operation lease changed")
		}
		actual, e := json.Marshal(current)
		if e != nil {
			return e
		}
		defer clear(actual)
		if !bytes.Equal(actual, frozen) {
			return fmt.Errorf("Credit operation changed; retry re-preparation")
		}
		for _, input := range intent.Inputs {
			commitment, e := input.State.Commitment()
			if e != nil {
				return e
			}
			hash := common.Hash(commitment.Bytes())
			g, exists := state.Grants[GrantKey(hash, credit.Spender)]
			if !exists || g.ReservedBy != o.id || g.Consumed || g.GrantStatus != 1 {
				return fmt.Errorf("reserved Credit input unavailable")
			}
			if e = validateStored(g); e != nil {
				return e
			}
			if g.TreeID != input.TreeID || g.LeafIndex != input.LeafIndex {
				return fmt.Errorf("Credit location mismatch")
			}
			for _, other := range state.Grants {
				if other.Commitment == hash && (other.ReservedBy != o.id || other.Consumed || other.GrantStatus != 1) {
					return fmt.Errorf("Credit reservation changed")
				}
			}
		}
		state.Operations[o.id] = Journal{ID: o.id, WorkerID: worker, Status: "prepared", Intent: intent, SubmissionValidation: validation, Preparation: evidence}
		return nil
	})
	if err != nil {
		return err
	}
	o.prepared = fresh
	o.worker = worker
	return nil
}
