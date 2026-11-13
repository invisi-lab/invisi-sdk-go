package session

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
	"github.com/invisi-lab/invisi-sdk-go/rpc/broadcaster"
)

type rejectQuotedCreditStore struct{ Store }

func (s rejectQuotedCreditStore) Update(ctx context.Context, update func(*State) error) error {
	return s.Store.Update(ctx, func(state *State) error {
		if err := update(state); err != nil {
			return err
		}
		for _, record := range state.Operations {
			if record.Intent.Broadcaster != nil {
				return errors.New("quote persistence unavailable")
			}
		}
		return nil
	})
}

func TestClientQuoteManagedCreditRollbackAndResume(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	managed, err := New(Options{Store: store})
	if err != nil {
		t.Fatal(err)
	}
	prepared, grant := testPlan(t)
	if err = managed.Track(ctx, grant); err != nil {
		t.Fatal(err)
	}
	original, err := managed.Prepare(ctx, "quoted-credit", prepared)
	if err != nil {
		t.Fatal(err)
	}
	intent, err := prepared.ExportIntent()
	if err != nil {
		t.Fatal(err)
	}
	key, err := crypto.ToECDSA(common.LeftPadBytes([]byte{9}, 32))
	if err != nil {
		t.Fatal(err)
	}
	collector := common.HexToAddress("0x1234")
	calls := 0
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		var request struct {
			ID     uint64
			Method string
			Params []struct{ Data string }
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatal(err)
		}
		var result any
		switch request.Method {
		case "broadcaster_estimateQuote":
			result = broadcaster.Fee{FeeRecipient: collector, FeeAmount: "1", MaxGas: "2000000"}
		case "broadcaster_getQuote":
			data := common.FromHex(request.Params[0].Data)
			decoded, err := broadcaster.DecodeIntent(data)
			if err != nil {
				t.Fatal(err)
			}
			hash, err := broadcaster.RequestHash(data)
			if err != nil {
				t.Fatal(err)
			}
			quote := broadcaster.Quote{Token: decoded.Token, FeeRecipient: collector, FeeAmount: decoded.FeeAmount.String(), Deadline: decoded.Deadline.String(), MaxGas: "2000000", RequestHash: hash, Execution: "direct"}
			trust := broadcaster.Trust{ChainID: primitives.ElementToBigInt(intent.Context.ChainID), Protocol: common.BigToAddress(primitives.ElementToBigInt(intent.Context.ProtocolAddress)), Signer: crypto.PubkeyToAddress(key.PublicKey)}
			typed, err := broadcaster.QuoteTypedData(quote, trust)
			if err != nil {
				t.Fatal(err)
			}
			digest, _, err := apitypes.TypedDataAndHash(typed)
			if err != nil {
				t.Fatal(err)
			}
			signature, err := crypto.Sign(digest, key)
			if err != nil {
				t.Fatal(err)
			}
			result = broadcaster.SignedQuote{Quote: quote, Signature: "0x" + common.Bytes2Hex(signature)}
		default:
			t.Fatalf("unexpected method %s", request.Method)
		}
		json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": request.ID, "result": result})
	}))
	defer server.Close()
	client, err := broadcaster.NewClient(broadcaster.Configuration{Endpoint: server.URL, QuoteSigner: crypto.PubkeyToAddress(key.PublicKey), FeeRecipient: collector}, broadcaster.Options{HTTPClient: server.Client()})
	if err != nil {
		t.Fatal(err)
	}
	options := broadcaster.QuoteOptions{MaxFeeAmount: big.NewInt(2)}
	managed.store = rejectQuotedCreditStore{store}
	if err = client.Quote(ctx, original, options); err == nil {
		t.Fatal("quote succeeded without durable state")
	}
	current, err := original.prepared.ExportIntent()
	if err != nil || current.Broadcaster != nil {
		t.Fatal("failed persistence changed prepared state", err)
	}
	snapshot, err := store.ReadSnapshot(ctx)
	if err != nil || snapshot.Operations[original.id].Intent.Broadcaster != nil || snapshot.Grants[grant.key()].ReservedBy != original.id {
		t.Fatal("failed persistence changed journal or reservations", err)
	}
	managed.store = store
	resumed, err := managed.Resume(ctx, original.id, nil)
	if err != nil {
		t.Fatal(err)
	}
	before := calls
	if err = client.Quote(ctx, original, options); err == nil || calls != before {
		t.Fatal("stale worker reached quote", err)
	}
	if err = client.Quote(ctx, resumed, options); err != nil {
		t.Fatal(err)
	}
	quote, err := resumed.ExportBroadcasterQuote(ctx)
	if err != nil || quote.Quote.FeeAmount != "1" {
		t.Fatal("resumed worker missing quote", err)
	}
	snapshot, err = store.ReadSnapshot(ctx)
	if err != nil || snapshot.Operations[resumed.id].Intent.Broadcaster == nil || snapshot.Grants[grant.key()].ReservedBy != resumed.id {
		t.Fatal("quote did not preserve managed ownership", err)
	}
}
