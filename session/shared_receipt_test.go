package session

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
	"github.com/invisi-lab/invisi-sdk-go/wallet"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

type countingReceiptProvider struct {
	sessionProvider
	receipt *types.Receipt
	calls   *atomic.Int32
}

func (p countingReceiptProvider) TransactionReceipt(context.Context, common.Hash) (*types.Receipt, error) {
	p.calls.Add(1)
	return p.receipt, nil
}
func TestTwoAccountsReuseDownloadedCanonicalReceiptAndPositions(t *testing.T) {
	ctx := context.Background()
	s, n, p := publicNoteFixture(t)
	base := s.options.Provider.(sessionProvider)
	tx := p.Transactions[0]
	o := tx.Batch.Outputs[0]
	abi, _ := protocol.ProtocolABI()
	event := abi.Events["BatchOutputs"]
	topics := []common.Hash{event.ID}
	values := []any{}
	for _, arg := range event.Inputs {
		var v any
		switch arg.Name {
		case "treeId":
			v = big.NewInt(1)
		case "firstOutputOffset":
			v = big.NewInt(0)
		case "startLeafIndex":
			v = uint32(0)
		case "noteCommitments":
			v = [][32]byte{n.Commitment}
		case "tokens":
			v = []common.Address{o.Token}
		case "packedDeliveries":
			v = append(append(binary.BigEndian.AppendUint16(nil, uint16(len(o.Ciphertext))), o.Ciphertext...), o.Clue...)
		default:
			t.Fatal(arg.Name)
		}
		if arg.Indexed {
			switch v := v.(type) {
			case *big.Int:
				topics = append(topics, common.BigToHash(v))
			case uint32:
				topics = append(topics, common.BigToHash(new(big.Int).SetUint64(uint64(v))))
			}
		} else {
			values = append(values, v)
		}
	}
	data, e := event.Inputs.NonIndexed().Pack(values...)
	if e != nil {
		t.Fatal(e)
	}
	tx.Receipt.Logs = []*types.Log{{Address: s.scope.ProtocolAddress, Topics: topics, Data: data, BlockNumber: 10, BlockHash: p.Block.Hash, TxHash: tx.Receipt.TxHash}}
	var receiptCalls, updateCalls atomic.Int32
	provider := countingReceiptProvider{base, tx.Receipt, &receiptCalls}
	deployment := s.protocol.Deployment()
	root, _ := primitives.HexToElement(n.Commitment.Hex())
	for range 16 {
		root, _ = primitives.Poseidon2Hash(root, primitives.ElementFromUint64(0))
	}
	anchor := map[string]any{"treeId": 1, "root": common.Hash(root.Bytes()), "leafCount": 1, "block": p.Block}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			ID     uint64
			Method string
		}
		_ = json.NewDecoder(r.Body).Decode(&request)
		var result any
		switch request.Method {
		case "invisi_getStatus":
			result = map[string]any{"apiVersion": "v1", "protocolVersion": "1", "deployment": deployment, "indexedHead": p.Block}
		case "invisi_getNoteUpdates":
			updateCalls.Add(1)
			result = map[string]any{"blockHash": p.Block.Hash, "notes": []any{map[string]any{"block": p.Block, "treeId": 1, "leafIndex": 0, "commitment": n.Commitment, "token": o.Token, "ciphertext": wallet.Hex(o.Ciphertext), "clue": wallet.Hex(o.Clue), "transactionHash": tx.Receipt.TxHash, "logIndex": 0, "kind": "BatchOutputs"}}}
		case "invisi_getCreditTransitions":
			result = []any{}
		case "invisi_getRoots":
			result = []any{anchor}
		case "invisi_getMembership":
			result = map[string]any{"anchor": anchor, "leafIndex": 0, "commitment": n.Commitment, "siblings": make([]common.Hash, 16)}
		default:
			t.Errorf("unexpected %s", request.Method)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": request.ID, "result": result})
	}))
	defer server.Close()
	client, e := protocolrpc.NewProtocolRpcClient(server.URL+"/v1", protocolrpc.ClientOptions{Deployment: &deployment})
	if e != nil {
		t.Fatal(e)
	}
	s.rpc = client.WithProvider(provider)
	s.options.Provider = provider
	s.publicHistory, e = NewPublicHistory(PublicHistoryOptions{Protocol: s.protocol, Provider: provider, ProtocolRPC: s.rpc})
	if e != nil {
		t.Fatal(e)
	}
	from := uint64(10)
	if e = s.SynchronizeHistory(ctx, ScanOptions{FromBlock: &from}); e != nil {
		t.Fatal(e)
	}
	if _, e = s.ScanHistory(ctx, testDeliveryDecryptor(s), ScanOptions{FromBlock: &from, Offline: true}); e != nil {
		t.Fatal(e)
	}
	b := *s
	b.options = s.options
	b.options.Store = NewMemoryWalletStore()
	b.refreshGate = make(chan struct{}, 1)
	b.wallet, e = localTestWallet("0x" + strings.Repeat("0", 63) + "2")
	if e != nil {
		t.Fatal(e)
	}
	b.scope.OwnerAddress = b.wallet.OwnerAddress()
	if _, e = b.ScanHistory(ctx, testDeliveryDecryptor(&b), ScanOptions{FromBlock: &from, Offline: true}); e != nil {
		t.Fatal(e)
	}
	if receiptCalls.Load() != 1 || updateCalls.Load() != 1 {
		t.Fatalf("shared download repeated receipts=%d updates=%d", receiptCalls.Load(), updateCalls.Load())
	}
	aState, _ := s.options.Store.ReadSnapshot(ctx, s.scope)
	bState, _ := b.options.Store.ReadSnapshot(ctx, b.scope)
	if len(aState.Notes) != 1 || len(bState.Notes) != 0 || len(bState.Scan.Coverage) != 1 {
		t.Fatal("account private results or coverage not independent")
	}
}
