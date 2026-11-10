package session

import (
	"context"
	"encoding/binary"
	"fmt"
	"math/big"
	"reflect"
	"testing"

	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
	"github.com/invisi-lab/invisi-sdk-go/wallet"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/invisi-lab/invisi-sdk-go/primitives"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
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
	abi, err := protocol.ProtocolABI()
	if err != nil {
		t.Fatal(err)
	}
	event := abi.Events["BatchOutputs"]
	// Pack the current event dynamically so the test is bound to the published ABI.
	values := []any{}
	topics := []common.Hash{event.ID}
	for _, arg := range event.Inputs {
		var value any
		switch arg.Name {
		case "treeId":
			value = big.NewInt(1)
		case "firstOutputOffset":
			value = big.NewInt(0)
		case "startLeafIndex":
			value = uint32(0)
		case "noteCommitments":
			value = [][32]byte{stored.Commitment}
		case "tokens":
			value = []common.Address{token}
		case "packedDeliveries":
			clue := make([]byte, 68)
			clue[64] = 24
			value = append(append(binary.BigEndian.AppendUint16(nil, uint16(len(ciphertext))), ciphertext...), clue...)
		default:
			t.Fatalf("unexpected BatchOutputs field %s", arg.Name)
		}
		if arg.Indexed {
			switch v := value.(type) {
			case *big.Int:
				topics = append(topics, common.BigToHash(v))
			case uint32:
				topics = append(topics, common.BigToHash(new(big.Int).SetUint64(uint64(v))))
			default:
				t.Fatal("unsupported indexed test field")
			}
		} else {
			values = append(values, value)
		}
	}
	data, err := event.Inputs.NonIndexed().Pack(values...)
	if err != nil {
		t.Fatal(err)
	}
	tx := common.HexToHash("0x1234")
	log := &types.Log{Address: session.scope.ProtocolAddress, Topics: topics, Data: data, BlockNumber: 10, BlockHash: original.head.Hash(), TxHash: tx}
	receipt := &types.Receipt{TxHash: tx, Status: types.ReceiptStatusSuccessful, BlockNumber: big.NewInt(10), BlockHash: original.head.Hash(), Logs: []*types.Log{log}}
	provider := receiptProvider{original, receipt}
	session.options.Provider = provider
	session.rpc = session.rpc.WithProvider(provider)

	// An App error wrapping the nonrecipient sentinel must not commit inventory or checkpoints.
	for _, appErr := range []error{wallet.ErrNotRecipient, fmt.Errorf("App failure: %w", wallet.ErrNotRecipient)} {
		clue := make([]byte, 68)
		clue[64] = 24
		pending := &PendingWalletRefresh{session: session, base: state, options: RefreshOptions{FromBlock: 10}, pages: &refreshPages{
			notePages:   map[string]protocolrpc.ProtocolNoteUpdates{"10:10": {Block: *state.ObservedThrough, Notes: []protocolrpc.ProtocolNoteOutput{{Block: *state.ObservedThrough, TreeID: 1, LeafIndex: 0, Commitment: stored.Commitment, TokenAddress: token, Ciphertext: ciphertext, Clue: clue, TransactionHash: tx}}}},
			creditPages: map[string]protocolrpc.ProtocolCreditUpdates{"10:10": {Block: *state.ObservedThrough}},
		}}
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
