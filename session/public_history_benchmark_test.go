package session

import (
	"context"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	protocolrpc "github.com/invisi-labs/invisi-sdk-go/rpc"
)

// Exercise the actual cold PublicHistory.Page pipeline through HTTP, including
// all 256 complete headers, update decoding and final canonical fencing. Each
// HTTP round trip has 1ms injected latency; the server has no real finality field
// or transactions, so this measures discovery rather than proving or dense logs.
func BenchmarkPublicHistoryColdPage(b *testing.B) {
	s, _, _ := testSession(b)
	d := s.protocol.Deployment()
	headers := make(map[uint64]*types.Header, 256)
	for n := uint64(1); n <= 256; n++ {
		h := &types.Header{Number: new(big.Int).SetUint64(n), Time: n, Difficulty: big.NewInt(0)}
		if n > 1 {
			h.ParentHash = headers[n-1].Hash()
		}
		headers[n] = h
	}
	block := func(n uint64) protocolrpc.CanonicalBlock {
		h := headers[n]
		return protocolrpc.CanonicalBlock{Number: n, Hash: h.Hash(), ParentHash: h.ParentHash, Timestamp: h.Time}
	}
	var calls, rows, statuses atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		time.Sleep(time.Millisecond)
		var raw json.RawMessage
		json.NewDecoder(r.Body).Decode(&raw)
		type request struct {
			ID     uint64
			Method string
			Params []json.RawMessage
		}
		var requests []request
		batch := len(raw) > 0 && raw[0] == '['
		if batch {
			json.Unmarshal(raw, &requests)
		} else {
			requests = make([]request, 1)
			json.Unmarshal(raw, &requests[0])
		}
		responses := make([]map[string]any, 0, len(requests))
		for _, q := range requests {
			rows.Add(1)
			var result any
			switch q.Method {
			case "invisi_readChainState":
				var item struct{ Kind, Block string }
				json.Unmarshal(q.Params[0], &item)
				n := uint64(256)
				if item.Block != "latest" {
					v, ok := new(big.Int).SetString(item.Block[2:], 16)
					if !ok {
						b.Errorf("invalid block %s", item.Block)
						return
					}
					n = v.Uint64()
				}
				h := headers[n]
				result = map[string]any{"apiVersion": "v1", "protocolVersion": "1", "deployment": d, "header": h, "headerHash": h.Hash()}
			case "invisi_getStatus":
				statuses.Add(1)
				result = map[string]any{"apiVersion": "v1", "protocolVersion": "1", "deployment": d, "indexedHead": block(256)}
			case "invisi_getNoteUpdates":
				result = map[string]any{"blockHash": headers[256].Hash(), "notes": []any{}}
			case "invisi_getCreditTransitions":
				result = []any{}
			default:
				b.Errorf("unexpected method %s", q.Method)
			}
			responses = append(responses, map[string]any{"jsonrpc": "2.0", "id": q.ID, "result": result})
		}
		if batch {
			json.NewEncoder(w).Encode(responses)
		} else {
			json.NewEncoder(w).Encode(responses[0])
		}
	}))
	defer server.Close()
	client, err := protocolrpc.NewProtocolRpcClient(server.URL+"/v1", protocolrpc.ClientOptions{Deployment: &d})
	if err != nil {
		b.Fatal(err)
	}
	provider, err := protocolrpc.NewTrustedProvider(client, s.protocol.Context())
	if err != nil {
		b.Fatal(err)
	}
	client = client.WithProvider(provider)
	b.ResetTimer()
	for range b.N {
		h, err := NewPublicHistory(PublicHistoryOptions{Protocol: s.protocol, Provider: provider, ProtocolRPC: client})
		if err != nil {
			b.Fatal(err)
		}
		if _, err = h.Page(context.Background(), BlockRange{1, 256}, false); err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
	b.ReportMetric(float64(calls.Load())/float64(b.N), "http/page")
	b.ReportMetric(float64(rows.Load())/float64(b.N), "rpc/page")
	b.ReportMetric(float64(statuses.Load())/float64(b.N), "status/page")
}
