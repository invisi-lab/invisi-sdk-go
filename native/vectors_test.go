package native

import (
	"context"
	"encoding/json"
	"github.com/invisi-lab/invisi-sdk-go/wallet"
	"os"
	"reflect"
	"testing"
)

func TestPortableWalletWireVectors(t *testing.T) {
	raw, err := os.ReadFile("testdata/wallet-wire-vectors.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		PrivateKey string `json:"privateKey"`
		Cases      []struct {
			Request  wallet.Request  `json:"request"`
			Response wallet.Response `json:"response"`
		} `json:"cases"`
	}
	if err = json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	executor, err := NewExecutor(fixture.PrivateKey)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range fixture.Cases {
		t.Run(string(c.Request.Method), func(t *testing.T) {
			got, err := executor.Execute(context.Background(), c.Request)
			if err != nil {
				t.Fatal(err)
			}
			var actual, expected any
			json.Unmarshal(wallet.Encode(got), &actual)
			json.Unmarshal(wallet.Encode(c.Response), &expected)
			if !reflect.DeepEqual(actual, expected) {
				t.Fatalf("wire response mismatch\nactual: %s\nexpected: %s", wallet.Encode(got), wallet.Encode(c.Response))
			}
		})
	}
}
