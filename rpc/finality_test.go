package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
)

type finalityFixture struct {
	tags    map[int64]*types.Header
	calls   []int64
	err     error
	receipt *types.Receipt
}

func (f *finalityFixture) HeaderByNumber(_ context.Context, n *big.Int) (*types.Header, error) {
	key := int64(-2)
	if n != nil {
		key = n.Int64()
	}
	f.calls = append(f.calls, key)
	if key < 0 && f.err != nil {
		return nil, f.err
	}
	if h := f.tags[key]; h != nil {
		return types.CopyHeader(h), nil
	}
	return nil, nil
}
func (f *finalityFixture) TransactionReceipt(context.Context, common.Hash) (*types.Receipt, error) {
	return f.receipt, nil
}
func TestNativeFinalityIndependentLevels(t *testing.T) {
	h := func(n int64) *types.Header {
		return &types.Header{Number: big.NewInt(n), Difficulty: big.NewInt(1), Time: uint64(n)}
	}
	f := &finalityFixture{tags: map[int64]*types.Header{}}
	for _, n := range []int64{10, 12, 15} {
		f.tags[n] = h(n)
	}
	f.tags[-3] = f.tags[10]
	f.tags[-4] = f.tags[12]
	f.tags[-2] = f.tags[15]
	got := ReadChainFinality(context.Background(), f)
	if got.Finalized.Status != FinalityAvailable || got.Finalized.Block.Number != 10 || got.Safe.Block.Number != 12 {
		t.Fatalf("%+v", got)
	}
	if f.calls[0] != -3 {
		t.Fatal("read higher head before lower", f.calls)
	}
	hash := common.HexToHash("0x1234")
	f.receipt = &types.Receipt{TxHash: hash, BlockNumber: big.NewInt(12), BlockHash: f.tags[12].Hash()}
	c := ReadTransactionConfirmation(context.Background(), f, hash)
	if c.Level != "safe" || c.Status != FinalityAvailable {
		t.Fatalf("%+v", c)
	}
	f.err = ErrFinalityUnsupported
	got = ReadChainFinality(context.Background(), f)
	if got.Finalized.Status != FinalityUnsupported {
		t.Fatal(got)
	}
	f.err = errors.New("temporary read failure")
	got = ReadChainFinality(context.Background(), f)
	if got.Finalized.Status != FinalityUnavailable {
		t.Fatal(got)
	}
}

func TestNativeFinalityNullStallOrderingAndReceiptReorg(t *testing.T) {
	h := func(n int64) *types.Header {
		return &types.Header{Number: big.NewInt(n), Difficulty: big.NewInt(1), Time: uint64(n)}
	}
	f := &finalityFixture{tags: map[int64]*types.Header{}}
	got := ReadChainFinality(context.Background(), f)
	if got.Finalized.Status != FinalityUnavailable {
		t.Fatal(got)
	}
	f.tags[-3] = h(20)
	f.tags[20] = f.tags[-3]
	f.tags[-4] = h(10)
	f.tags[10] = f.tags[-4]
	f.tags[-2] = h(30)
	f.tags[30] = f.tags[-2]
	got = ReadChainFinality(context.Background(), f)
	if got.Safe.Status != FinalityInconsistent || got.Finalized.Block != nil {
		t.Fatalf("unordered native heads promoted %+v", got)
	}
	f.tags[-3] = f.tags[10]
	f.tags[-4] = f.tags[20]
	for i := 0; i < 3; i++ {
		got = ReadChainFinality(context.Background(), f)
		if got.Finalized.Status != FinalityAvailable || got.Finalized.Block.Number != 10 {
			t.Fatal("stall synthesized progression", got)
		}
	}
	hash := common.HexToHash("0x1234")
	f.receipt = &types.Receipt{TxHash: hash, BlockNumber: big.NewInt(10), BlockHash: common.HexToHash("0xdead")}
	c := ReadTransactionConfirmation(context.Background(), f, hash)
	if c.Status != FinalityInconsistent || c.Level != "pending" {
		t.Fatal("reorged receipt promoted", c)
	}
}

type finalityChainFixture struct{ *finalityFixture }

func (f finalityChainFixture) ChainID(context.Context) (*big.Int, error) { return big.NewInt(1), nil }
func (f finalityChainFixture) CodeAt(context.Context, common.Address, *big.Int) ([]byte, error) {
	return nil, nil
}
func (f finalityChainFixture) CallContract(context.Context, ethereum.CallMsg, *big.Int) ([]byte, error) {
	return nil, nil
}
func TestIndexerFinalityCannotPromotePrimaryUnsupported(t *testing.T) {
	h := &types.Header{Number: big.NewInt(10), Difficulty: big.NewInt(1), Time: 10}
	block := CanonicalBlock{Number: 10, Hash: h.Hash(), ParentHash: h.ParentHash, Timestamp: h.Time}
	claimed := ChainFinality{Latest: FinalityHead{Status: FinalityAvailable, Block: &block}, Safe: FinalityHead{Status: FinalityAvailable, Block: &block}, Finalized: FinalityHead{Status: FinalityAvailable, Block: &block}}
	raw, _ := json.Marshal(claimed)
	f := finalityChainFixture{&finalityFixture{tags: map[int64]*types.Header{10: h}, err: ErrFinalityUnsupported}}
	got, err := checkedFinality(context.Background(), f, raw)
	if err != nil || got.Finalized.Status != FinalityUnsupported || got.Finalized.Block != nil {
		t.Fatalf("indexer promoted primary: %+v %v", got, err)
	}
	for _, tag := range []int64{-3, -4, -2} {
		f.tags[tag] = h
	}
	f.err = nil
	got, err = checkedFinality(context.Background(), f, raw)
	if err != nil || got.Finalized.Status != FinalityAvailable {
		t.Fatal(got, err)
	}
}

func TestNativeFinalityMalformedWireHashIsInconsistent(t *testing.T) {
	h := &types.Header{Number: big.NewInt(10), Difficulty: big.NewInt(1), Time: 10}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var q struct {
			ID json.RawMessage `json:"id"`
		}
		_ = json.NewDecoder(r.Body).Decode(&q)
		raw, _ := json.Marshal(h)
		var block map[string]any
		_ = json.Unmarshal(raw, &block)
		block["hash"] = common.HexToHash("0xdead")
		_ = json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": q.ID, "result": block})
	}))
	defer server.Close()
	p, err := ethclient.Dial(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	got := ReadChainFinality(context.Background(), p)
	if got.Finalized.Status != FinalityInconsistent || got.Finalized.Block != nil {
		t.Fatalf("malformed wire promoted %+v", got)
	}
}
