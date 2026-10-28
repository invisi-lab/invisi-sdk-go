package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/prover"
)

func (o *Operation) validateRecovered() error {
	p := o.plan
	if p.Version != 1 || p.WalletIdentity != o.session.wallet.identity || p.Namespace != o.session.scope.Namespace {
		return errors.New("recovered wallet scope mismatch")
	}
	for _, input := range p.Inputs {
		note := input.Note
		if note.ProtocolContext != o.session.coreContext() || note.WalletIdentity != core.BigIntToElement(p.WalletIdentity.Big()) || note.TokenAddress != core.BigIntToElement(p.Context.TokenAddress.Big()) {
			return errors.New("recovered note scope mismatch")
		}
		identity, err := core.ComputeNoteIdentity(note.WalletIdentity, note.Nonce)
		if err != nil {
			return err
		}
		leaf, err := core.ComputeNoteCommitment(note.ChainID, identity, note.TokenAddress, note.Amount)
		if err != nil || fieldHash(leaf) != input.Commitment {
			return errors.New("recovered note commitment mismatch")
		}
		nf, err := o.session.wallet.ExportNullifier().ForNote(note)
		if err != nil || nf != input.Nullifier {
			return errors.New("recovered note nullifier mismatch")
		}
	}
	for _, input := range p.CreditInputs {
		note := input.State.Note
		actor := note.OwnerWalletIdentity
		if input.Role == "spender" {
			actor = note.SpenderWalletIdentity
		} else if input.Role != "owner" {
			return errors.New("invalid recovered Credit role")
		}
		if note.ProtocolContext != o.session.coreContext() || actor != core.BigIntToElement(p.WalletIdentity.Big()) || note.TokenAddress != core.BigIntToElement(p.Context.TokenAddress.Big()) {
			return errors.New("recovered Credit scope mismatch")
		}
		leaf, err := creditCommitment(note)
		if err != nil || leaf != input.Commitment {
			return errors.New("recovered Credit commitment mismatch")
		}
		nf, err := core.ComputeCreditNullifier(note.ChainID, note.CreditSharedKey, core.BigIntToElement(leaf.Big()))
		if err != nil || fieldHash(nf) != input.Nullifier {
			return errors.New("recovered Credit nullifier mismatch")
		}
	}
	if p.Kind != "shield" {
		words, err := o.words()
		if err != nil {
			return err
		}
		inputs := make([]common.Hash, len(words))
		for i := range words {
			inputs[i] = words[i]
		}
		statement, err := prover.Statement(o.circuitName(), p.Target, inputs)
		if err != nil {
			return err
		}
		if o.bundle != nil {
			if err = prover.AssertProofBundle(*o.bundle, statement); err != nil {
				return err
			}
		}
	}
	if len(o.data) > 0 {
		if p.Kind != "shield" && o.bundle == nil {
			return errors.New("recovered calldata has no proof")
		}
		stored := append([]byte(nil), o.data...)
		if err := o.encode(); err != nil {
			return err
		}
		if !bytes.Equal(stored, o.data) {
			return errors.New("recovered transaction differs from frozen operation")
		}
	}
	return nil
}
func (s *WalletSession) Resume(ctx context.Context, id string) (*Operation, error) {
	worker, err := randomOperationID()
	if err != nil {
		return nil, err
	}
	var operation *Operation
	err = s.options.Store.Update(ctx, s.scope, func(state *WalletState) error {
		record, ok := state.Operations[id]
		if !ok {
			return errors.New("unknown operation")
		}
		var plan operationPlan
		if err := json.Unmarshal(record.Plan, &plan); err != nil {
			return err
		}
		if plan.ID != id || plan.Context.ChainID == nil || !plan.Context.ChainID.IsUint64() || plan.Context.ChainID.Uint64() != s.scope.ChainID || plan.Context.ProtocolAddress != s.scope.ProtocolAddress {
			return errors.New("saved operation deployment mismatch")
		}
		operation = &Operation{session: s, plan: plan, status: record.Status, data: append([]byte(nil), record.CallData...), transactionHash: record.TransactionHash, workerID: worker}
		if len(record.Proof) > 0 {
			if err := json.Unmarshal(record.Proof, &operation.bundle); err != nil {
				return err
			}
		}
		if err := operation.validateRecovered(); err != nil {
			return err
		}
		if operation.status == "authorized" {
			operation.status = "prepared"
			record.Status = "prepared"
		}
		record.WorkerID = worker
		state.Operations[id] = record
		return nil
	})
	return operation, err
}
