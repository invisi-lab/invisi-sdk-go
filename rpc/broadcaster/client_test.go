package broadcaster

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/invisi-lab/invisi-sdk-go/protocol/contracts"
)

func intent(t *testing.T) ([]byte, Trust, *ecdsa.PrivateKey) {
	t.Helper()
	key, _ := crypto.GenerateKey()
	trust := Trust{big.NewInt(97), common.HexToAddress("0x11"), crypto.PubkeyToAddress(key.PublicKey)}
	a, e := contracts.InvisiProtocolMetaData.GetAbi()
	if e != nil {
		t.Fatal(e)
	}
	p := contracts.IInvisiProtocolShieldParams{Token: common.HexToAddress("0x22"), Deadline: big.NewInt(time.Now().Unix() + 3600), FeeRecipient: common.HexToAddress("0x33"), FeeAmount: big.NewInt(5)}
	f := contracts.IInvisiProtocolFundingParams{Amount: big.NewInt(105), SigNonce: big.NewInt(0), Deadline: p.Deadline, Permit: contracts.IMetaTxTokenPermit{Deadline: big.NewInt(0)}}
	data, e := a.Pack("shield", p, []contracts.IInvisiProtocolShieldOutput{}, []contracts.IInvisiProtocolDelivery{}, f)
	if e != nil {
		t.Fatal(e)
	}
	return data, trust, key
}
func signed(t *testing.T, data []byte, trust Trust, key *ecdsa.PrivateKey) SignedQuote {
	t.Helper()
	i, e := DecodeIntent(data)
	if e != nil {
		t.Fatal(e)
	}
	h, e := RequestHash(data)
	if e != nil {
		t.Fatal(e)
	}
	q := Quote{Token: i.Token, FeeRecipient: i.FeeRecipient, FeeAmount: "5", Deadline: i.Deadline.String(), MaxGas: "100000", RequestHash: h, Execution: "direct"}
	td, _ := QuoteTypedData(q, trust)
	hsh, _, e := apitypes.TypedDataAndHash(td)
	if e != nil {
		t.Fatal(e)
	}
	sig, e := crypto.Sign(hsh, key)
	if e != nil {
		t.Fatal(e)
	}
	return SignedQuote{Quote: q, Signature: "0x" + common.Bytes2Hex(sig)}
}
func TestQuoteAuthenticatesActualContractCalldata(t *testing.T) {
	data, trust, key := intent(t)
	s := signed(t, data, trust, key)
	if e := VerifyQuote(s, trust, data, uint64(time.Now().Unix())); e != nil {
		t.Fatal(e)
	}
	wrong := trust
	wrong.ChainID = big.NewInt(98)
	if VerifyQuote(s, wrong, data, 0) == nil {
		t.Fatal("wrong chain accepted")
	}
	wrong = trust
	wrong.Protocol = common.HexToAddress("0x99")
	if VerifyQuote(s, wrong, data, 0) == nil {
		t.Fatal("wrong protocol accepted")
	}
	s.Quote.FeeAmount = "6"
	if VerifyQuote(s, trust, data, 0) == nil {
		t.Fatal("wrong fee accepted")
	}
	if _, e := RequestHash(append(data, make([]byte, 32)...)); e == nil {
		t.Fatal("noncanonical data accepted")
	}
}
func TestClientPositionalJSONRPCAndSafeSubmission(t *testing.T) {
	data, trust, key := intent(t)
	quote := signed(t, data, trust, key)
	var methods []string
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			JSONRPC string            `json:"jsonrpc"`
			ID      uint64            `json:"id"`
			Method  string            `json:"method"`
			Params  []json.RawMessage `json:"params"`
		}
		if e := json.NewDecoder(r.Body).Decode(&req); e != nil || len(req.Params) != 1 || req.JSONRPC != "2.0" {
			t.Error("not positional JSONRPC2", e)
		}
		methods = append(methods, req.Method)
		var result any
		switch req.Method {
		case "broadcaster_getQuote":
			result = quote
		case "broadcaster_estimateQuote":
			result = Fee{quote.Quote.FeeRecipient, "5", "100000"}
		default:
			result = Status{ID: "key", State: "uncertain"}
		}
		json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": result})
	}))
	defer server.Close()
	c, e := NewClient(Configuration{server.URL, trust.Signer, quote.Quote.FeeRecipient}, Options{server.Client()})
	if e != nil {
		t.Fatal(e)
	}
	if _, e = c.EstimateQuote(context.Background(), data); e != nil {
		t.Fatal(e)
	}
	verified, e := c.GetQuote(context.Background(), data, trust)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = c.SubmitTransaction(context.Background(), data, quote); e == nil {
		t.Fatal("unverified context accepted")
	}
	if _, e = c.SubmitTransaction(context.Background(), data, verified); e != nil {
		t.Fatal(e)
	}
	if _, e = c.Status(context.Background(), "key"); e != nil {
		t.Fatal(e)
	}
	if len(methods) != 4 {
		t.Fatal(methods)
	}
}

func TestEverySupportedIntentCanonicalHash(t *testing.T) {
	a, _ := contracts.InvisiProtocolMetaData.GetAbi()
	var initialize func(reflect.Value)
	initialize = func(v reflect.Value) {
		switch v.Kind() {
		case reflect.Pointer:
			if v.Type() == reflect.TypeOf((*big.Int)(nil)) {
				v.Set(reflect.ValueOf(big.NewInt(0)))
			}
		case reflect.Struct:
			for j := 0; j < v.NumField(); j++ {
				initialize(v.Field(j))
			}
		case reflect.Slice:
			v.Set(reflect.MakeSlice(v.Type(), 0, 0))
		}
	}
	found := map[string]bool{}
	for _, m := range a.Methods {
		switch m.RawName {
		case "shield", "transact", "createCredit", "creditDraw", "updateCredit", "revokeCredit":
		default:
			continue
		}
		if m.RawName == "transact" && (len(m.Inputs) < 2 || m.Inputs[1].Type.T != abi.TupleTy) {
			continue
		}
		args := make([]any, len(m.Inputs))
		for j, input := range m.Inputs {
			v := reflect.New(input.Type.GetType()).Elem()
			initialize(v)
			args[j] = v.Interface()
		}
		encoded, e := m.Inputs.Pack(args...)
		if e != nil {
			t.Fatal(m.RawName, e)
		}
		data := append(append([]byte{}, m.ID...), encoded...)
		hash, err := RequestHash(data)
		if err != nil {
			t.Fatal(m.RawName, err)
		}
		if m.RawName != "shield" && m.RawName != "revokeCredit" {
			args[0] = []byte{1, 2, 3}
		}
		if m.RawName == "revokeCredit" {
			args[7] = []byte{4, 5}
		}
		if m.RawName == "transact" {
			v := reflect.ValueOf(args[1])
			copy := reflect.New(v.Type()).Elem()
			copy.Set(v)
			field(field(copy, "Funding"), "Signature").SetBytes([]byte{4, 5})
			args[1] = copy.Interface()
		} else if m.RawName == "shield" {
			v := reflect.ValueOf(args[3])
			copy := reflect.New(v.Type()).Elem()
			copy.Set(v)
			field(copy, "Signature").SetBytes([]byte{4, 5})
			args[3] = copy.Interface()
		}
		changed, err := m.Inputs.Pack(args...)
		if err != nil {
			t.Fatal(err)
		}
		changed = append(append([]byte{}, m.ID...), changed...)
		other, err := RequestHash(changed)
		if err != nil || other != hash {
			t.Fatal(m.RawName, "proof/funding signature changed request hash", err)
		}
		found[m.RawName] = true
	}
	if len(found) != 6 {
		t.Fatal(found)
	}
}

func TestRejectsMalformedAndOversizedRPCResponses(t *testing.T) {
	for _, body := range []string{`{"jsonrpc":"2.0","id":99,"result":{}}`, `{"jsonrpc":"2.0","id":1,"result":{},"error":{"code":1,"message":"x"}}`, `{"jsonrpc":"2.0","id":1,"error":{"code":"1","message":"x"}}`, strings.Repeat(" ", 65537)} {
		t.Run(fmt.Sprint(len(body), body[:min(16, len(body))]), func(t *testing.T) {
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(body)) }))
			defer server.Close()
			c, _ := NewClient(Configuration{server.URL, common.HexToAddress("0x11"), common.HexToAddress("0x22")}, Options{server.Client()})
			if _, e := c.Status(context.Background(), "key"); e == nil {
				t.Fatal("invalid response accepted")
			}
		})
	}
}
