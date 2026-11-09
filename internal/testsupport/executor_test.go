package testsupport

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/invisi-lab/invisi-sdk-go/wallet"
)

func TestExecutorPublicAndOwnerBinding(t *testing.T) {
	e, err := NewExecutor("0x" + strings.Repeat("0", 63) + "1")
	if err != nil {
		t.Fatal(err)
	}
	req := wallet.Request{Version: 1, ID: "test", OwnerAddress: strings.ToLower(e.OwnerAddress().Hex()), Method: wallet.Public, Params: json.RawMessage(`{}`)}
	response, err := e.ExecuteOne(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if err = wallet.ValidateResponse(req, response); err != nil {
		t.Fatal(err)
	}
	var public wallet.PublicResult
	if err = wallet.Decode(response.Result, &public); err != nil {
		t.Fatal(err)
	}
	if _, err = wallet.Bytes(public.SpendingPublicKey, 65); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(response.Result), "detectionKey") {
		t.Fatal("private detection key exposed")
	}
	req.OwnerAddress = "0x1111111111111111111111111111111111111111"
	if _, err = e.ExecuteOne(context.Background(), req); err == nil {
		t.Fatal("accepted different owner")
	}
}

func TestMalformedClueIsAdvisoryFalse(t *testing.T) {
	e, err := NewExecutor("0x" + strings.Repeat("0", 63) + "1")
	if err != nil {
		t.Fatal(err)
	}
	q := wallet.Request{Version: 1, ID: "bad-clue", OwnerAddress: strings.ToLower(e.OwnerAddress().Hex()), Method: wallet.Detect, Params: wallet.Encode(wallet.DetectParams{Clues: []string{"0x" + strings.Repeat("00", 68)}})}
	response, err := e.ExecuteOne(context.Background(), q)
	if err != nil {
		t.Fatal(err)
	}
	var v wallet.DetectResult
	if err = json.Unmarshal(response.Result, &v); err != nil {
		t.Fatal(err)
	}
	if len(v.Matches) != 1 || v.Matches[0] {
		t.Fatal("invalid clue should not match")
	}
}

func TestBatchValidatesBeforeCryptoAndReturnsBoundFailures(t *testing.T) {
	e, err := NewExecutor("0x" + strings.Repeat("0", 63) + "1")
	if err != nil {
		t.Fatal(err)
	}
	q, err := wallet.Prepare("public", e.OwnerAddress(), wallet.Public, struct{}{})
	if err != nil {
		t.Fatal(err)
	}
	invalid := q
	invalid.ID = "second"
	invalid.Params = wallet.Encode(map[string]string{"privateKey": "forbidden"})
	if responses, err := e.Execute(context.Background(), []wallet.Request{q, invalid}); err == nil || responses != nil {
		t.Fatal("invalid batch partially executed")
	}
	responses, err := e.Execute(context.Background(), []wallet.Request{q})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = wallet.ValidateResponses([]wallet.Request{q}, responses); err != nil {
		t.Fatal(err)
	}
	empty, err := e.Execute(context.Background(), nil)
	if err != nil || len(empty) != 0 {
		t.Fatal("empty batch failed")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err = e.Execute(ctx, []wallet.Request{q}); err == nil {
		t.Fatal("cancelled batch executed")
	}
}
