package session

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/credit"
	"github.com/invisi-lab/invisi-sdk-go/prover"
	"sync"
)

type Options struct {
	Validate     func(context.Context, credit.Intent) error
	Store        Store
	OwnerAddress common.Address
	Context      *core.ProtocolContext
}
type Session struct {
	validate func(context.Context, credit.Intent) error
	store    Store
	owner    common.Address
	context  *core.ProtocolContext
}

func New(options Options) (*Session, error) {
	if options.Store == nil {
		return nil, fmt.Errorf("Credit store required")
	}
	var pinned *core.ProtocolContext
	if options.Context != nil {
		copy := *options.Context
		pinned = &copy
	}
	return &Session{store: options.Store, owner: options.OwnerAddress, context: pinned, validate: options.Validate}, nil
}
func validateStored(s StoredState) error {
	if s.Role != credit.Owner && s.Role != credit.Spender || s.TreeID == 0 {
		return fmt.Errorf("invalid stored Credit role or tree")
	}
	if err := s.Delivery.Validate(); err != nil {
		return err
	}
	commitment, err := s.Delivery.State.Commitment()
	if err != nil {
		return err
	}
	nf, err := s.Delivery.State.Nullifier()
	if err != nil {
		return err
	}
	if common.Hash(commitment.Bytes()) != s.Commitment || common.Hash(nf.Bytes()) != s.Nullifier {
		return fmt.Errorf("stored Credit commitment or nullifier mismatch")
	}
	return nil
}

// Track imports a grant whose delivery and canonical membership the caller has authenticated.
// Preparation still checks the current consumed/revoked status through the composed wallet session.
func (s *Session) Track(ctx context.Context, record StoredState) error {
	identity := record.Delivery.State.OwnerWalletIdentity
	if record.Role == credit.Spender {
		identity = record.Delivery.State.SpenderWalletIdentity
	}
	if s.owner != (common.Address{}) && common.BigToAddress(core.ElementToBigInt(identity)) != s.owner {
		return fmt.Errorf("Credit role differs from managed wallet")
	}
	if s.context != nil && record.Delivery.State.ProtocolContext != *s.context {
		return fmt.Errorf("Credit deployment differs from managed wallet")
	}
	if err := validateStored(record); err != nil {
		return err
	}
	record.ReservedBy = ""
	return s.store.Update(ctx, func(state *State) error {
		if state.Version != 2 {
			return fmt.Errorf("legacy Credit store is unsupported")
		}
		if old, ok := state.Grants[record.key()]; ok {
			if old.Nullifier != record.Nullifier {
				return fmt.Errorf("conflicting stored Credit")
			}
			record.ReservedBy = old.ReservedBy
			record.Consumed = record.Consumed || old.Consumed
			record.Revoked = record.Revoked || old.Revoked
		}
		state.Grants[record.key()] = record
		return nil
	})
}
func workerID() (string, error) {
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return hex.EncodeToString(raw), nil
}

type Operation struct {
	mu         sync.Mutex
	session    *Session
	id, worker string
	prepared   *credit.Prepared
}

func (s *Session) Prepare(ctx context.Context, id string, prepared *credit.Prepared) (*Operation, error) {
	if id == "" || prepared == nil {
		return nil, fmt.Errorf("Credit operation ID and preparation required")
	}
	intent, err := prepared.ExportIntent()
	if err != nil {
		return nil, err
	}
	if s.context != nil && intent.Context.ProtocolContext != *s.context {
		return nil, fmt.Errorf("Credit intent deployment differs from managed wallet")
	}
	if s.owner != (common.Address{}) {
		switch {
		case intent.Create != nil:
			if intent.Create.Owner != s.owner {
				return nil, fmt.Errorf("Credit owner differs from managed wallet")
			}
		case intent.Update != nil:
			if intent.Update.Owner != s.owner {
				return nil, fmt.Errorf("Credit owner differs from managed wallet")
			}
		case intent.Revoke != nil:
			if intent.Revoke.Owner != s.owner {
				return nil, fmt.Errorf("Credit owner differs from managed wallet")
			}
		case intent.Draw != nil:
			for _, input := range intent.Inputs {
				if common.BigToAddress(core.ElementToBigInt(input.State.SpenderWalletIdentity)) != s.owner {
					return nil, fmt.Errorf("Credit spender differs from managed wallet")
				}
			}
		}
	}
	if s.validate != nil {
		if err := s.validate(ctx, intent); err != nil {
			return nil, err
		}
	}
	worker, err := workerID()
	if err != nil {
		return nil, err
	}
	err = s.store.Update(ctx, func(state *State) error {
		if state.Version != 2 {
			return fmt.Errorf("legacy Credit store is unsupported")
		}
		if _, exists := state.Operations[id]; exists {
			return fmt.Errorf("Credit operation ID already exists")
		}
		for _, input := range intent.Inputs {
			commitment, err := input.State.Commitment()
			if err != nil {
				return err
			}
			hash := common.Hash(commitment.Bytes())
			role := credit.Spender
			if intent.Update != nil {
				role = credit.Owner
			}
			key := StoredState{Commitment: hash, Role: role}.key()
			record, ok := state.Grants[key]
			if !ok || record.Consumed || record.Revoked || record.ReservedBy != "" {
				return fmt.Errorf("available Credit input not found")
			}
			if err = validateStored(record); err != nil {
				return err
			}
			if record.TreeID != input.TreeID || record.LeafIndex != input.LeafIndex {
				return fmt.Errorf("Credit location mismatch")
			}
			for k, other := range state.Grants {
				if other.Commitment != hash {
					continue
				}
				if other.ReservedBy != "" && other.ReservedBy != id {
					return fmt.Errorf("Credit already reserved")
				}
				other.ReservedBy = id
				state.Grants[k] = other
			}
		}
		state.Operations[id] = Journal{ID: id, WorkerID: worker, Status: "prepared", Intent: intent}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &Operation{session: s, id: id, worker: worker, prepared: prepared}, nil
}
func (o *Operation) check(ctx context.Context) error {
	state, err := o.session.store.ReadSnapshot(ctx)
	if err != nil {
		return err
	}
	record, ok := state.Operations[o.id]
	if !ok || record.WorkerID != o.worker {
		return fmt.Errorf("Credit operation lease changed")
	}
	if record.Status == "cancelled" || record.Status == "confirmed" || record.Status == "reverted" {
		return fmt.Errorf("Credit operation is terminal")
	}
	return nil
}
func (o *Operation) update(ctx context.Context, change func(*Journal) error) error {
	return o.session.store.Update(ctx, func(state *State) error {
		record, ok := state.Operations[o.id]
		if !ok || record.WorkerID != o.worker {
			return fmt.Errorf("Credit operation lease changed")
		}
		if err := change(&record); err != nil {
			return err
		}
		state.Operations[o.id] = record
		return nil
	})
}
func (o *Operation) SigningRequest(ctx context.Context) (apitypes.TypedData, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if err := o.checkFresh(ctx); err != nil {
		return apitypes.TypedData{}, err
	}
	return o.prepared.SigningRequest()
}
func (o *Operation) WithAuthority(ctx context.Context, auth credit.Authority) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	if err := o.checkFresh(ctx); err != nil {
		return err
	}
	return o.prepared.WithAuthority(auth)
}
func (o *Operation) Prove(ctx context.Context, backend prover.Prover, options prover.Options) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	if err := o.checkFresh(ctx); err != nil {
		return err
	}
	if err := o.prepared.Prove(ctx, backend, options); err != nil {
		return err
	}
	proof, err := o.prepared.Proof()
	if err != nil {
		return err
	}
	return o.update(ctx, func(record *Journal) error { record.Proof = proof; record.Status = "proved"; return nil })
}
func (o *Operation) Calldata(ctx context.Context, caller common.Address) ([]byte, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if err := o.checkFresh(ctx); err != nil {
		return nil, err
	}
	data, err := o.prepared.Calldata(caller)
	if err != nil {
		return nil, err
	}
	if err = o.update(ctx, func(record *Journal) error {
		record.Calldata = append([]byte(nil), data...)
		record.Status = "ready"
		return nil
	}); err != nil {
		return nil, err
	}
	return data, nil
}

// Externalize must commit before submitting calldata, including when no transaction hash is known yet.
func (o *Operation) Externalize(ctx context.Context, hash common.Hash) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	if err := o.check(ctx); err != nil {
		return err
	}
	return o.update(ctx, func(record *Journal) error {
		if len(record.Calldata) == 0 {
			return fmt.Errorf("frozen Credit calldata required before submission")
		}
		if record.TransactionHash != (common.Hash{}) && hash != (common.Hash{}) && record.TransactionHash != hash {
			return fmt.Errorf("conflicting Credit transaction hash")
		}
		record.Externalized = true
		record.Status = "uncertain"
		if hash != (common.Hash{}) {
			record.TransactionHash = hash
		}
		return nil
	})
}
func (o *Operation) Cancel(ctx context.Context) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	if err := o.check(ctx); err != nil {
		return err
	}
	return o.session.store.Update(ctx, func(state *State) error {
		record, ok := state.Operations[o.id]
		if !ok || record.WorkerID != o.worker {
			return fmt.Errorf("Credit operation lease changed")
		}
		if record.Externalized {
			return fmt.Errorf("externalized Credit operation requires canonical reconciliation")
		}
		record.Status = "cancelled"
		state.Operations[o.id] = record
		for k, grant := range state.Grants {
			if grant.ReservedBy == o.id {
				grant.ReservedBy = ""
				state.Grants[k] = grant
			}
		}
		return nil
	})
}

type AuthorizationProvider interface {
	OwnerAddress() common.Address
	Authorize(context.Context, core.WalletAuthorizationRequest) (core.WalletAuthorizationWitness, error)
}

func (o *Operation) Authorize(ctx context.Context, provider AuthorizationProvider) error {
	if provider == nil {
		return fmt.Errorf("Credit wallet authorization provider required")
	}
	o.mu.Lock()
	if err := o.checkFresh(ctx); err != nil {
		o.mu.Unlock()
		return err
	}
	request, err := o.prepared.AuthorizationRequest()
	o.mu.Unlock()
	if err != nil {
		return err
	}
	witness, err := provider.Authorize(ctx, request)
	if err != nil {
		return err
	}
	public := append([]byte{4}, witness.PublicKeyX[:]...)
	public = append(public, witness.PublicKeyY[:]...)
	return o.WithAuthority(ctx, credit.Authority{PublicKey: public, Signature: witness.Signature[:]})
}
func (o *Operation) WithRevocationSignature(ctx context.Context, signature []byte, verify func(context.Context, common.Address, apitypes.TypedData, []byte) (bool, error)) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	if err := o.checkFresh(ctx); err != nil {
		return err
	}
	return o.prepared.WithRevocationSignature(ctx, signature, verify)
}

func (o *Operation) checkFresh(ctx context.Context) error {
	if err := o.check(ctx); err != nil {
		return err
	}
	if o.session.validate == nil {
		return nil
	}
	state, err := o.session.store.ReadSnapshot(ctx)
	if err != nil {
		return err
	}
	record, ok := state.Operations[o.id]
	if !ok || record.WorkerID != o.worker {
		return fmt.Errorf("Credit operation lease changed")
	}
	return o.session.validate(ctx, record.Intent)
}
