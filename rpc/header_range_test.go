package rpc

import (
	"context"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func rangeFixture(t testing.TB, mode string) (*TrustedProvider, *atomic.Int64, *atomic.Int64) {
	t.Helper()
	d := ProtocolDeployment{ChainID: "1", Protocol: ContractDeployment{Address: common.HexToAddress("1"), RuntimeCodeHash: common.HexToHash("2")}}
	var ranges, live atomic.Int64
	chain := make([]*types.Header, 601)
	for i := range chain {
		chain[i] = &types.Header{Number: big.NewInt(int64(i)), Difficulty: big.NewInt(0), Extra: []byte{byte(i)}}
		if mode == "rich" || strings.HasPrefix(mode, "gap-") {
			fee := big.NewInt(9)
			hash := common.HexToHash("abc")
			gas := uint64(7)
			if strings.HasPrefix(mode, "gap-") {
				fee = big.NewInt(0)
				hash = common.Hash{}
				gas = 0
			}
			h := chain[i]
			h.BaseFee = fee
			h.WithdrawalsHash = &hash
			h.BlobGasUsed = &gas
			h.ExcessBlobGas = &gas
			h.ParentBeaconRoot = &hash
			h.RequestsHash = &hash
			h.SlotNumber = &gas
		}
		if i > 0 {
			chain[i].ParentHash = chain[i-1].Hash()
		}
	}
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var raw json.RawMessage
		json.NewDecoder(r.Body).Decode(&raw)
		if len(raw) > 0 && raw[0] == '[' {
			live.Add(1)
			var qs []struct {
				ID     uint64
				Method string
				Params []trustedRequest
			}
			json.Unmarshal(raw, &qs)
			out := make([]any, len(qs))
			for i, q := range qs {
				if q.Method != "invisi_readChainState" {
					t.Errorf("live method %s", q.Method)
				}
				n, _ := new(big.Int).SetString(q.Params[0].Block[2:], 16)
				h := chain[n.Int64()]
				out[i] = map[string]any{"jsonrpc": "2.0", "id": q.ID, "result": trustedResponse{APIVersion: "v1", ProtocolVersion: "1", Deployment: d, Header: h, HeaderHash: h.Hash()}}
			}
			json.NewEncoder(w).Encode(out)
			return
		}
		var q struct {
			ID     uint64
			Method string
			Params []struct {
				From uint64 `json:"fromBlock"`
				To   uint64 `json:"toBlock"`
			}
		}
		json.Unmarshal(raw, &q)
		ranges.Add(1)
		response := map[string]any{"jsonrpc": "2.0", "id": q.ID}
		if q.Method != "invisi_getBlockHeaders" || len(q.Params) != 1 {
			t.Errorf("wrong range query: %s", raw)
			http.Error(w, "wrong query", 400)
			return
		}
		a, z := q.Params[0].From, q.Params[0].To
		if a > z || z-a >= 256 {
			t.Errorf("unbounded query: %s", raw)
		}
		if mode == "method" || mode == "history" || mode == "remote" || mode == "ambiguous" || mode == "bad-error" || (mode == "history-later" && a > 256) {
			code := -32601
			if mode == "history" || mode == "history-later" {
				code = -32004
			}
			if mode == "remote" {
				code = -32000
			}
			response["error"] = map[string]any{"code": code, "message": "unavailable"}
			if mode == "ambiguous" {
				response["result"] = nil
			}
			if mode == "bad-error" {
				response["error"] = map[string]any{"code": -32601}
			}
			json.NewEncoder(w).Encode(response)
			return
		}
		rows := []map[string]any{}
		for n := a; n <= z; n++ {
			h := types.CopyHeader(chain[n])
			if (mode == "parent" && n == a+1) || (mode == "boundary-parent" && n == 257) {
				h.ParentHash = common.HexToHash("ff")
			}
			rows = append(rows, map[string]any{"header": h, "headerHash": h.Hash()})
		}
		if strings.HasPrefix(mode, "gap-") {
			raw, _ := json.Marshal(rows[0]["header"])
			var h map[string]any
			json.Unmarshal(raw, &h)
			field := strings.TrimPrefix(mode, "gap-")
			if strings.HasPrefix(field, "null-") {
				h[strings.TrimPrefix(field, "null-")] = nil
			} else {
				delete(h, field)
			}
			rows[0]["header"] = h
		}
		anchor := chain[z].Hash()
		pins := d
		api := "v1"
		protocol := "1"
		switch mode {
		case "hash":
			rows[0]["headerHash"] = common.HexToHash("99")
		case "count":
			rows = rows[:len(rows)-1]
		case "order":
			rows[0], rows[1] = rows[1], rows[0]
		case "range":
			a++
		case "anchor":
			anchor = common.HexToHash("99")
		case "pins":
			pins.Protocol.RuntimeCodeHash = common.HexToHash("99")
		case "api":
			api = "v2"
		case "protocol":
			protocol = "2"
		case "nil":
			rows[0]["header"] = nil
		case "missing-seal":
			raw, _ := json.Marshal(rows[0]["header"])
			var h map[string]any
			json.Unmarshal(raw, &h)
			delete(h, "nonce")
			delete(h, "mixHash")
			delete(h, "miner")
			rows[0]["header"] = h
		case "incomplete":
			rows[0]["header"] = map[string]any{"number": "0x1"}
		case "envelope":
			response["jsonrpc"] = "1.0"
		}
		response["result"] = map[string]any{"apiVersion": api, "protocolVersion": protocol, "deployment": pins, "range": map[string]any{"fromBlock": a, "toBlock": z, "blockHash": anchor}, "headers": rows}
		json.NewEncoder(w).Encode(response)
	}))
	t.Cleanup(s.Close)
	c, e := NewProtocolRpcClient(s.URL+"/v1", ClientOptions{Deployment: &d})
	if e != nil {
		t.Fatal(e)
	}
	p, e := NewTrustedProvider(c, DeploymentContext{1, d.Protocol.Address})
	if e != nil {
		t.Fatal(e)
	}
	return p, &ranges, &live
}
func headerNumbers(n int) []*big.Int {
	out := make([]*big.Int, n)
	for i := range out {
		out[i] = big.NewInt(int64(i + 1))
	}
	return out
}
func TestHeaderRangeChunks(t *testing.T) {
	p, r, l := rangeFixture(t, "ok")
	got, e := p.HeadersByNumber(context.Background(), headerNumbers(600))
	if e != nil {
		t.Fatal(e)
	}
	if len(got) != 600 || r.Load() != 3 || l.Load() != 0 {
		t.Fatalf("headers %d range %d live %d", len(got), r.Load(), l.Load())
	}
	for i, h := range got {
		if h.Number.Int64() != int64(i+1) {
			t.Fatal("order")
		}
	}
}
func TestHeaderRangeFallbackOnlyExplicitUnavailable(t *testing.T) {
	for _, mode := range []string{"method", "history", "remote", "ambiguous", "bad-error", "hash", "count", "order", "range", "anchor", "pins", "api", "protocol", "parent", "nil", "incomplete", "missing-seal", "envelope"} {
		t.Run(mode, func(t *testing.T) {
			p, r, l := rangeFixture(t, mode)
			got, e := p.HeadersByNumber(context.Background(), headerNumbers(3))
			fallback := mode == "method" || mode == "history"
			if fallback {
				if e != nil || len(got) != 3 || l.Load() != 1 {
					t.Fatalf("fallback headers=%d live=%d err=%v", len(got), l.Load(), e)
				}
			} else if e == nil || got != nil || l.Load() != 0 {
				t.Fatalf("corrupt response accepted/fell back: live=%d err=%v", l.Load(), e)
			}
			if r.Load() != 1 {
				t.Fatal("range not attempted")
			}
		})
	}
}
func TestHeaderRangePreflightAndNonconsecutive(t *testing.T) {
	p, r, l := rangeFixture(t, "ok")
	for _, n := range []*big.Int{nil, big.NewInt(-2), new(big.Int).Lsh(big.NewInt(1), 63)} {
		if _, e := p.HeadersByNumber(context.Background(), []*big.Int{big.NewInt(1), n}); e == nil {
			t.Fatal("invalid height accepted")
		}
	}
	got, e := p.HeadersByNumber(context.Background(), nil)
	if e != nil || len(got) != 0 || r.Load() != 0 || l.Load() != 0 {
		t.Fatal("empty/preflight made network call")
	}
	got, e = p.HeadersByNumber(context.Background(), []*big.Int{big.NewInt(1), big.NewInt(3)})
	if e != nil || len(got) != 2 || r.Load() != 0 || l.Load() != 1 {
		t.Fatalf("nonconsecutive: %v", e)
	}
}
func BenchmarkHistoricalHeaders(b *testing.B) {
	for _, mode := range []string{"ok", "method"} {
		b.Run(mode, func(b *testing.B) {
			p, r, l := rangeFixture(b, mode)
			ns := headerNumbers(600)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, e := p.HeadersByNumber(context.Background(), ns); e != nil {
					b.Fatal(e)
				}
			}
			b.ReportMetric(float64(r.Load()+l.Load())/float64(b.N), "requests/op")
		})
	}
}

func TestHeaderRangeCrossChunkParent(t *testing.T) {
	p, r, l := rangeFixture(t, "boundary-parent")
	got, e := p.HeadersByNumber(context.Background(), headerNumbers(300))
	if e == nil || got != nil || r.Load() != 2 || l.Load() != 0 {
		t.Fatalf("cross-chunk discontinuity accepted: %v", e)
	}
}
func TestHeaderRangeMissingLaterChunkDiscardsPartialHistory(t *testing.T) {
	p, r, l := rangeFixture(t, "history-later")
	got, e := p.HeadersByNumber(context.Background(), headerNumbers(300))
	if e != nil || len(got) != 300 || r.Load() != 2 || l.Load() != 15 {
		t.Fatalf("partial fallback: headers %d range %d live %d err %v", len(got), r.Load(), l.Load(), e)
	}
}
func TestHeaderRangeKeepsLiveSingleHeaderAndFinalityReads(t *testing.T) {
	d := ProtocolDeployment{ChainID: "1", Protocol: ContractDeployment{Address: common.HexToAddress("1"), RuntimeCodeHash: common.HexToHash("2")}}
	var blocks []string
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var q struct {
			ID     uint64
			Method string
			Params []trustedRequest
		}
		json.NewDecoder(r.Body).Decode(&q)
		if q.Method != "invisi_readChainState" {
			t.Errorf("live read used history: %s", q.Method)
		}
		blocks = append(blocks, q.Params[0].Block)
		h := &types.Header{Number: big.NewInt(10), Difficulty: big.NewInt(0), Extra: []byte{byte(len(blocks))}}
		json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": q.ID, "result": trustedResponse{APIVersion: "v1", ProtocolVersion: "1", Deployment: d, Header: h, HeaderHash: h.Hash()}})
	}))
	defer s.Close()
	c, _ := NewProtocolRpcClient(s.URL+"/v1", ClientOptions{Deployment: &d})
	p, _ := NewTrustedProvider(c, DeploymentContext{1, d.Protocol.Address})
	a, e := p.HeaderByNumber(context.Background(), big.NewInt(10))
	if e != nil {
		t.Fatal(e)
	}
	z, e := p.HeaderByNumber(context.Background(), big.NewInt(10))
	if e != nil || a.Hash() == z.Hash() {
		t.Fatal("canonical fence reused header", e)
	}
	for _, tag := range []FinalityTag{FinalityLatest, FinalitySafe, FinalityFinalized} {
		if _, e = p.HeaderByTag(context.Background(), tag); e != nil {
			t.Fatal(e)
		}
	}
	want := []string{"0xa", "0xa", "latest", "safe", "finalized"}
	for i, v := range want {
		if blocks[i] != v {
			t.Fatalf("live block %s want %s", blocks[i], v)
		}
	}
}

func TestHeaderRangePreservesPostForkFields(t *testing.T) {
	p, _, _ := rangeFixture(t, "rich")
	got, e := p.HeadersByNumber(context.Background(), headerNumbers(1))
	if e != nil {
		t.Fatal(e)
	}
	h := got[0]
	hash := common.HexToHash("abc")
	if h.BaseFee == nil || h.BaseFee.Int64() != 9 || h.WithdrawalsHash == nil || *h.WithdrawalsHash != hash || h.BlobGasUsed == nil || *h.BlobGasUsed != 7 || h.ExcessBlobGas == nil || *h.ExcessBlobGas != 7 || h.ParentBeaconRoot == nil || *h.ParentBeaconRoot != hash || h.RequestsHash == nil || *h.RequestsHash != hash || h.SlotNumber == nil || *h.SlotNumber != 7 {
		t.Fatal("post-fork header fields lost")
	}
}
func TestHeaderRangeRejectsZeroValuedOptionalGaps(t *testing.T) {
	for _, field := range []string{"baseFeePerGas", "withdrawalsRoot", "blobGasUsed", "excessBlobGas", "parentBeaconBlockRoot", "requestsHash"} {
		for _, prefix := range []string{"gap-", "gap-null-"} {
			t.Run(prefix+field, func(t *testing.T) {
				p, _, live := rangeFixture(t, prefix+field)
				got, e := p.HeadersByNumber(context.Background(), headerNumbers(1))
				if e == nil || got != nil || live.Load() != 0 {
					t.Fatalf("accepted missing zero-valued %s or used fallback: %v", field, e)
				}
			})
		}
	}
}
