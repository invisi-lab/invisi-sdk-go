package session

import (
	"bytes"
	"context"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/invisi-lab/invisi-sdk-go/credit"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
	"github.com/invisi-lab/invisi-sdk-go/rpc/contracts"
	creditsession "github.com/invisi-lab/invisi-sdk-go/session/credit"
)

type creditReconciler struct {
	session  *WalletSession
	finality protocolrpc.FinalityTag
}

// CreditReconciler authenticates exact transaction calldata and canonical receipts before grant journals settle.
func (s *WalletSession) CreditReconciler(finality protocolrpc.FinalityTag) creditsession.Reconciler {
	if finality == "" {
		finality = protocolrpc.FinalityLatest
	}
	return creditReconciler{s, finality}
}
func (r creditReconciler) Reconcile(ctx context.Context, intent credit.Intent, data []byte, hash common.Hash) (*creditsession.Outcome, error) {
	s := r.session
	if hash == (common.Hash{}) {
		return nil, nil
	}
	if s == nil {
		return nil, errors.New("Credit reconciliation requires the wallet viewing capability")
	}
	if primitives.ElementToBigInt(intent.Context.ChainID).Uint64() != s.scope.ChainID || common.BigToAddress(primitives.ElementToBigInt(intent.Context.ProtocolAddress)) != s.scope.ProtocolAddress {
		return nil, errors.New("Credit intent deployment mismatch")
	}
	if err := s.protocol.ValidateDeployment(ctx, s.options.Provider); err != nil {
		return nil, err
	}
	receipt, err := s.options.Provider.TransactionReceipt(ctx, hash)
	if err != nil {
		return nil, err
	}
	if receipt == nil {
		return nil, nil
	}
	if receipt.TxHash != hash || receipt.BlockNumber == nil || !receipt.BlockNumber.IsUint64() || receipt.Status > 1 {
		return nil, errors.New("invalid Credit receipt")
	}
	boundary, err := protocolrpc.RequireFinality(ctx, s.options.Provider, r.finality)
	if err != nil {
		return nil, err
	}
	if boundary.Number < receipt.BlockNumber.Uint64() {
		return nil, nil
	}

	header, err := s.options.Provider.HeaderByNumber(ctx, receipt.BlockNumber)
	if err != nil {
		return nil, err
	}
	if header == nil || header.Number == nil || header.Number.Cmp(receipt.BlockNumber) != 0 || header.Hash() != receipt.BlockHash {
		return nil, errors.New("Credit receipt is not canonical")
	}
	reader, ok := s.options.Provider.(interface {
		TransactionByHash(context.Context, common.Hash) (*types.Transaction, bool, error)
	})
	if !ok {
		return nil, errors.New("Credit reconciliation requires transaction lookup")
	}
	tx, pending, err := reader.TransactionByHash(ctx, hash)
	if err != nil {
		return nil, err
	}
	if tx == nil || pending || tx.Hash() != hash || tx.To() == nil || *tx.To() != s.scope.ProtocolAddress || tx.ChainId().Cmp(new(big.Int).SetUint64(s.scope.ChainID)) != 0 || !bytes.Equal(tx.Data(), data) {
		return nil, errors.New("Credit receipt transaction differs from frozen calldata")
	}
	outcome := &creditsession.Outcome{TransactionHash: hash, Calldata: append([]byte(nil), data...), BlockHash: receipt.BlockHash, BlockNumber: receipt.BlockNumber.Uint64(), Success: receipt.Status == types.ReceiptStatusSuccessful}
	if outcome.Success {
		anchor := boundary
		notes, states, draws, err := s.readTransaction(ctx, hash, anchor, s.storedWalletResolver(ctx))
		if err != nil {
			return nil, err
		}
		outcome.Successors = states
		if intent.Draw != nil {
			snapshot, err := s.options.Store.ReadSnapshot(ctx, s.scope)
			if err != nil {
				return nil, err
			}
			seen := map[string]bool{}
			for _, state := range states {
				seen[creditsession.GrantKey(state.Commitment, state.Role)] = true
			}
			for _, draw := range draws {
				for i, nf := range draw.Params.Nullifiers {
					for _, input := range intent.Inputs {
						expected, err := input.State.Nullifier()
						if err != nil {
							return nil, err
						}
						if common.Hash(expected.Bytes()) != common.Hash(nf) {
							continue
						}
						if i >= len(draw.Params.CreditCommitments) || i >= len(draw.Params.RecoveryAmounts) || i >= len(draw.LeafIndices) {
							return nil, errors.New("Credit draw event vector mismatch")
						}
						next, err := input.State.Recover(primitives.BigIntToElement(common.Hash(nf).Big()), primitives.BigIntToElement(common.Hash(draw.Params.CreditCommitments[i]).Big()), primitives.BigIntToElement(common.Hash(draw.Params.RecoveryAmounts[i]).Big()), draw.Timestamp)
						if err != nil {
							return nil, err
						}
						previousCommitment, err := input.State.Commitment()
						if err != nil {
							return nil, err
						}
						for _, old := range snapshot.PostpaidCredit.Grants {
							if old.Commitment != common.Hash(previousCommitment.Bytes()) {
								continue
							}
							commitment := common.Hash(draw.Params.CreditCommitments[i])
							key := creditsession.GrantKey(commitment, old.Role)
							if seen[key] {
								continue
							}
							old.Delivery.State = next
							old.Commitment = commitment
							nullifier, err := next.Nullifier()
							if err != nil {
								return nil, err
							}
							old.Nullifier = common.Hash(nullifier.Bytes())
							old.TreeID = draw.OutputTreeId.Uint64()
							old.LeafIndex = uint16(draw.LeafIndices[i])
							old.Consumed = false
							old.ReservedBy = ""
							outcome.Successors = append(outcome.Successors, old)
							seen[key] = true
						}
					}
				}
			}
		}
		if intent.Revoke != nil {
			filter, err := contracts.NewInvisiProtocolFilterer(s.scope.ProtocolAddress, nil)
			if err != nil {
				return nil, err
			}
			found := false
			for _, log := range receipt.Logs {
				if log == nil || log.Address != s.scope.ProtocolAddress {
					continue
				}
				event, e := filter.ParseCreditRevoked(*log)
				if e == nil && event.Owner == intent.Revoke.Owner && event.Token == intent.Revoke.Token && event.Shard == intent.Revoke.Shard && event.Slot == intent.Revoke.Slot {
					outcome.RevokedOwner = event.Owner
					outcome.RevokedToken = event.Token
					outcome.RevokedShard = event.Shard
					outcome.RevokedSlot = event.Slot
					found = true
				}
			}
			if !found {
				return nil, errors.New("canonical Credit revocation event missing")
			}
		}
		if len(notes) > 0 {
			if err = s.options.Store.Update(ctx, s.scope, func(state *WalletState) error {
				for _, note := range notes {
					if old, ok := state.Notes[note.Commitment]; ok {
						note.Spent = old.Spent
						note.ReservedBy = old.ReservedBy
					}
					state.Notes[note.Commitment] = note
				}
				return nil
			}); err != nil {
				return nil, err
			}
		}
	}
	final, err := s.options.Provider.HeaderByNumber(ctx, receipt.BlockNumber)
	if err != nil {
		return nil, err
	}
	if final == nil || final.Hash() != receipt.BlockHash {
		return nil, errors.New("Credit receipt changed during reconciliation")
	}
	return outcome, nil
}
