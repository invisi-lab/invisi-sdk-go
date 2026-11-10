package provenance

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestTrackingScopedWireAndTerminalStates(t *testing.T) {
	for _, status := range []string{"recognized", "expired", "superseded", "stale_generation", "unknown"} {
		t.Run(status, func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				var q struct {
					ID     any                 `json:"id"`
					Method string              `json:"method"`
					Params []map[string]string `json:"params"`
				}
				json.NewDecoder(r.Body).Decode(&q)

				if q.Method != "provenance_getTracking" || len(q.Params) != 1 || len(q.Params[0]) != 1 || q.Params[0]["trackingId"] != "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" {
					t.Error("wrong tracking query")
				}
				digest, _ := word("9")
				tracking := map[string]any{"id": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "status": status, "expiresAt": "1", "operationDigest": digest}
				if status == "recognized" {
					tx, _ := word("1")
					block, _ := word("2")
					tracking["matched"] = map[string]string{"txHash": tx, "blockHash": block, "operationIndex": "0"}
				}
				json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": q.ID, "result": map[string]any{"scope": Scope{ChainID: "1", ProtocolAddress: "2"}, "chainGeneration": "3", "providerId": "4", "signingKeyId": "5", "manifestSha256": strings.Repeat("a", 64), "tracking": tracking}})
			}))
			defer server.Close()
			id, _ := word("5")
			c := Client{Endpoint: server.URL, Trust: TrustContext{Scope: Scope{ChainID: "1", ProtocolAddress: "2"}, ChainGeneration: "3", ProviderID: "4", ManifestSHA256: strings.Repeat("a", 64), SigningKeys: map[string]ed25519.PublicKey{id: make([]byte, 32)}}}
			r, e := c.WaitTracking(context.Background(), "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", time.Millisecond, 20*time.Millisecond)
			if status == "unknown" || status == "stale_generation" {
				if e == nil {
					t.Fatal("unsafe tracking accepted")
				}
			} else if e != nil {
				t.Fatal(e)
			} else if tracking, e := TrackingFromResult(r, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"); e != nil || tracking.Status != status {
				t.Fatal(e)
			}
			if calls != 1 {
				t.Fatal("terminal tracking polled again")
			}
		})
	}
}

func TestPreflightTrackingRejectsTerminalAcknowledgement(t *testing.T) {
	for _, status := range []string{"expired", "superseded"} {
		digest, _ := word("9")
		raw, _ := json.Marshal(Tracking{ID: strings.Repeat("a", 48), Status: status, ExpiresAt: "1", OperationDigest: digest})
		d, _ := json.Marshal(digest)
		r := Result{"broadcastDecision": json.RawMessage(`"allow"`), "operationDigest": d, "tracking": raw}
		if _, e := PreflightTracking(r, "0", digest); e == nil {
			t.Fatal("terminal task permitted broadcast")
		}
	}
}
