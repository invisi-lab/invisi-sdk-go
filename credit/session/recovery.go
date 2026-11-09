package session

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/credit"
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
	if record.Status == "confirmed" || record.Status == "reverted" || record.Status == "cancelled" {
		return nil, fmt.Errorf("Credit operation is terminal")
	}
	prepared, err := credit.Restore(ctx, record.Intent, record.Proof, backend)
	if err != nil {
		return nil, err
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
		return fmt.Errorf("Credit confirmation is still uncertain")
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
					if common.BigToAddress(core.ElementToBigInt(s.OwnerWalletIdentity)) == p.Owner && common.BigToAddress(core.ElementToBigInt(s.TokenAddress)) == p.Token && core.ElementToBigInt(s.RevocationShard).Uint64() == p.Shard && core.ElementToBigInt(s.RevocationSlot).Uint64() == uint64(p.Slot) {
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
