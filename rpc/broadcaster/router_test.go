package broadcaster

import (
	"context"
	"encoding/json"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestRouterEnvelopeGoldenVectors(t *testing.T) {
	b, e := os.ReadFile("testdata/payment-router-envelope-vectors.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Core     common.Address
		Envelope RouterEnvelope
		Hash     common.Hash
	}
	if e = json.Unmarshal(b, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 3 {
		t.Fatal("expected signed authorization vector")
	}
	for _, v := range rows {
		h, e := EnvelopeHash(big.NewInt(31337), v.Core, v.Envelope)
		if e != nil || h != v.Hash {
			t.Fatal(h, e)
		}
		v.Envelope.Value = "1"
		if _, e = EnvelopeHash(big.NewInt(31337), v.Core, v.Envelope); e == nil {
			t.Fatal("principal accepted")
		}
	}
}

func TestRouterQuoteBindsOuterAndSubmitWire(t *testing.T) {
	data, trust, key := intent(t)
	var rows []struct {
		Core     common.Address
		Envelope RouterEnvelope
	}
	raw, err := os.ReadFile("testdata/payment-router-envelope-vectors.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(raw, &rows); err != nil {
		t.Fatal(err)
	}
	env := rows[0].Envelope
	trust.ChainID = big.NewInt(31337)
	trust.Protocol = rows[0].Core
	args, err := unpackRelay("payTokenFor", env.Data)
	if err != nil {
		t.Fatal(err)
	}
	core := relayMethods["transact"]
	encoded, err := core.Inputs.Pack(args[0], args[1])
	if err != nil {
		t.Fatal(err)
	}
	data = append(core.ID, encoded...)
	s := signed(t, data, trust, key)
	decoded, err := DecodeIntent(data)
	if err != nil {
		t.Fatal(err)
	}
	s.Quote.FeeAmount = decoded.FeeAmount.String()
	h, e := EnvelopeHash(trust.ChainID, trust.Protocol, env)
	if e != nil {
		t.Fatal(e)
	}
	s.Quote.RequestHash = h
	s.RequestHashFormat = RouterRequestHashFormat
	td, e := QuoteTypedData(s.Quote, trust)
	if e != nil {
		t.Fatal(e)
	}
	digest, _, e := apitypes.TypedDataAndHash(td)
	if e != nil {
		t.Fatal(e)
	}
	sig, e := crypto.Sign(digest, key)
	if e != nil {
		t.Fatal(e)
	}
	s.Signature = hexutil.Encode(sig)
	s.Context = &trust
	if e = VerifyRouterQuote(s, trust, data, env, 0); e != nil {
		t.Fatal(e)
	}
	s.ProvenanceRequired = true
	if e = VerifyRouterQuote(s, trust, data, env, 0); e != nil {
		t.Fatal("capability metadata rejected", e)
	}
	s.ProvenanceRequired = false
	changed := append([]byte(nil), data...)
	changed[len(changed)-32] ^= 1
	if e = VerifyRouterQuote(s, trust, changed, env, 0); e == nil {
		t.Fatal("accepted different embedded Core payload")
	}
	wrong := env
	wrong.To = common.HexToAddress("0x55")
	if e = VerifyRouterQuote(s, trust, data, wrong, 0); e == nil {
		t.Fatal("target substitution accepted")
	}
	wrong = env
	wrong.Data = []byte{4, 3, 2, 1}
	if e = VerifyRouterQuote(s, trust, data, wrong, 0); e == nil {
		t.Fatal("calldata substitution accepted")
	}
	if e = VerifyQuote(s, trust, data, 0); e == nil {
		t.Fatal("Router quote accepted on Core-only API")
	}
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID     uint64
			Method string
			Params []map[string]json.RawMessage
		}
		if e := json.NewDecoder(r.Body).Decode(&req); e != nil {
			t.Fatal(e)
		}
		if req.Method != "broadcaster_submitTransaction" || len(req.Params) != 1 {
			t.Error("wrong RPC shape")
		}
		p := req.Params[0]
		var quote Quote
		if e := json.Unmarshal(p["quote"], &quote); e != nil {
			t.Error("quote must be inner Quote object", e)
		}
		var sig string
		json.Unmarshal(p["signature"], &sig)
		if sig != s.Signature {
			t.Error("signature missing")
		}
		var actual RouterEnvelope
		if e := json.Unmarshal(p["router"], &actual); e != nil || actual.To != env.To {
			t.Error("Router envelope missing", e)
		}
		json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": Status{ID: "router-job", State: "pending"}})
	}))
	defer server.Close()
	client, e := NewClient(Configuration{Endpoint: server.URL, QuoteSigner: trust.Signer, FeeRecipient: s.Quote.FeeRecipient}, Options{HTTPClient: server.Client()})
	if e != nil {
		t.Fatal(e)
	}
	status, e := client.SubmitRouterTransaction(context.Background(), data, env, s)
	if e != nil || status.ID != "router-job" {
		t.Fatal(status, e)
	}
}

func TestRouterBindingUsesInnerAndAccountAuthorizationDeadlines(t *testing.T) {
	raw, e := os.ReadFile("testdata/payment-router-envelope-vectors.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct{ Envelope RouterEnvelope }
	if e = json.Unmarshal(raw, &rows); e != nil {
		t.Fatal(e)
	}
	for _, native := range []bool{false, true} {
		env := rows[0].Envelope
		name := "payTokenFor"
		inner := []byte(env.Data)
		var account []any
		if native {
			env = rows[1].Envelope
			account, e = unpackRelay("executeNativePayment", env.Data)
			if e != nil {
				t.Fatal(e)
			}
			inner = account[0].([]byte)
			name = "payNative"
		}
		args, e := unpackRelay(name, inner)
		if e != nil {
			t.Fatal(e)
		}
		m := relayMethods["transact"]
		core, e := m.Inputs.Pack(args[0], args[1])
		if e != nil {
			t.Fatal(e)
		}
		core = append(append([]byte(nil), m.ID...), core...)
		intent := abi.ConvertType(args[2], new(relayIntent)).(*relayIntent)
		intent.Deadline = new(big.Int).Sub(intent.Deadline, big.NewInt(100))
		args[2] = *intent
		encoded, e := relayMethods[name].Inputs.Pack(args...)
		if e != nil {
			t.Fatal(e)
		}
		env.Data = append(append([]byte(nil), relayMethods[name].ID...), encoded...)
		expected := new(big.Int).Set(intent.Deadline)
		if native {
			account[0] = []byte(env.Data)
			account[3] = new(big.Int).Sub(expected, big.NewInt(100))
			expected.Set(account[3].(*big.Int))
			encoded, e = relayMethods["executeNativePayment"].Inputs.Pack(account...)
			if e != nil {
				t.Fatal(e)
			}
			env.Data = append(append([]byte(nil), relayMethods["executeNativePayment"].ID...), encoded...)
		}
		limit, e := relayDeadline(core, env)
		if e != nil || limit.Cmp(expected) != 0 {
			t.Fatal(limit, e)
		}
	}
}
