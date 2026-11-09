package native

import (
	"context"
	"encoding/json"
	"github.com/invisi-lab/invisi-sdk-go/wallet"
	"strings"
	"testing"
)

func TestExecutorPublicAndOwnerBinding(t *testing.T) {
	e, err := NewExecutor("0x" + strings.Repeat("0", 63) + "1")
	if err != nil {
		t.Fatal(err)
	}
	req := wallet.Request{Version: 1, ID: "test", OwnerAddress: strings.ToLower(e.OwnerAddress().Hex()), Method: wallet.Public, Params: json.RawMessage(`{}`)}
	response, err := e.Execute(context.Background(), req)
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
	if _, err = e.Execute(context.Background(), req); err == nil {
		t.Fatal("accepted different owner")
	}
}

func TestMalformedClueIsAdvisoryFalse(t *testing.T) {
	e, err := NewExecutor("0x" + strings.Repeat("0", 63) + "1")
	if err != nil {
		t.Fatal(err)
	}
	q := wallet.Request{Version: 1, ID: "bad-clue", OwnerAddress: strings.ToLower(e.OwnerAddress().Hex()), Method: wallet.Detect, Params: wallet.Encode(wallet.DetectParams{Clues: []string{"0x" + strings.Repeat("00", 68)}})}
	response, err := e.Execute(context.Background(), q)
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
