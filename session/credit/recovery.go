package session

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/invisi-lab/invisi-sdk-go/credit"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
	"github.com/invisi-lab/invisi-sdk-go/prover"
)

func (s *Session) Resume(ctx context.Context, id string, backend prover.Prover) (*Operation, error) {
	snapshot, err := s.store.ReadSnapshot(ctx)
	if err != nil {
		return nil, err
	}
	record, ok := snapshot.Operations[id]
	if !ok {
		return nil, fmt.Errorf("Credit operation not found")
	}
	if record.Status == "confirmed" || record.Status == "reverted" || record.Status == "failed" {
		return nil, fmt.Errorf("Credit operation is terminal")
	}
	prepared, err := credit.Restore(ctx, record.Intent, record.Proof, backend)
	if err != nil {
		return nil, err
	}
	// EOA authorization is rechecked locally; contract-wallet recovery uses its live verifier.
	if len(record.RevocationSignature) > 0 {
		if err = prepared.WithRevocationSignature(ctx, record.RevocationSignature, func(ctx context.Context, owner common.Address, typed apitypes.TypedData, signature []byte) (bool, error) {
			digest, _, e := apitypes.TypedDataAndHash(typed)
			if e != nil {
				return false, e
			}
			sig := append([]byte(nil), signature...)
			if len(sig) == 64 {
				expanded := make([]byte, 65)
				copy(expanded, sig)
				expanded[64] = expanded[32] >> 7
				expanded[32] &= 0x7f
				sig = expanded
			}
			if len(sig) == 65 {
				if sig[64] >= 27 {
					sig[64] -= 27
				}
				public, e := crypto.SigToPub(digest, sig)
				if e == nil && crypto.PubkeyToAddress(*public) == owner {
					return true, nil
				}
			}
			if s.verifyRevocation == nil {
				return false, fmt.Errorf("retained contract-wallet revocation requires a verifier")
			}
			return s.verifyRevocation(ctx, owner, typed, signature)
		}); err != nil {
			return nil, err
		}
	}
	worker, err := workerID()
	if err != nil {
		return nil, err
	}
	frozen, err := json.Marshal(record)
	if err != nil {
		return nil, err
	}
	defer clear(frozen)
	err = s.store.Update(ctx, func(state *State) error {
		current, exists := state.Operations[id]
		if !exists {
			return fmt.Errorf("Credit operation disappeared")
		}
		actual, err := json.Marshal(current)
		if err != nil {
			return err
		}
		defer clear(actual)
		if !bytes.Equal(actual, frozen) {
			return fmt.Errorf("Credit operation changed; retry resume")
		}
		current.WorkerID = worker
		state.Operations[id] = current
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &Operation{session: s, id: id, worker: worker, prepared: prepared}, nil
}

type Outcome struct {
	TransactionHash            common.Hash
	Calldata                   []byte
	BlockHash                  common.Hash
	BlockNumber                uint64
	Success                    bool
	Successors                 []StoredState
	RevokedOwner, RevokedToken common.Address
	RevokedShard               uint64
	RevokedSlot                uint16
}

// Reconciler authenticates canonical chain receipts, exact calldata and successor event positions.
// Protocol RPC candidates alone cannot implement this interface safely.
type Reconciler interface {
	Reconcile(context.Context, credit.Intent, []byte, common.Hash) (*Outcome, error)
}

func (o *Operation) Reconcile(ctx context.Context, reconciler Reconciler) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	if reconciler == nil {
		return fmt.Errorf("canonical Credit reconciler required")
	}
	if err := o.check(ctx); err != nil {
		return err
	}
	snapshot, err := o.session.store.ReadSnapshot(ctx)
	if err != nil {
		return err
	}
	record := snapshot.Operations[o.id]
	if !record.Externalized || len(record.Calldata) == 0 {
		return fmt.Errorf("Credit operation was not externalized")
	}
	intent, err := clone(record.Intent)
	if err != nil {
		return err
	}
	outcome, err := reconciler.Reconcile(ctx, intent, append([]byte(nil), record.Calldata...), record.TransactionHash)
	if err != nil {
		return err
	}
	if outcome == nil {
		return ErrConfirmationPending
	}
	if outcome.BlockHash == (common.Hash{}) || outcome.TransactionHash == (common.Hash{}) || record.TransactionHash != (common.Hash{}) && outcome.TransactionHash != record.TransactionHash || !bytes.Equal(outcome.Calldata, record.Calldata) {
		return fmt.Errorf("Credit receipt does not match journal")
	}
	for _, successor := range outcome.Successors {
		if err := validateStored(successor); err != nil {
			return err
		}
	}
	return o.session.store.Update(ctx, func(state *State) error {
		current, ok := state.Operations[o.id]
		if !ok || current.WorkerID != o.worker {
			return fmt.Errorf("Credit operation lease changed")
		}
		if current.TransactionHash != record.TransactionHash || !bytes.Equal(current.Calldata, record.Calldata) {
			return fmt.Errorf("Credit journal changed during reconciliation")
		}
		for k, grant := range state.Grants {
			if grant.ReservedBy != o.id {
				continue
			}
			grant.ReservedBy = ""
			if outcome.Success {
				grant.Consumed = true
			}
			state.Grants[k] = grant
		}
		if outcome.Success {
			for _, successor := range outcome.Successors {
				successor.ReservedBy = ""
				state.Grants[successor.key()] = successor
			}
			if intent.Revoke != nil {
				p := intent.Revoke
				if outcome.RevokedOwner != p.Owner || outcome.RevokedToken != p.Token || outcome.RevokedShard != p.Shard || outcome.RevokedSlot != p.Slot {
					return fmt.Errorf("Credit revocation event namespace mismatch")
				}
				for k, grant := range state.Grants {
					s := grant.Delivery.State
					if common.BigToAddress(primitives.ElementToBigInt(s.OwnerWalletIdentity)) == p.Owner && common.BigToAddress(primitives.ElementToBigInt(s.TokenAddress)) == p.Token && primitives.ElementToBigInt(s.RevocationShard).Uint64() == p.Shard && primitives.ElementToBigInt(s.RevocationSlot).Uint64() == uint64(p.Slot) {
						grant.Revoked = true
						state.Grants[k] = grant
					}
				}
			}
		}
		current.TransactionHash = outcome.TransactionHash
		if outcome.Success {
			current.Status = "confirmed"
		} else {
			current.Status = "reverted"
		}
		state.Operations[o.id] = current
		return nil
	})
}

// ReconcilePending checks externalized journals without replacing their worker leases.
func (s *Session) ReconcilePending(ctx context.Context, reconciler Reconciler) error {
	snapshot, e := s.store.ReadSnapshot(ctx)
	if e != nil {
		return e
	}
	for id, r := range snapshot.Operations {
		if !r.Externalized || len(r.Calldata) == 0 || r.TransactionHash == (common.Hash{}) || r.Status == "confirmed" || r.Status == "failed" || r.Status == "reverted" {
			continue
		}
		o := &Operation{session: s, id: id, worker: r.WorkerID}
		if e = o.Reconcile(ctx, reconciler); e != nil && !errors.Is(e, ErrConfirmationPending) {
			return e
		}
	}
	return nil
}

var ErrConfirmationPending = errors.New("Credit confirmation is still uncertain")
