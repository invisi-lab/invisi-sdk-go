package ipp

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/vmihailenco/msgpack/v5"
)

func TestCurrentAuthorizeWire(t *testing.T) {
	var body map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/ipp/v1/authorize" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		io.WriteString(w, `{"status":"accepted","nonce":"1","credit_commitment":"2","credit_shared_key_hash":"3","credit_shared_key":"4"}`)
	}))
	defer srv.Close()
	got, err := NewIPPClient(srv.URL).Authorize(AuthorizeRequest{SpenderIdentityCommitment: "spender", OwnerIdentityCommitment: "owner", Amount: "10", Expiry: "20"})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]interface{}{"spender_identity_commitment": "spender", "owner_identity_commitment": "owner", "amount": "10", "expiry": "20"}
	if !reflect.DeepEqual(body, want) {
		t.Fatalf("authorize wire = %#v", body)
	}
	if got.CreditSharedKeyHash == nil || *got.CreditSharedKeyHash != "3" {
		t.Fatalf("missing credit_shared_key_hash: %+v", got)
	}
}

func TestCurrentRequestRequiredFields(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls++; io.WriteString(w, `{}`) }))
	defer srv.Close()
	c := NewIPPClient(srv.URL)
	for _, req := range []AuthorizeRequest{
		{Amount: "1", Expiry: "2", OwnerIdentityCommitment: "owner"},
		{Amount: "1", Expiry: "2", SpenderIdentityCommitment: "spender"},
	} {
		if _, err := c.Authorize(req); err == nil {
			t.Error("authorization without both identities accepted")
		}
	}
	for _, req := range []WithdrawRequest{
		{Operator: "o", TokenAddress: "t", TreeID: "0", NoteCommitments: []string{"1"}},
		{Network: "eip155:1", Operator: "o", TokenAddress: "t", TreeID: "0"},
		{Network: "eip155:1", Operator: "o", TokenAddress: "t", TreeID: "0", NoteCommitments: []string{""}},
	} {
		if _, err := c.Withdraw(req); err == nil {
			t.Error("incomplete withdraw accepted")
		}
	}
	if calls != 0 {
		t.Fatalf("invalid requests reached transport %d times", calls)
	}
}

func TestCurrentWireRejectsRetiredFields(t *testing.T) {
	cases := []struct {
		name, body string
		dst        interface{}
	}{
		{"authorize old identity", `{"identity_commitment":"old","spender_identity_commitment":"s","owner_identity_commitment":"o","amount":"1","expiry":"2"}`, &AuthorizeRequest{}},
		{"authorize future tree", `{"status":"rejected","tree_id":"1"}`, &AuthorizeResponse{}},
		{"authorize old hash", `{"status":"rejected","nullifier_key_hash":"1"}`, &AuthorizeResponse{}},
		{"authorize misspelled hash", `{"status":"rejected","credit_shared_keyhash":"1"}`, &AuthorizeResponse{}},
		{"propose future tree", `{"status":"rejected","tree_id":"1"}`, &ProposeResponse{}},
		{"withdraw advisory identity", `{"identity_commitment":"old"}`, &WithdrawRequest{}},
		{"withdraw circuit hint", `{"status":"rejected","circuit":"settle"}`, &WithdrawResponse{}},
		{"withdraw private nullifier", `{"merkle_root":"1","total_amount":"2","nullifiers":["3"]}`, &WithdrawWitness{}},
		{"withdraw wrong proof shape", `{"leaf_index":"0","siblings":[]}`, &MerkleMembershipProof{}},
		{"relay sync async error field", `{"status":"queued","error_code":"relay_submit_failed"}`, &RelayResponse{}},
		{"relay signature mode", `{"mode":"signature","operation":"shield"}`, &RelayRequest{}},
		{"relay signature field", `{"mode":"calldata","operation":"shield","network":"eip155:1","protocol_address":"0x1234567890123456789012345678901234567890","calldata":"0x12","signature":"0x12"}`, &RelayRequest{}},
		{"relay old witness_json", `{"mode":"prove","operation":"create-credit","circuit":"create_credit","witness_json":"{}"}`, &RelayRequest{}},
		{"relay prove network even empty", `{"mode":"prove","operation":"create-credit","circuit":"create_credit","witness":{},"network":""}`, &RelayRequest{}},
		{"relay calldata extra circuit even empty", `{"mode":"calldata","operation":"shield","network":"eip155:1","protocol_address":"0x1234567890123456789012345678901234567890","calldata":"0x12","circuit":""}`, &RelayRequest{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := json.Unmarshal([]byte(tc.body), tc.dst); err == nil {
				t.Fatalf("accepted invalid wire: %s", tc.body)
			}
		})
	}
}

func TestCurrentVoucherSecretWire(t *testing.T) {
	voucher := sampleVoucher()
	encoded, err := EncodeVoucher(voucher, FormatJSON)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(encoded, `"note_secret"`) || strings.Contains(encoded, `"nullifier_key"`) {
		t.Fatal(encoded)
	}
	for _, format := range []Format{FormatJSON, FormatCompact} {
		raw := strings.Replace(encoded, `"note_secret":`, `"nullifier_key":`, 1)
		var message map[string]interface{}
		if err := json.Unmarshal([]byte(raw), &message); err != nil {
			t.Fatal(err)
		}
		// Even a current secret alongside the retired name cannot rescue the payload.
		for _, both := range []bool{false, true} {
			payload := message["payload"].(map[string]interface{})
			if both {
				payload["note_secret"] = voucher.Payload.NoteSecret
			}
			if format == FormatJSON {
				b, _ := json.Marshal(message)
				raw = string(b)
			} else {
				b, err := msgpack.Marshal(message)
				if err != nil {
					t.Fatal(err)
				}
				raw = "ipp1:" + base64.RawURLEncoding.EncodeToString(b)
			}
			if _, err := DecodeVoucher(raw); err == nil {
				t.Fatalf("accepted retired voucher secret in %s (both=%v)", format, both)
			}
		}
	}
}

func TestCurrentAcceptedResponsesValidated(t *testing.T) {
	cases := []struct{ path, body string }{
		{"propose", `{"status":"accepted","nonce":"1"}`},
		{"propose", `{"status":"pending"}`},
		{"propose", `{"status":"rejected","tree_id":"1"}`},
		{"authorize", `{"status":"accepted","nonce":"1","credit_commitment":"2","credit_shared_key":"3"}`},
		{"authorize", `{"status":"accepted","nonce":"1","credit_commitment":"2","credit_shared_key_hash":"3","credit_shared_key":"4","nullifier_key_hash":"5"}`},
		{"withdraw", `{"status":"accepted"}`},
		{"withdraw", `{"status":"accepted","witness":{"merkle_root":"1","total_amount":"2","merkle_proofs":[]}}`},
		{"withdraw", `{"status":"accepted","witness":{"merkle_root":"1","total_amount":"2","merkle_proofs":[{}]}}`},
		{"withdraw", `{"status":"accepted","witness":{"merkle_root":"1","total_amount":"2","merkle_proofs":[{"leaf_index":"0","merkle_path":["1"]}]}}`},
	}
	for i, tc := range cases {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { io.WriteString(w, tc.body) }))
			defer srv.Close()
			c := NewIPPClient(srv.URL)
			var err error
			switch tc.path {
			case "propose":
				_, err = c.Propose(ProposeRequest{IdentityCommitment: "1", Amount: "2"})
			case "authorize":
				_, err = c.Authorize(AuthorizeRequest{SpenderIdentityCommitment: "1", OwnerIdentityCommitment: "2", Amount: "3", Expiry: "4"})
			case "withdraw":
				_, err = c.Withdraw(WithdrawRequest{Network: "eip155:1", Operator: "1", TokenAddress: "2", TreeID: "0", NoteCommitments: []string{"3"}})
			}
			if err == nil {
				t.Fatalf("accepted invalid %s response %s", tc.path, tc.body)
			}
		})
	}
}

func TestCurrentWithdrawPublicProofWire(t *testing.T) {
	for _, length := range []int{15, 16, 17} {
		t.Run(fmt.Sprint(length), func(t *testing.T) {
			path := make([]string, length)
			for i := range path {
				path[i] = "1"
			}
			body, _ := json.Marshal(map[string]interface{}{"status": "accepted", "witness": map[string]interface{}{"merkle_root": "root", "total_amount": "999999", "merkle_proofs": []interface{}{map[string]interface{}{"leaf_index": "0", "merkle_path": path}}}})
			var request map[string]interface{}
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { json.NewDecoder(r.Body).Decode(&request); w.Write(body) }))
			defer srv.Close()
			got, err := NewIPPClient(srv.URL).Withdraw(WithdrawRequest{Network: "eip155:1", Operator: "o", TokenAddress: "t", TreeID: "0", NoteCommitments: []string{"n"}})
			if length != 16 {
				if err == nil {
					t.Fatal("accepted non-depth-16 proof")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got.Witness.TotalAmount != "999999" || got.Witness.MerkleProofs[0].LeafIndex != "0" {
				t.Fatalf("public proof lost: %+v", got)
			}
			if len(request) != 5 || request["network"] != "eip155:1" || request["operator"] != "o" || request["token_address"] != "t" {
				t.Fatalf("wrong public routing wire: %#v", request)
			}
		})
	}
}

func TestCurrentRelayOperationsAndProofCircuits(t *testing.T) {
	const address = "0x1234567890123456789012345678901234567890"
	circuits := map[RelayOperation][]string{
		RelayTransact:     {"transact_user_slots14", "transact_user_slots56", "transact_user_slots112", "transact_user_slots168", "transact_aggregate_14_4", "transact_aggregate_14_8", "transact_aggregate_14_12"},
		RelayCreateCredit: {"create_credit"}, RelayCreditDraw: {"credit_draw_slots1", "credit_draw_slots12", "credit_draw_slots50", "credit_draw_slots120"}, RelayCancelCredit: {"credit_cancel"}, RelayTopUpCredit: {"credit_top_up"},
	}
	for _, op := range []RelayOperation{RelayShield, RelayTransact, RelayCreateCredit, RelayCreditDraw, RelayCancelCredit, RelayTopUpCredit} {
		if err := validateRelayRequest(RelayRequest{Mode: "calldata", Operation: op, Network: "eip155:1", ProtocolAddress: address, Calldata: "0x1234"}); err != nil {
			t.Fatalf("%s: %v", op, err)
		}
	}
	for op, names := range circuits {
		for _, circuit := range names {
			for _, mode := range []string{"witness", "prove"} {
				req := RelayRequest{Mode: mode, Operation: op, Circuit: circuit, Witness: map[string]interface{}{}}
				if mode == "witness" {
					req.Network = "eip155:1"
				}
				if err := validateRelayRequest(req); err != nil {
					t.Fatalf("%s/%s/%s: %v", mode, op, circuit, err)
				}
			}
		}
	}
	invalid := []RelayRequest{
		{Mode: "calldata", Operation: "pay-open", Network: "eip155:1", ProtocolAddress: address, Calldata: "0x12"},
		{Mode: "calldata", Operation: "pay-blind", Network: "eip155:1", ProtocolAddress: address, Calldata: "0x12"},
		{Mode: "calldata", Operation: "credit-reclaim", Network: "eip155:1", ProtocolAddress: address, Calldata: "0x12"},
		{Mode: "signature", Operation: RelayShield},
		{Mode: "calldata", Operation: RelayShield, Network: "eip155:1", Calldata: "0x12"},
		{Mode: "calldata", Operation: RelayShield, ProtocolAddress: address, Calldata: "0x12"},
		{Mode: "calldata", Operation: RelayShield, Network: "eip155:1", ProtocolAddress: "0x" + strings.Repeat("0", 40), Calldata: "0x12"},
		{Mode: "calldata", Operation: RelayShield, Network: "eip155:1", ProtocolAddress: "0x12", Calldata: "0x12"},
		{Mode: "calldata", Operation: RelayShield, Network: "eip155:1", ProtocolAddress: address, Calldata: "0x"},
		{Mode: "calldata", Operation: RelayShield, Network: "eip155:1", ProtocolAddress: address, Calldata: "0x123"},
		{Mode: "prove", Operation: RelayTransact, Circuit: "settle", Witness: map[string]interface{}{}},
		{Mode: "prove", Operation: RelayCreateCredit, Circuit: "credit_draw_slots1", Witness: map[string]interface{}{}},
		{Mode: "prove", Operation: RelayShield, Circuit: "shield", Witness: map[string]interface{}{}},
		{Mode: "prove", Operation: RelayCreateCredit, Circuit: "create_credit", Witness: []interface{}{}},
		{Mode: "prove", Operation: RelayCreateCredit, Circuit: "create_credit", Witness: "{}"},
		{Mode: "prove", Operation: RelayCreateCredit, Circuit: "create_credit"},
	}
	for i, req := range invalid {
		if err := validateRelayRequest(req); err == nil {
			t.Fatalf("accepted invalid relay %d: %+v", i, req)
		}
	}
}

func TestCurrentRelayWireAndProveResult(t *testing.T) {
	for _, mode := range []string{"calldata", "witness", "prove"} {
		t.Run(mode, func(t *testing.T) {
			var body map[string]interface{}
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				json.NewDecoder(r.Body).Decode(&body)
				io.WriteString(w, `{"status":"completed","proof":"proof","calldata":"0x12"}`)
			}))
			defer srv.Close()
			req := RelayRequest{Mode: mode, Operation: RelayCreateCredit}
			want := map[string]interface{}{"mode": mode, "operation": "create-credit"}
			if mode != "prove" {
				req.Network = "eip155:1"
				want["network"] = req.Network
			}
			if mode == "calldata" {
				req.ProtocolAddress = "0x1234567890123456789012345678901234567890"
				req.Calldata = "0x12"
				want["protocol_address"] = req.ProtocolAddress
				want["calldata"] = req.Calldata
			} else {
				req.Circuit = "create_credit"
				req.Witness = map[string]interface{}{"credit_shared_key": "7"}
				want["circuit"] = req.Circuit
				want["witness"] = req.Witness
			}
			got, err := NewIPPClient(srv.URL).Relay(req)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(body, want) {
				t.Fatalf("wire got %#v want %#v", body, want)
			}
			if got.Proof == nil || *got.Proof != "proof" || got.Calldata == nil {
				t.Fatalf("proof result lost: %+v", got)
			}
		})
	}
}
