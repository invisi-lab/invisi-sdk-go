package session

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	transaction "github.com/invisi-lab/invisi-sdk-go/internal/transaction"

	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
)

func (o *Operation) persist(ctx context.Context, status string) error {

	encoded, e := json.Marshal(o.plan)
	if e != nil {
		return e
	}
	return o.session.options.Store.Update(ctx, o.session.scope, func(state *WalletState) error {
		record, ok := state.Operations[o.plan.ID]
		if !ok {
			return errors.New("operation record missing")
		}
		if record.WorkerID != o.workerID {
			return errors.New("operation worker was replaced; resume operation")
		}
		record.Status = status
		if status == "uncertain" {
			record.Externalized = true
		}
		record.TransactionHash = o.transactionHash
		record.CallData = append([]byte(nil), o.data...)
		record.Plan = encoded
		if o.bundle != nil {
			raw, e := json.Marshal(o.bundle)
			if e != nil {
				return e
			}
			record.Proof = raw
		}
		state.Operations[o.plan.ID] = record
		return nil
	})
}

type ReconcileOptions struct{ TransactionHash common.Hash }

func (o *Operation) reconcileInternal(ctx context.Context, options ...ReconcileOptions) (transaction.Confirmation, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if e := o.checkLease(ctx); e != nil {
		return transaction.Confirmation{}, e
	}
	if len(options) > 1 {
		return transaction.Confirmation{}, errors.New("expected at most one reconciliation option")
	}
	if len(options) == 1 && options[0].TransactionHash != (common.Hash{}) {
		o.transactionHash = options[0].TransactionHash
	}

	if e := o.session.protocol.ValidateDeployment(ctx, o.session.options.Provider); e != nil {
		return transaction.Confirmation{}, e
	}

	result := transaction.Confirmation{Status: o.status, TransactionHash: o.transactionHash}
	if o.transactionHash == (common.Hash{}) {
		return o.reconcileSpendState(ctx)
	}
	receipt, e := o.session.options.Provider.TransactionReceipt(ctx, o.transactionHash)
	if errors.Is(e, ethereum.NotFound) {
		return o.reconcileSpendState(ctx)
	}
	if e != nil {
		return result, e
	}
	if receipt == nil {
		return o.reconcileSpendState(ctx)
	}
	if receipt.TxHash != o.transactionHash || receipt.BlockNumber == nil || !receipt.BlockNumber.IsUint64() {
		return result, errors.New("receipt does not identify the submitted transaction and block")
	}
	header, e := o.session.options.Provider.HeaderByNumber(ctx, receipt.BlockNumber)
	if e != nil {
		return result, e
	}
	if header == nil || header.Number == nil || header.Number.Cmp(receipt.BlockNumber) != 0 || header.Hash() != receipt.BlockHash {
		return result, errors.New("transaction receipt is not canonical")
	}
	reader, ok := o.session.options.Provider.(interface {
		TransactionByHash(context.Context, common.Hash) (*types.Transaction, bool, error)
	})
	if !ok {
		return result, errors.New("provider must authenticate submitted transaction calldata")
	}
	tx, pending, e := reader.TransactionByHash(ctx, o.transactionHash)
	if e != nil {
		return result, e
	}
	if pending || tx == nil || tx.Hash() != o.transactionHash || tx.To() == nil || *tx.To() != o.plan.Context.ProtocolAddress || !o.matchesTransactionData(tx.Data()) {
		return result, errors.New("submitted transaction differs from frozen operation")
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		return o.reconcileSpendState(ctx)
	}
	receiptBlock := protocolrpc.CanonicalBlock{Number: header.Number.Uint64(), Hash: header.Hash(), ParentHash: header.ParentHash, Timestamp: header.Time}
	latest, e := o.session.options.Provider.HeaderByNumber(ctx, nil)
	if e != nil {
		return result, e
	}
	if latest == nil || latest.Number == nil || !latest.Number.IsUint64() || latest.Number.Cmp(header.Number) < 0 {
		return result, errors.New("latest block predates receipt")
	}
	block := protocolrpc.CanonicalBlock{Number: latest.Number.Uint64(), Hash: latest.Hash(), ParentHash: latest.ParentHash, Timestamp: latest.Time}
	spent, e := o.spentAt(ctx, latest.Number)
	if e != nil {
		return result, e
	}
	for _, value := range spent {
		if !value {
			return result, errors.New("successful transaction has unspent input")
		}
	}
	if e = o.session.canonical(ctx, receiptBlock); e != nil {
		return result, e
	}
	if e = o.session.canonical(ctx, block); e != nil {
		return result, e
	}
	e = o.session.options.Store.Update(ctx, o.session.scope, func(state *WalletState) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		record, ok := state.Operations[o.plan.ID]
		if !ok || state.Invalidated || record.WorkerID != o.workerID {
			return errors.New("stale operation worker during confirmation")
		}
		if e := checkObservation(state, block); e != nil {
			return e
		}
		for _, input := range o.plan.Inputs {
			if note, ok := state.Notes[input.Commitment]; ok {
				if note.ReservedBy != "" && note.ReservedBy != o.plan.ID {
					return errors.New("input reserved by another operation")
				}
				note.Spent = true
				note.ReservedBy = ""
				state.Notes[input.Commitment] = note
			}
		}
		for _, input := range o.plan.CreditInputs {
			for key, credit := range state.Credits {
				if credit.Commitment == input.Commitment {
					if credit.ReservedBy != "" && credit.ReservedBy != o.plan.ID {
						return errors.New("Credit reserved by another operation")
					}
					credit.Spent = true
					credit.ReservedBy = ""
					state.Credits[key] = credit
				}
			}
		}
		record.Status = "confirmed"
		record.TransactionHash = o.transactionHash
		record.ResolvedAt = &block
		state.Operations[o.plan.ID] = record
		state.CheckedThrough = &block
		return nil
	})
	if e != nil {
		return result, e
	}
	o.status = "confirmed"
	result.Status = "confirmed"
	return result, nil
}

func (o *Operation) Abandon(ctx context.Context) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	if e := o.checkLease(ctx); e != nil {
		return e
	}
	if o.status != "prepared" {
		return errors.New("only an unexternalized prepared operation can be abandoned")
	}
	e := o.session.options.Store.Update(ctx, o.session.scope, func(state *WalletState) error {
		if record, ok := state.Operations[o.plan.ID]; !ok || record.WorkerID != o.workerID {
			return errors.New("operation worker replaced")
		}
		for hash, n := range state.Notes {
			if n.ReservedBy == o.plan.ID {
				n.ReservedBy = ""
				state.Notes[hash] = n
			}
		}
		for key, c := range state.Credits {
			if c.ReservedBy == o.plan.ID {
				c.ReservedBy = ""
				state.Credits[key] = c
			}
		}
		record := state.Operations[o.plan.ID]
		record.Status = "failed"
		state.Operations[o.plan.ID] = record
		return nil
	})
	if e == nil {
		o.status = "failed"
	}
	return e
}

func (o *Operation) checkLease(ctx context.Context) error {

	state, e := o.session.options.Store.ReadSnapshot(ctx, o.session.scope)
	if e != nil {
		return e
	}
	if state.Invalidated {
		return errors.New("wallet requires reset refresh")
	}
	record, ok := state.Operations[o.plan.ID]
	if !ok || record.WorkerID != o.workerID {
		return errors.New("operation worker was replaced; resume operation")
	}
	return nil
}
