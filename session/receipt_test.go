package session

import (
	"context"
	"encoding/binary"
	"fmt"
	"math/big"
	"reflect"
	"testing"

	"github.com/invisi-labs/invisi-sdk-go/protocol/receipts"
	protocolrpc "github.com/invisi-labs/invisi-sdk-go/rpc"
	"github.com/invisi-labs/invisi-sdk-go/wallet"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/invisi-labs/invisi-sdk-go/primitives"
	"github.com/invisi-labs/invisi-sdk-go/protocol"
)

type receiptProvider struct {
	sessionProvider
	receipt *types.Receipt
}

func (p receiptProvider) TransactionReceipt(context.Context, common.Hash) (*types.Receipt, error) {
	return p.receipt, nil
}

func TestOwnedReceiptRequiresCanonicalInsertionAndMembership(t *testing.T) {
	ctx := context.Background()
	session, token, _ := testSession(t)
	original := session.options.Provider.(sessionProvider)
	state, err := session.options.Store.ReadSnapshot(ctx, session.scope)
	if err != nil {
		t.Fatal(err)
	}
	var stored StoredNote
	for _, note := range state.Notes {
		stored = note
	}
	ciphertext, err := primitives.EncryptPrivateTransferNote(session.wallet.ViewingPublicKey(), stored.Note, primitives.NoteEncryptionContext{ProtocolContext: session.coreContext(), OutputCommitment: primitives.BigIntToElement(stored.Commitment.Big())})
	if err != nil {
		t.Fatal(err)
	}
	tx := common.HexToHash("0x1234")
	clue := make([]byte, 68)
	clue[64] = 24
	logs := unifiedNoteLogs(t, session.scope.ProtocolAddress, tx, original.head.Hash(), token, stored.Commitment, ciphertext, clue)
	log := logs[2]
	receipt := &types.Receipt{TxHash: tx, Status: types.ReceiptStatusSuccessful, BlockNumber: big.NewInt(10), BlockHash: original.head.Hash(), Logs: logs}
	provider := receiptProvider{original, receipt}
	session.options.Provider = provider
	session.rpc = session.rpc.WithProvider(provider)

	// An App error wrapping the nonrecipient sentinel must not commit inventory or checkpoints.
	for _, appErr := range []error{wallet.ErrNotRecipient, fmt.Errorf("App failure: %w", wallet.ErrNotRecipient)} {
		decoder, _ := receipts.NewDecoder()
		batch, e := decoder.Decode(receipts.Block{Number: 10, Hash: receipt.BlockHash, ParentHash: original.head.ParentHash, Timestamp: original.head.Time}, []types.Log{*logs[0], *logs[1], *logs[2], *logs[3]})
		if e != nil {
			t.Fatal(e)
		}
		session.publicHistory, e = NewPublicHistory(PublicHistoryOptions{Protocol: session.protocol, Provider: provider, ProtocolRPC: session.rpc})
		if e != nil {
			t.Fatal(e)
		}
		block := *state.ObservedThrough
		page := PublicPage{Range: BlockRange{10, 10}, Block: block, Headers: []protocolrpc.CanonicalBlock{block}, Transactions: []PublicTransaction{{PositionValidated: true, Receipt: receipt, Header: original.head, Batch: batch}}}
		if e = PutPublicPage(ctx, session.publicHistory.Store(), session.publicHistory.Key(), page, 0); e != nil {
			t.Fatal(e)
		}
		height := uint64(10)
		pending := &PendingWalletRefresh{session: session, base: state, options: RefreshOptions{FromBlock: 10, ToBlock: &height}}

		failed := testDeliveryDecryptor(session)
		failed.Note = func(context.Context, []byte, primitives.NoteDecryptionContext) ([]byte, error) { return nil, appErr }
		if _, e := pending.Decrypt(ctx, failed); e == nil {
			t.Fatal("App error was skipped as nonrecipient")
		}
		after, e := session.options.Store.ReadSnapshot(ctx, session.scope)
		if e != nil {
			t.Fatal(e)
		}
		if !reflect.DeepEqual(state, after) {
			t.Fatal("App error advanced inventory or checkpoints")
		}
	}
	notes, credits, _, err := readTransactionForTest(session, ctx, tx, *state.ObservedThrough)
	if err != nil || len(notes) != 1 || len(credits) != 0 || notes[0].Commitment != stored.Commitment {
		t.Fatal(notes, err)
	}
	log.BlockHash = common.HexToHash("0xdead")
	if _, _, _, err := readTransactionForTest(session, ctx, tx, *state.ObservedThrough); err == nil {
		t.Fatal("misbound event accepted")
	}
	log.BlockHash = original.head.Hash()
	receipt.TxHash = common.HexToHash("0xbeef")
	if _, _, _, err := readTransactionForTest(session, ctx, tx, *state.ObservedThrough); err == nil {
		t.Fatal("substituted receipt accepted")
	}
}

func unifiedNoteLogs(t *testing.T, core common.Address, tx, block common.Hash, token common.Address, commitment common.Hash, cipher, clue []byte) []*types.Log {
	t.Helper()
	a, e := protocol.ProtocolABI()
	if e != nil {
		t.Fatal(e)
	}
	id, err := protocol.NoteOperationID(protocol.TransactionContext{ChainID: big.NewInt(31337), ProtocolAddress: core}, 0, protocol.CrossTreeTransactParams{Token: token, Sources: []protocol.InputSource{}, Nullifiers: [][32]byte{}, NoteCommitments: [][32]byte{commitment}, WithdrawalAmount: new(big.Int), Deadline: big.NewInt(9999999999), FeeAmount: new(big.Int)}, []protocol.Delivery{{Ciphertext: cipher, Clue: clue}}, protocol.FundingParams{Payer: token, Amount: big.NewInt(1)})
	if err != nil {
		t.Fatal(err)
	}
	pack := func(name string, index uint, indexed []common.Hash, args ...any) *types.Log {
		event := a.Events[name]
		data, e := event.Inputs.NonIndexed().Pack(args...)
		if e != nil {
			t.Fatal(e)
		}
		return &types.Log{Address: core, Topics: append([]common.Hash{event.ID}, indexed...), Data: data, BlockNumber: 10, BlockHash: block, TxHash: tx, Index: index}
	}
	packed := binary.BigEndian.AppendUint16(nil, uint16(len(cipher)))
	packed = append(packed, cipher...)
	if len(cipher) > 0 {
		packed = append(packed, clue...)
	}
	return []*types.Log{pack("NoteOperation", 0, []common.Hash{id, common.BytesToHash(token.Bytes())}, uint8(0), big.NewInt(1), big.NewInt(9999999999), common.Address{}, big.NewInt(0)), pack("CommitmentsInserted", 1, []common.Hash{common.BigToHash(big.NewInt(1))}, uint32(0), [][32]byte{commitment}), pack("NoteOutputs", 2, []common.Hash{id, common.BigToHash(big.NewInt(1))}, uint32(0), packed), pack("PublicDeposit", 3, []common.Hash{id, common.BytesToHash(token.Bytes()), common.BytesToHash(token.Bytes())}, big.NewInt(1))}
}
