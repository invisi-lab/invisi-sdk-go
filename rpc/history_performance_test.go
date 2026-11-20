package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func batchHeaderFixture(t testing.TB, mode string, delay time.Duration) (*TrustedProvider, *atomic.Int64) {
	t.Helper()
	d := ProtocolDeployment{ChainID: "1", Protocol: ContractDeployment{Address: common.HexToAddress("1"), RuntimeCodeHash: common.HexToHash("2")}}
	var requests atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		if delay > 0 {
			time.Sleep(delay)
		}
		var body json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
			return
		}
		var rows []struct {
			ID     uint64
			Method string
			Params []trustedRequest
		}
		batch := len(body) > 0 && body[0] == '['
		if batch {
			if err := json.Unmarshal(body, &rows); err != nil {
				t.Error(err)
				return
			}
		} else {
			rows = make([]struct {
				ID     uint64
				Method string
				Params []trustedRequest
			}, 1)
			json.Unmarshal(body, &rows[0])
		}
		if len(rows) > 20 {
			t.Errorf("batch exceeds server limit: %d", len(rows))
		}
		responses := []map[string]any{}
		for _, q := range rows {
			n, _ := new(big.Int).SetString(q.Params[0].Block[2:], 16)
			h := &types.Header{Number: n, Difficulty: big.NewInt(0)}
			out := trustedResponse{APIVersion: "v1", ProtocolVersion: "1", Deployment: d, Header: h, HeaderHash: h.Hash()}
			response := map[string]any{"jsonrpc": "2.0", "id": q.ID, "result": out}
			switch mode {
			case "hash":
				out.HeaderHash = common.HexToHash("99")
				response["result"] = out
			case "pins":
				out.Deployment.Protocol.RuntimeCodeHash = common.HexToHash("99")
				response["result"] = out
			case "number":
				out.Header = types.CopyHeader(h)
				out.Header.Number = big.NewInt(999)
				out.HeaderHash = out.Header.Hash()
				response["result"] = out
			case "error":
				delete(response, "result")
				response["error"] = map[string]any{"code": -32000, "message": "unavailable"}
			case "both":
				response["error"] = map[string]any{"code": -32000, "message": "unavailable"}
			case "both-null":
				response["error"] = nil
			case "null-error":
				delete(response, "result")
				response["error"] = nil
			case "malformed-error":
				delete(response, "result")
				response["error"] = map[string]any{"code": -32000}
			case "version":
				response["jsonrpc"] = "1.0"
			}
			responses = append(responses, response)
		}
		if batch {
			if mode == "missing" {
				responses = responses[:len(responses)-1]
			}
			if mode == "duplicate" && len(responses) > 1 {
				responses[1]["id"] = responses[0]["id"]
			}
			if mode == "unknown" && len(responses) > 0 {
				responses[0]["id"] = uint64(1 << 60)
			}
			// Responses are deliberately in reverse order; IDs are authoritative.
			for i, j := 0, len(responses)-1; i < j; i, j = i+1, j-1 {
				responses[i], responses[j] = responses[j], responses[i]
			}
			json.NewEncoder(w).Encode(responses)
		} else {
			json.NewEncoder(w).Encode(responses[0])
		}
	}))
	t.Cleanup(server.Close)
	c, err := NewProtocolRpcClient(server.URL+"/v1", ClientOptions{Deployment: &d})
	if err != nil {
		t.Fatal(err)
	}
	p, err := NewTrustedProvider(c, DeploymentContext{1, d.Protocol.Address})
	if err != nil {
		t.Fatal(err)
	}
	return p, &requests
}

func TestTrustedProviderBatchedHeaders(t *testing.T) {
	for _, mode := range []string{"ok", "missing", "duplicate", "unknown", "hash", "pins", "number", "error", "both", "both-null", "null-error", "malformed-error", "version"} {
		t.Run(mode, func(t *testing.T) {
			p, calls := batchHeaderFixture(t, mode, 0)
			numbers := make([]*big.Int, 45)
			for i := range numbers {
				numbers[i] = big.NewInt(int64(i + 1))
			}
			headers, err := p.HeadersByNumber(context.Background(), numbers)
			if mode != "ok" {
				if mode == "error" {
					var remote *ProtocolRpcError
					if !errors.As(err, &remote) || remote.Code != -32000 {
						t.Fatalf("lost typed remote error: %v", err)
					}
				}
				if err == nil || headers != nil {
					t.Fatal("invalid batch accepted", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(headers) != len(numbers) || calls.Load() != 3 {
				t.Fatalf("headers=%d requests=%d", len(headers), calls.Load())
			}
			for i, h := range headers {
				if h.Number.Cmp(numbers[i]) != 0 {
					t.Fatal("response order mismatch")
				}
			}
			for _, n := range []*big.Int{nil, big.NewInt(-1)} {
				if _, err := p.HeadersByNumber(context.Background(), []*big.Int{n}); err == nil {
					t.Fatal("mutable/invalid block allowed")
				}
			}
		})
	}
}

func BenchmarkTrustedHistoryHeaders(b *testing.B) {
	for _, batched := range []bool{false, true} {
		b.Run(fmt.Sprintf("batch=%t", batched), func(b *testing.B) {
			p, calls := batchHeaderFixture(b, "ok", time.Millisecond)
			numbers := make([]*big.Int, 256)
			for i := range numbers {
				numbers[i] = big.NewInt(int64(i + 1))
			}
			b.ResetTimer()
			for range b.N {
				if batched {
					if _, err := p.HeadersByNumber(context.Background(), numbers); err != nil {
						b.Fatal(err)
					}
				} else {
					for _, n := range numbers {
						if _, err := p.HeaderByNumber(context.Background(), n); err != nil {
							b.Fatal(err)
						}
					}
				}
			}
			b.StopTimer()
			b.ReportMetric(float64(calls.Load())/float64(b.N), "http/page")
		})
	}
}
