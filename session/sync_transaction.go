package session

import (
	"context"
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/internal/receipts"
	transaction "github.com/invisi-lab/invisi-sdk-go/internal/transaction"
	"github.com/invisi-lab/invisi-sdk-go/protocol/contracts"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
)

type authenticatedCreditDraw struct {
	*contracts.InvisiProtocolCreditDraw
	Timestamp uint64
}

func (s *WalletSession) readTransaction(ctx context.Context, hash common.Hash, anchor protocolrpc.CanonicalBlock) ([]StoredNote, []StoredCredit, []*authenticatedCreditDraw, error) {
	receipt, e := s.options.Provider.TransactionReceipt(ctx, hash)
	if e != nil {
		return nil, nil, nil, e
	}
	if receipt == nil || receipt.Status != types.ReceiptStatusSuccessful || receipt.TxHash != hash || receipt.BlockNumber == nil || !receipt.BlockNumber.IsUint64() || receipt.BlockNumber.Uint64() > anchor.Number {
		return nil, nil, nil, errors.New("invalid candidate receipt")
	}
	header, e := s.options.Provider.HeaderByNumber(ctx, receipt.BlockNumber)
	if e != nil {
		return nil, nil, nil, e
	}
	if header == nil || header.Number == nil || header.Number.Cmp(receipt.BlockNumber) != 0 || header.Hash() != receipt.BlockHash {
		return nil, nil, nil, errors.New("candidate receipt is not canonical")
	}
	logs := []types.Log{}
	for _, l := range receipt.Logs {
		if l == nil {
			return nil, nil, nil, errors.New("nil receipt log")
		}
		if l.Address != s.scope.ProtocolAddress {
			continue
		}
		if l.Removed || l.TxHash != hash || l.BlockHash != receipt.BlockHash || l.BlockNumber != receipt.BlockNumber.Uint64() {
			return nil, nil, nil, errors.New("invalid receipt log binding")
		}
		logs = append(logs, *l)
	}
	decoder, e := receipts.NewDecoder()
	if e != nil {
		return nil, nil, nil, e
	}
	batch, e := decoder.Decode(receipts.Block{Number: receipt.BlockNumber.Uint64(), Hash: receipt.BlockHash, ParentHash: header.ParentHash, Timestamp: header.Time}, logs)
	if e != nil {
		return nil, nil, nil, e
	}
	notes := []StoredNote{}
	credits := []StoredCredit{}
	draws := []*authenticatedCreditDraw{}
	filter, e := contracts.NewInvisiProtocolFilterer(s.scope.ProtocolAddress, nil)
	if e != nil {
		return nil, nil, nil, e
	}
	for _, l := range logs {
		if draw, e := filter.ParseCreditDraw(l); e == nil {
			draws = append(draws, &authenticatedCreditDraw{draw, header.Time})
		}
	}
	for _, output := range batch.Outputs {
		// Omitted Funds delivery keeps its tree position but has no envelope to decrypt.
		if len(output.Ciphertext) == 0 && len(output.Clue) == 0 {
			continue
		}
		context := core.NoteDecryptionContext{NoteEncryptionContext: core.NoteEncryptionContext{ProtocolContext: s.coreContext(), OutputCommitment: core.BigIntToElement(output.Commitment.Big())}, TokenAddress: core.BigIntToElement(output.Token.Big()), OwnerAddress: core.BigIntToElement(s.wallet.OwnerAddress().Big())}
		if note, e := s.viewing.DecryptNote(output.Ciphertext, context); e == nil {
			if _, e = s.rpc.GetMembership(ctx, s.protocol.Context(), protocolrpc.MembershipInput{TreeID: output.TreeID, LeafIndex: output.LeafIndex, Commitment: output.Commitment}, anchor); e != nil {
				return nil, nil, nil, e
			}
			nf, e := s.wallet.ExportNullifier().ForNote(note)
			if e != nil {
				return nil, nil, nil, e
			}
			notes = append(notes, StoredNote{Note: note, Commitment: output.Commitment, Nullifier: nf, TreeID: output.TreeID, LeafIndex: output.LeafIndex, TransactionHash: hash, Block: &protocolrpc.CanonicalBlock{Number: header.Number.Uint64(), Hash: header.Hash(), ParentHash: header.ParentHash, Timestamp: header.Time}})
			continue
		}
		for _, role := range []core.CreditDeliveryRole{core.CreditOwner, core.CreditSpender} {
			state, e := s.viewing.DecryptCredit(output.Ciphertext, core.CreditNoteDecryptionContext{NoteDecryptionContext: context, Role: role})
			if e != nil {
				continue
			}
			if _, e = s.rpc.GetMembership(ctx, s.protocol.Context(), protocolrpc.MembershipInput{TreeID: output.TreeID, LeafIndex: output.LeafIndex, Commitment: output.Commitment}, anchor); e != nil {
				return nil, nil, nil, e
			}
			nf, e := core.ComputeCreditNullifier(state.Note.ChainID, state.Note.CreditSharedKey, context.OutputCommitment)
			if e != nil {
				return nil, nil, nil, e
			}
			credits = append(credits, StoredCredit{State: state, Commitment: output.Commitment, Nullifier: transaction.FieldHash(nf), TreeID: output.TreeID, LeafIndex: output.LeafIndex, Role: string(role)})
		}
	}
	return notes, credits, draws, nil
}
