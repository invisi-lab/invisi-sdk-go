package session

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
	"github.com/invisi-lab/invisi-sdk-go/protocol/receipts"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
	"github.com/invisi-lab/invisi-sdk-go/wallet"
	"math/big"
	"testing"
)

type detectingDecryptor struct {
	wallet.DeliveryDecryptorFuncs
	detect func([]wallet.DeliveryDetectionItem) (wallet.DeliveryDetectionResult, error)
}

func (d detectingDecryptor) Detect(_ context.Context, items []wallet.DeliveryDetectionItem) (wallet.DeliveryDetectionResult, error) {
	return d.detect(items)
}
func publicNoteFixture(t *testing.T) (*WalletSession, StoredNote, PublicPage) {
	t.Helper()
	s, _, _ := testSession(t)
	v, e := s.options.Store.ReadSnapshot(context.Background(), s.scope)
	if e != nil {
		t.Fatal(e)
	}
	var n StoredNote
	for _, v := range v.Notes {
		n = v
	}
	h := s.options.Provider.(sessionProvider).head
	h.Difficulty = big.NewInt(0)
	b := protocolrpc.CanonicalBlock{Number: 10, Hash: h.Hash(), ParentHash: h.ParentHash, Timestamp: h.Time}
	cipher, e := primitives.EncryptPrivateTransferNote(s.wallet.ViewingPublicKey(), n.Note, primitives.NoteEncryptionContext{ProtocolContext: s.coreContext(), OutputCommitment: primitives.BigIntToElement(n.Commitment.Big())})
	if e != nil {
		t.Fatal(e)
	}
	clue := make([]byte, 68)
	clue[64] = 24
	hash := common.HexToHash("0x1234")
	tx := PublicTransaction{PositionValidated: true, Receipt: &types.Receipt{TxHash: hash, BlockNumber: h.Number, BlockHash: h.Hash(), Status: types.ReceiptStatusSuccessful, Logs: []*types.Log{}}, Header: h, Batch: receipts.Batch{Outputs: []receipts.Output{{Token: common.BigToAddress(primitives.ElementToBigInt(n.Note.TokenAddress)), Commitment: n.Commitment, Ciphertext: cipher, Clue: clue, TreeID: 1, LeafIndex: 0, TransactionHash: hash}}}}
	p := PublicPage{Range: BlockRange{10, 10}, Block: b, Headers: []protocolrpc.CanonicalBlock{b}, Transactions: []PublicTransaction{tx}}
	if e = PutPublicPage(context.Background(), s.publicHistory.Store(), s.publicHistory.Key(), p, 0); e != nil {
		t.Fatal(e)
	}
	raw, _ := json.Marshal(p)
	var check PublicPage
	if e = json.Unmarshal(raw, &check); e != nil {
		t.Fatal("page codec", e)
	}
	s.options.Store = NewMemoryWalletStore()
	return s, n, p
}
func TestNativeBatchContextAndAmbiguityPersistAndRetry(t *testing.T) {
	s, n, _ := publicNoteFixture(t)
	ctx := context.Background()
	from := uint64(10)
	called := false
	d := detectingDecryptor{DeliveryDecryptorFuncs: wallet.DeliveryDecryptorFuncs{Note: func(context.Context, []byte, primitives.NoteDecryptionContext) ([]byte, error) {
		return nil, wallet.ErrAuthenticationFailed
	}}, detect: func(items []wallet.DeliveryDetectionItem) (wallet.DeliveryDetectionResult, error) {
		called = true
		if len(items) != 1 {
			t.Fatalf("%+v", items)
		}
		i := items[0]
		if len(i.Clue) != 68 || i.Owner != s.wallet.OwnerAddress() || i.ChainID != s.scope.ChainID || i.ProtocolAddress != s.scope.ProtocolAddress || i.OutputCommitment != n.Commitment || i.TokenAddress == ([20]byte{}) || len(i.Ciphertext) == 0 {
			t.Fatalf("incomplete native context %+v", i)
		}
		return wallet.DeliveryDetectionResult{Owner: s.wallet.OwnerAddress(), Results: []wallet.DeliveryDetectionDecision{{ID: i.ID, Status: "candidate"}}}, nil
	}}
	if _, e := s.ScanHistory(ctx, d, ScanOptions{FromBlock: &from, Offline: true}); e != nil {
		t.Fatal(e)
	}
	v, e := s.options.Store.ReadSnapshot(ctx, s.scope)
	if e != nil {
		t.Fatal(e)
	}
	if !called || len(v.Scan.Unresolved) != 1 || len(v.Notes) != 0 || len(v.Scan.Coverage) != 1 {
		t.Fatal("ambiguous delivery lost or cursor missing")
	}
	resolved, e := s.RetryUnresolved(ctx, testDeliveryDecryptor(s))
	if e != nil || resolved != 1 {
		t.Fatal(resolved, e)
	}
	v, e = s.options.Store.ReadSnapshot(ctx, s.scope)
	if e != nil || len(v.Notes) != 1 || len(v.Scan.Unresolved) != 0 {
		t.Fatal("retry failed", e)
	}
}
func TestOperationalFailureDoesNotAdvancePrivateBatch(t *testing.T) {
	s, _, _ := publicNoteFixture(t)
	from := uint64(10)
	d := wallet.DeliveryDecryptorFuncs{Note: func(context.Context, []byte, primitives.NoteDecryptionContext) ([]byte, error) {
		return nil, errors.New("wallet locked")
	}}
	if _, e := s.ScanHistory(context.Background(), d, ScanOptions{FromBlock: &from, Offline: true}); e == nil {
		t.Fatal("locked wallet skipped")
	}
	v, e := s.options.Store.ReadSnapshot(context.Background(), s.scope)
	if e != nil || len(v.Scan.Coverage) != 0 || len(v.Notes) != 0 {
		t.Fatal("failed batch published")
	}
}
