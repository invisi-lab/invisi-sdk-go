package rpc

import (
	"context"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func TestHistoryUpdatesShareOnlyVerifiedStatusAndKeepCanonicalFence(t *testing.T) {
	for _, reorg := range []bool{false, true} {
		t.Run(map[bool]string{false: "stable", true: "reorg"}[reorg], func(t *testing.T) {
			p := &testProvider{header: &types.Header{Number: big.NewInt(10), Time: 123}}
			block := CanonicalBlock{Number: 10, Hash: p.header.Hash(), Timestamp: 123}
			d := ProtocolDeployment{ChainID: "31337", Protocol: ContractDeployment{Address: common.HexToAddress("1"), RuntimeCodeHash: common.HexToHash("2"), FromBlock: 1}}
			statuses := 0
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var q struct {
					ID     uint64
					Method string
				}
				json.NewDecoder(r.Body).Decode(&q)
				var result any
				switch q.Method {
				case "invisi_getStatus":
					statuses++
					result = map[string]any{"apiVersion": "v1", "protocolVersion": "1", "deployment": d, "indexedHead": block}
				case "invisi_getNoteUpdates":
					result = map[string]any{"blockHash": block.Hash, "notes": []any{}}
				case "invisi_getCreditTransitions":
					result = []any{}
					if reorg {
						p.header = types.CopyHeader(p.header)
						p.header.Extra = []byte{1}
					}
				default:
					t.Errorf("unexpected method %s", q.Method)
				}
				json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": q.ID, "result": result})
			}))
			defer srv.Close()
			c, _ := NewProtocolRpcClient(srv.URL+"/v1", ClientOptions{Deployment: &d})
			c = c.WithProvider(p)
			notes, credits, err := c.GetHistoryUpdates(context.Background(), DeploymentContext{31337, d.Protocol.Address}, ProtocolUpdateRange{1, 10})
			if reorg {
				if err == nil {
					t.Fatal("canonical reorg accepted")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if notes.Block != block || credits.Block != block || statuses != 1 {
				t.Fatalf("blocks or status calls incorrect: %d", statuses)
			}
			// A second operation must observe status again; this is not a global cache.
			if _, _, err = c.GetHistoryUpdates(context.Background(), DeploymentContext{31337, d.Protocol.Address}, ProtocolUpdateRange{1, 10}); err != nil || statuses != 2 {
				t.Fatalf("status escaped operation boundary: %d %v", statuses, err)
			}
		})
	}
}

type forkBetweenHistoryStreams struct {
	testProvider
	lowerReads         int
	lowerOld, lowerNew *types.Header
}

func (p *forkBetweenHistoryStreams) HeaderByNumber(_ context.Context, n *big.Int) (*types.Header, error) {
	if n != nil && n.Uint64() == 9 {
		p.lowerReads++
		if p.lowerReads > 2 {
			return p.lowerNew, nil
		}
		return p.lowerOld, nil
	}
	return p.header, nil
}
func TestHistoryUpdatesRejectIndividuallyCanonicalMixedForks(t *testing.T) {
	old := &types.Header{Number: big.NewInt(9), Time: 9}
	next := types.CopyHeader(old)
	next.Extra = []byte{1}
	top := &types.Header{Number: big.NewInt(10), Time: 10, ParentHash: old.Hash()}
	p := &forkBetweenHistoryStreams{testProvider: testProvider{header: top}, lowerOld: old, lowerNew: next}
	d := ProtocolDeployment{ChainID: "31337", Protocol: ContractDeployment{Address: common.HexToAddress("1"), RuntimeCodeHash: common.HexToHash("2"), FromBlock: 1}}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var q struct {
			ID     uint64
			Method string
		}
		json.NewDecoder(r.Body).Decode(&q)
		var result any
		switch q.Method {
		case "invisi_getStatus":
			result = map[string]any{"apiVersion": "v1", "protocolVersion": "1", "deployment": d, "indexedHead": CanonicalBlock{Number: 10, Hash: top.Hash(), ParentHash: top.ParentHash, Timestamp: 10}}
		case "invisi_getNoteUpdates":
			result = map[string]any{"blockHash": old.Hash(), "notes": []any{}}
		case "invisi_getCreditTransitions":
			result = []any{}
		default:
			t.Errorf("unexpected method %s", q.Method)
		}
		json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": q.ID, "result": result})
	}))
	defer srv.Close()
	c, _ := NewProtocolRpcClient(srv.URL+"/v1", ClientOptions{Deployment: &d})
	c = c.WithProvider(p)
	if _, _, err := c.GetHistoryUpdates(context.Background(), DeploymentContext{31337, d.Protocol.Address}, ProtocolUpdateRange{1, 9}); err == nil {
		t.Fatal("returned combined discovery streams from different canonical forks")
	}
}
