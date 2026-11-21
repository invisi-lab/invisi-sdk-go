package broadcaster

import (
	"bytes"
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"math/big"
	"os"
	"testing"
)

func TestPreproveSemanticVectors(t *testing.T) {
	raw, e := os.ReadFile("testdata/payment-router-envelope-vectors.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Core     common.Address
		Envelope RouterEnvelope
	}
	if e = json.Unmarshal(raw, &rows); e != nil {
		t.Fatal(e)
	}
	var vectors []any
	for _, v := range rows {
		draft, e := NormalizeRouterEnvelope(v.Envelope)
		if e != nil {
			t.Fatal(e)
		}
		hash, e := RouterIntentHash(big.NewInt(31337), v.Core, v.Envelope)
		if e != nil {
			t.Fatal(e)
		}
		dh, e := RouterIntentHash(big.NewInt(31337), v.Core, draft)
		if e != nil || dh != hash {
			t.Fatal("proof/signature changed semantic hash", e)
		}
		rawFinal, _ := EnvelopeHash(big.NewInt(31337), v.Core, v.Envelope)
		rawDraft, _ := EnvelopeHash(big.NewInt(31337), v.Core, draft)
		if rawFinal == rawDraft {
			t.Fatal("raw identity did not change")
		}
		coreData := func(env RouterEnvelope) []byte {
			inner := []byte(env.Data)
			name := "payTokenFor"
			if bytes.Equal(inner[:4], relayMethods["executeNativePayment"].ID) {
				a, _ := unpackRelay("executeNativePayment", inner)
				inner = a[0].([]byte)
				name = "payNative"
			}
			a, _ := unpackRelay(name, inner)
			b, _ := relayMethods["transact"].Inputs.Pack(a[0], a[1])
			return append(bytes.Clone(relayMethods["transact"].ID), b...)
		}
		if e := ValidateRouterSubmission(coreData(draft), draft); e == nil {
			t.Fatal("unsigned draft accepted at submission")
		}
		if _, e := relayDeadline(coreData(v.Envelope), draft); e == nil {
			t.Fatal("wrong-phase Core calldata accepted")
		}
		changed := draft
		changed.To = common.HexToAddress("0x1234")
		h, _ := RouterIntentHash(big.NewInt(31337), v.Core, changed)
		if h == hash {
			t.Fatal("target unbound")
		}
		vectors = append(vectors, map[string]any{"chainId": "31337", "core": v.Core, "finalEnvelope": v.Envelope, "draftEnvelope": draft, "finalCoreData": hexutil.Encode(coreData(v.Envelope)), "draftCoreData": hexutil.Encode(coreData(draft)), "intentHash": hash})
	}
	if os.Getenv("WRITE_ROUTER_VECTORS") == "" {
		stored, e := os.ReadFile("testdata/router-preprove-intent-vectors.json")
		if e != nil {
			t.Fatal(e)
		}
		var expected []struct {
			ChainID                      string `json:"chainId"`
			Core                         common.Address
			FinalEnvelope, DraftEnvelope RouterEnvelope
			IntentHash                   common.Hash
		}
		if e = json.Unmarshal(stored, &expected); e != nil {
			t.Fatal(e)
		}
		for _, v := range expected {
			for _, env := range []RouterEnvelope{v.FinalEnvelope, v.DraftEnvelope} {
				h, e := RouterIntentHash(big.NewInt(31337), v.Core, env)
				if e != nil || h != v.IntentHash {
					t.Fatal("frozen semantic hash parity", h, e)
				}
			}
		}
	}
	if os.Getenv("WRITE_ROUTER_VECTORS") != "" {
		b, _ := json.MarshalIndent(vectors, "", "  ")
		if e := os.WriteFile("testdata/router-preprove-intent-vectors.json", append(b, '\n'), 0644); e != nil {
			t.Fatal(e)
		}
	}
}
func TestRouterGasSharedVectors(t *testing.T) {
	inputs := []RouterGasInput{{CoreGasUnits: 100000, Mode: 0}, {CoreGasUnits: 100000, Mode: 1, Permit: true}, {CoreGasUnits: 100000, Mode: 2, Channel: 1, HopCount: 1, RouteBytes: 43}, {CoreGasUnits: 100000, Mode: 3, Channel: 1, HopCount: 8, RouteBytes: 204, AuthorizationCount: 1}, {CoreGasUnits: 100000, Mode: 2, Channel: 2, RouteBytes: 128}, {CoreGasUnits: 100000, Mode: 3, Channel: 2, RouteBytes: 128, ZeroExGasUnits: 250000, AuthorizationCount: 1}, {CoreGasUnits: routerMaxSafe, Mode: 0}, {CoreGasUnits: 100000, Mode: 2, Channel: 1, HopCount: 9}, {CoreGasUnits: 100000, Mode: 1, AuthorizationCount: 1}, {CoreGasUnits: 100000, Mode: 0, RouteBytes: 1}, {CoreGasUnits: 100000, Mode: 2, Channel: 1, HopCount: 1, RouteBytes: 131073}}
	var rows []any
	for _, i := range inputs {
		r := EstimateRouterGas(i)
		rows = append(rows, map[string]any{"input": i, "output": r})
	}
	if got := EstimateRouterGas(inputs[2]); !got.Available || got.GasUnits != 335688 {
		t.Fatal(got)
	}
	if got := EstimateRouterGas(inputs[4]); got.Available || got.Reason != "zeroex-route-gas-unavailable" {
		t.Fatal(got)
	}
	if got := EstimateRouterGas(inputs[6]); got.Available || got.Reason != "router-gas-overflow" {
		t.Fatal(got)
	}
	if os.Getenv("WRITE_ROUTER_VECTORS") != "" {
		b, _ := json.MarshalIndent(rows, "", "  ")
		if e := os.WriteFile("testdata/router-gas-model-vectors.json", append(b, '\n'), 0644); e != nil {
			t.Fatal(e)
		}
	}
	raw, e := os.ReadFile("testdata/router-gas-model-vectors.json")
	if e != nil {
		t.Fatal(e)
	}
	var stored []struct {
		Input  RouterGasInput
		Output RouterGasResult
	}
	if e = json.Unmarshal(raw, &stored); e != nil {
		t.Fatal(e)
	}
	for _, row := range stored {
		got, _ := json.Marshal(EstimateRouterGas(row.Input))
		want, _ := json.Marshal(row.Output)
		if !bytes.Equal(got, want) {
			t.Fatalf("model parity: %s != %s", got, want)
		}
	}
}
