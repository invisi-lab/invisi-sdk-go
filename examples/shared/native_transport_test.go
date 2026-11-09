package shared

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/wallet"
)

func TestNativeTransportBindsExplicitBatchAndRejectsNull(t *testing.T) {
	q, err := wallet.Prepare("public", common.HexToAddress("0x1111111111111111111111111111111111111111"), wallet.Public, struct{}{})
	if err != nil {
		t.Fatal(err)
	}
	r := wallet.Response{Version: 1, ID: q.ID, OwnerAddress: q.OwnerAddress, Method: q.Method, Error: &wallet.ResponseError{Code: "WALLET_LOCKED"}}
	raw, _ := json.Marshal([]wallet.Response{r})
	var output bytes.Buffer
	host := NativeWalletClient{Encoder: json.NewEncoder(&output), Decoder: json.NewDecoder(bytes.NewReader(raw))}
	got, err := host.Execute(context.Background(), []wallet.Request{q})
	if err != nil || got[0].Error.Code != "WALLET_LOCKED" {
		t.Fatalf("explicit failure lost: %v", err)
	}
	var sent []wallet.Request
	if err = json.Unmarshal(output.Bytes(), &sent); err != nil || len(sent) != 1 || sent[0].ID != q.ID {
		t.Fatal("request not sent")
	}
	host.Decoder = json.NewDecoder(bytes.NewBufferString("null"))
	if _, err = host.Execute(context.Background(), nil); err == nil {
		t.Fatal("null batch accepted")
	}
	host.Decoder = json.NewDecoder(bytes.NewBufferString("[]"))
	output.Reset()
	if _, err = host.Execute(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	if output.String() != "[]\n" {
		t.Fatal("empty batch encoded as null")
	}
}
