package session

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/invisi-lab/invisi-sdk-go/rpc/broadcaster"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/invisi-lab/invisi-sdk-go/credit"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
	"github.com/invisi-lab/invisi-sdk-go/prover"
)

type Options struct {
	WrapPreparation     func(string, error) error
	Build               func(context.Context, any) (*credit.Prepared, error)
	CheckPreparation    func(context.Context, PreparationEvidence) error
	ValidatePreparation func(context.Context, PreparationEvidence) error
	Prepare             func(context.Context, credit.Intent) (*PreparationEvidence, error)
	VerifyRevocation    func(context.Context, common.Address, apitypes.TypedData, []byte) (bool, error)
	Validate            func(context.Context, credit.Intent) error
	CheckCancellation   func(context.Context, credit.Intent) error
	Store               Store
	OwnerAddress        common.Address
	Context             *primitives.ProtocolContext
}
type Session struct {
	wrapPreparation     func(string, error) error
	build               func(context.Context, any) (*credit.Prepared, error)
	checkPreparation    func(context.Context, PreparationEvidence) error
	validatePreparation func(context.Context, PreparationEvidence) error
	prepare             func(context.Context, credit.Intent) (*PreparationEvidence, error)
	verifyRevocation    func(context.Context, common.Address, apitypes.TypedData, []byte) (bool, error)
	validate            func(context.Context, credit.Intent) error
	checkCancellation   func(context.Context, credit.Intent) error
	store               Store
	owner               common.Address
	context             *primitives.ProtocolContext
}

func New(options Options) (*Session, error) {
	if options.Store == nil {
		return nil, fmt.Errorf("Credit store required")
	}
	var pinned *primitives.ProtocolContext
	if options.Context != nil {
		copy := *options.Context
		pinned = &copy
	}
	return &Session{checkCancellation: options.CheckCancellation, wrapPreparation: options.WrapPreparation, build: options.Build, checkPreparation: options.CheckPreparation, validatePreparation: options.ValidatePreparation, store: options.Store, owner: options.OwnerAddress, context: pinned, validate: options.Validate, prepare: options.Prepare, verifyRevocation: options.VerifyRevocation}, nil
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
	if s.owner != (common.Address{}) && common.BigToAddress(primitives.ElementToBigInt(identity)) != s.owner {
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
	broadcaster.OperationTarget `json:"-"`
	mu                          sync.Mutex
	session                     *Session
	id, worker                  string
	prepared                    *credit.Prepared
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
				if common.BigToAddress(primitives.ElementToBigInt(input.State.SpenderWalletIdentity)) != s.owner {
					return nil, fmt.Errorf("Credit spender differs from managed wallet")
				}
			}
		}
	}
	var evidence *PreparationEvidence
	if s.prepare != nil {
		evidence, err = s.prepare(ctx, intent)
	} else if s.validate != nil {
		err = s.validate(ctx, intent)
	}
	if err != nil {
		return nil, err
	}
	validation := ""
	if evidence != nil {
		validation = "required"
	}
	worker, err := workerID()
	if err != nil {
		return nil, err
	}
	if evidence != nil && s.checkPreparation != nil {
		if err = s.checkPreparation(ctx, *evidence); err != nil {
			return nil, err
		}
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
		state.Operations[id] = Journal{ID: id, WorkerID: worker, Status: "prepared", Intent: intent, SubmissionValidation: validation, Preparation: evidence}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if evidence != nil && s.checkPreparation != nil {
		if err = s.checkPreparation(ctx, *evidence); err != nil {
			return nil, err
		}
	}
	return (&Operation{session: s, id: id, worker: worker, prepared: prepared}).bindBroadcasterQuote(), nil
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
	if record.Status == "failed" || record.Status == "confirmed" || record.Status == "reverted" || record.Status == "cancelled" {
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
	if err := o.validateSubmission(ctx); err != nil {
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
func (o *Operation) Abandon(ctx context.Context) error {
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
		if record.Status != "prepared" && record.Status != "proved" && record.Status != "ready" {
			return fmt.Errorf("Credit operation is not locally abandonable")
		}
		if record.Externalized || len(record.Calldata) > 0 || record.TransactionHash != (common.Hash{}) {
			return fmt.Errorf("externalized Credit operation requires canonical reconciliation")
		}
		record.Status = "failed"
		record.WorkerID = ""
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

func (o *Operation) WithRevocationSignature(ctx context.Context, signature []byte, verify func(context.Context, common.Address, apitypes.TypedData, []byte) (bool, error)) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	if err := o.checkFresh(ctx); err != nil {
		return err
	}
	state, err := o.session.store.ReadSnapshot(ctx)
	if err != nil {
		return err
	}
	if state.Operations[o.id].Externalized {
		return fmt.Errorf("externalized Credit revocation signature is frozen")
	}
	if err := o.prepared.WithRevocationSignature(ctx, signature, verify); err != nil {
		return err
	}
	return o.update(ctx, func(r *Journal) error {
		if r.Externalized {
			return fmt.Errorf("externalized Credit revocation signature is frozen")
		}
		r.RevocationSignature = append([]byte(nil), signature...)
		r.Calldata = nil
		r.Status = "prepared"
		return nil
	})
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
	if record.SubmissionValidation == "required" {
		if record.Preparation == nil {
			return fmt.Errorf("cached Credit preparation evidence missing")
		}
		if record.Preparation != nil && o.session.checkPreparation != nil {
			if err = o.session.checkPreparation(ctx, *record.Preparation); err != nil {
				return err
			}
		}
		now := uint64(time.Now().Unix())
		if record.Intent.Context.Timestamp > now {
			now = record.Intent.Context.Timestamp
		}
		return o.prepared.CheckTimestamp(now)
	}
	return o.session.validate(ctx, record.Intent)
}

// ValidateSubmission checks current chain state before any transaction is externalized.
func (o *Operation) ValidateSubmission(ctx context.Context) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.validateSubmission(ctx)
}
func (o *Operation) validateSubmission(ctx context.Context) error {
	if err := o.check(ctx); err != nil {
		return err
	}
	state, err := o.session.store.ReadSnapshot(ctx)
	if err != nil {
		return err
	}
	record := state.Operations[o.id]
	if record.SubmissionValidation == "required" && record.Preparation == nil {
		return fmt.Errorf("cached Credit preparation evidence missing")
	}
	if record.Preparation != nil && o.session.validatePreparation != nil {
		if err = o.session.validatePreparation(ctx, *record.Preparation); err != nil {
			return err
		}
	}
	if o.session.validate != nil {
		if err = o.session.validate(ctx, record.Intent); err != nil {
			return err
		}
	}
	if record.Intent.Revoke != nil && len(record.RevocationSignature) > 0 && o.session.verifyRevocation != nil {
		typed, e := o.prepared.SigningRequest()
		if e != nil {
			return e
		}
		valid, e := o.session.verifyRevocation(ctx, record.Intent.Revoke.Owner, typed, record.RevocationSignature)
		if e != nil {
			return e
		}
		if !valid {
			return fmt.Errorf("invalid current Credit revocation signature")
		}
	}
	if record.Preparation != nil && o.session.validatePreparation != nil {
		return o.session.validatePreparation(ctx, *record.Preparation)
	}
	return nil
}
