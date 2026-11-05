package session

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	sdk "github.com/invisi-lab/invisi-sdk-go/sdk"

	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
)

type checkpointProvider struct {
	sessionProvider
	middle *types.Header
}

func (p checkpointProvider) HeaderByNumber(ctx context.Context, n *big.Int) (*types.Header, error) {
	if n != nil && n.Uint64() == 5 {
		return p.middle, nil
	}
	return p.sessionProvider.HeaderByNumber(ctx, n)
}
func refreshFixture(t *testing.T, failSecond bool, entered, release chan struct{}) (*WalletSession, *atomic.Int32) {
	t.Helper()
	s, _, _ := testSession(t)
	base := s.options.Provider.(sessionProvider)
	middle := types.CopyHeader(base.head)
	middle.Number = big.NewInt(5)
	provider := checkpointProvider{base, middle}
	deployment := s.protocol.Deployment()
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			ID     uint64
			Method string
			Params []json.RawMessage
		}
		_ = json.NewDecoder(r.Body).Decode(&request)
		block := protocolrpc.CanonicalBlock{Number: 10, Hash: base.head.Hash(), ParentHash: base.head.ParentHash, Timestamp: base.head.Time}
		var result any
		switch request.Method {
		case "invisi_getStatus":
			result = map[string]any{"apiVersion": "v1", "protocolVersion": "1", "deployment": deployment, "indexedHead": block, "checkedBits": 8, "detectionEnabled": true}
		case "invisi_getUpdates":
			count := calls.Add(1)
			if entered != nil && count == 1 {
				close(entered)
				<-release
			}
			if failSecond && count == 2 {
				_ = json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": request.ID, "error": map[string]any{"code": -32000, "message": "incomplete scan"}})
				return
			}
			var selected struct {
				ToBlock uint64 `json:"toBlock"`
			}
			_ = json.Unmarshal(request.Params[0], &selected)
			if selected.ToBlock == 5 {
				block = protocolrpc.CanonicalBlock{Number: 5, Hash: middle.Hash(), ParentHash: middle.ParentHash, Timestamp: middle.Time}
			}
			result = map[string]any{"blockHash": block.Hash, "outputs": []any{}, "creditTransitions": []any{}}
		default:
			t.Errorf("unexpected method %s", request.Method)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": request.ID, "result": result})
	}))
	t.Cleanup(server.Close)
	client, err := sdk.NewProtocolRpcClient(server.URL+"/v1", protocolrpc.ClientOptions{Deployment: &deployment})
	if err != nil {
		t.Fatal(err)
	}
	s.options.Provider = provider
	s.rpc = client.WithProvider(provider)
	return s, &calls
}
func TestRefreshCheckpointRequiresCompleteContiguousScan(t *testing.T) {
	for _, kind := range []string{"complete", "partial", "failed"} {
		t.Run(kind, func(t *testing.T) {
			s, _ := refreshFixture(t, kind == "failed", nil, nil)
			from := uint64(1)
			if kind == "partial" {
				from = 10
			}
			_, err := s.Refresh(context.Background(), RefreshOptions{FromBlock: from, PageSize: 5})
			if (err != nil) != (kind == "failed") {
				t.Fatal(err)
			}
			state, err := s.options.Store.ReadSnapshot(context.Background(), s.scope)
			if err != nil {
				t.Fatal(err)
			}
			if kind == "complete" {
				if state.DiscoveryThrough == nil || state.DiscoveryThrough.Number != 10 {
					t.Fatal("complete scan did not advance discovery checkpoint")
				}
			} else if state.DiscoveryThrough != nil {
				t.Fatal("partial or failed scan claimed complete discovery")
			}
		})
	}
}
func TestRefreshSerializesAndCancelsQueuedWork(t *testing.T) {
	entered, release := make(chan struct{}), make(chan struct{})
	s, calls := refreshFixture(t, false, entered, release)
	first := make(chan error, 1)
	go func() { _, err := s.Refresh(context.Background(), RefreshOptions{FromBlock: 1}); first <- err }()
	<-entered
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if _, err := s.Refresh(ctx, RefreshOptions{FromBlock: 1}); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal("queued canceled refresh did not return", err)
	}
	if calls.Load() != 1 {
		t.Fatal("overlapping refresh entered RPC")
	}
	close(release)
	if err := <-first; err != nil {
		t.Fatal(err)
	}
}

type pauseSpentProvider struct {
	sdk.Provider
	once             sync.Once
	entered, release chan struct{}
}

func (p *pauseSpentProvider) CallContract(ctx context.Context, message ethereum.CallMsg, block *big.Int) ([]byte, error) {
	parsed, err := protocol.ProtocolABI()
	if err != nil {
		return nil, err
	}
	method, err := parsed.MethodById(message.Data[:4])
	if err == nil && method.RawName == "isNullifierSpent" {
		p.once.Do(func() { close(p.entered); <-p.release })
	}
	return p.Provider.CallContract(ctx, message, block)
}
func TestRefreshCannotCrossAnotherSessionsReset(t *testing.T) {
	for _, failedReset := range []bool{false, true} {
		t.Run(fmt.Sprint("failedReset=", failedReset), func(t *testing.T) {
			s, _ := refreshFixture(t, failedReset, nil, nil)
			entered, release := make(chan struct{}), make(chan struct{})
			provider := &pauseSpentProvider{Provider: s.options.Provider, entered: entered, release: release}
			s.options.Provider = provider
			s.rpc = s.rpc.WithProvider(provider)
			other := *s
			other.refreshGate = make(chan struct{}, 1)
			oldScan := make(chan error, 1)
			go func() { _, err := s.Refresh(context.Background(), RefreshOptions{FromBlock: 1}); oldScan <- err }()
			<-entered
			_, resetErr := other.Refresh(context.Background(), RefreshOptions{FromBlock: 1, Reset: true})
			if (resetErr != nil) != failedReset {
				close(release)
				t.Fatal("unexpected reset result", resetErr)
			}
			close(release)
			if err := <-oldScan; err == nil {
				t.Error("old scan committed across another session reset")
			}
			snapshot, err := s.options.Store.ReadSnapshot(context.Background(), s.scope)
			if err != nil {
				t.Fatal(err)
			}
			if len(snapshot.Notes) != 0 {
				t.Error("old scan resurrected cleared inventory")
			}
			if snapshot.Invalidated != failedReset {
				t.Error("old scan overwrote reset validity")
			}
			if failedReset && snapshot.DiscoveryThrough != nil {
				t.Error("failed reset claimed discovery coverage")
			}
		})
	}
}

type pauseFirstCommitStore struct {
	WalletStore
	paused           atomic.Bool
	entered, release chan struct{}
}

func (s *pauseFirstCommitStore) Update(ctx context.Context, scope WalletScope, apply func(*WalletState) error) error {
	if err := s.WalletStore.Update(ctx, scope, apply); err != nil {
		return err
	}
	if s.paused.CompareAndSwap(false, true) {
		close(s.entered)
		<-s.release
	}
	return nil
}
func TestDiscoveryCheckpointCannotCrossResetAfterLastPageCommit(t *testing.T) {
	session, _ := refreshFixture(t, true, nil, nil)
	entered, release := make(chan struct{}), make(chan struct{})
	store := &pauseFirstCommitStore{WalletStore: session.options.Store, entered: entered, release: release}
	session.options.Store = store
	other := *session
	other.refreshGate = make(chan struct{}, 1)
	result := make(chan error, 1)
	go func() { _, err := session.Refresh(context.Background(), RefreshOptions{FromBlock: 1}); result <- err }()
	<-entered
	if _, err := other.Refresh(context.Background(), RefreshOptions{FromBlock: 1, Reset: true}); err == nil {
		close(release)
		t.Fatal("reset fixture should fail after clearing inventory")
	}
	close(release)
	if err := <-result; err == nil {
		t.Fatal("old scan finalized discovery after reset")
	}
	state, err := store.ReadSnapshot(context.Background(), session.scope)
	if err != nil {
		t.Fatal(err)
	}
	if !state.Invalidated || state.DiscoveryThrough != nil || len(state.Notes) != 0 {
		t.Fatal("old scan overwrote failed reset")
	}
}
