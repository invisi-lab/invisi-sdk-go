package sdk

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/invisi-lab/invisi-sdk-go/core"
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
	ciphertext, err := core.EncryptPrivateTransferNote(session.viewing.GetPublicKey(), stored.Note, core.NoteEncryptionContext{ProtocolContext: session.coreContext(), OutputCommitment: core.BigIntToElement(stored.Commitment.Big())})
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
		case "commitments":
			value = [][32]byte{stored.Commitment}
		case "tokens":
			value = []common.Address{token}
		case "packedDeliveries":
			value = append(ciphertext, make([]byte, 68)...)
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
	notes, credits, _, err := session.readTransaction(ctx, tx, *state.ObservedThrough)
	if err != nil || len(notes) != 1 || len(credits) != 0 || notes[0].Commitment != stored.Commitment {
		t.Fatal(notes, err)
	}
	log.BlockHash = common.HexToHash("0xdead")
	if _, _, _, err := session.readTransaction(ctx, tx, *state.ObservedThrough); err == nil {
		t.Fatal("misbound event accepted")
	}
	log.BlockHash = original.head.Hash()
	receipt.TxHash = common.HexToHash("0xbeef")
	if _, _, _, err := session.readTransaction(ctx, tx, *state.ObservedThrough); err == nil {
		t.Fatal("substituted receipt accepted")
	}
}
