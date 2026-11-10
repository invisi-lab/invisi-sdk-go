package session

import (
	"context"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	transaction "github.com/invisi-lab/invisi-sdk-go/internal/transaction"
	"github.com/invisi-lab/invisi-sdk-go/rpc/contracts"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
)

// reconcileSpendState resolves dropped/reverted transactions without assuming that
// a reverted receipt revokes the reusable proof. Only expiry or spent inputs do.
func (o *Operation) reconcileSpendState(ctx context.Context) (transaction.Confirmation, error) {
	result := transaction.Confirmation{Status: o.status, TransactionHash: o.transactionHash}
	header, err := o.session.options.Provider.HeaderByNumber(ctx, nil)
	if err != nil {
		return result, err
	}
	if header == nil || header.Number == nil || !header.Number.IsUint64() {
		return result, errors.New("latest block unavailable")
	}
	block := protocolrpc.CanonicalBlock{Number: header.Number.Uint64(), Hash: header.Hash(), ParentHash: header.ParentHash, Timestamp: header.Time}
	spent, err := o.spentAt(ctx, header.Number)
	if err != nil {
		return result, err
	}
	expired := transaction.OperationExpired(o.plan, header.Time)
	impossible := false
	for _, value := range spent {
		impossible = impossible || value
	}
	status := o.status
	if expired {
		status = "expired"
	} else if impossible {
		status = "failed"
	} else {
		status = "uncertain"
	}
	if err = o.session.canonical(ctx, block); err != nil {
		return result, err
	}
	err = o.session.options.Store.Update(ctx, o.session.scope, func(state *WalletState) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		record, ok := state.Operations[o.plan.ID]
		if !ok || state.Invalidated || record.WorkerID != o.workerID {
			return errors.New("stale operation worker during reconciliation")
		}
		if err := checkObservation(state, block); err != nil {
			return err
		}
		if expired || impossible {
			for _, input := range o.plan.Inputs {
				if note, ok := state.Notes[input.Commitment]; ok {
					note.Spent = note.Spent || spent[input.Commitment]
					if note.ReservedBy == o.plan.ID {
						note.ReservedBy = ""
					}
					state.Notes[input.Commitment] = note
				}
			}

			record.ResolvedAt = &block
			state.CheckedThrough = &block
		} else if !record.Externalized && o.transactionHash == (common.Hash{}) {
			status = record.Status
		}
		record.Status = status
		record.TransactionHash = o.transactionHash
		state.Operations[o.plan.ID] = record
		return nil
	})
	if err != nil {
		return result, err
	}
	o.status = status
	result.Status = status
	return result, nil
}

func (o *Operation) spentAt(ctx context.Context, height *big.Int) (map[common.Hash]bool, error) {
	caller, err := contracts.NewInvisiProtocolCaller(o.plan.Context.ProtocolAddress, o.session.options.Provider)
	if err != nil {
		return nil, err
	}
	result := map[common.Hash]bool{}
	options := &bind.CallOpts{Context: ctx, BlockNumber: height}
	for _, input := range o.plan.Inputs {
		value, err := caller.IsNullifierSpent(options, input.Nullifier)
		if err != nil {
			return nil, err
		}
		result[input.Commitment] = value
	}

	return result, nil
}

func checkObservation(state *WalletState, block protocolrpc.CanonicalBlock) error {
	for _, point := range []*protocolrpc.CanonicalBlock{state.CheckedThrough, state.ObservedThrough} {
		if point != nil && (point.Number > block.Number || (point.Number == block.Number && point.Hash != block.Hash)) {
			return errors.New("wallet advanced or changed during reconciliation")
		}
	}
	return nil
}
