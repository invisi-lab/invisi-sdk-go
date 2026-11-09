package session

import (
	"context"
	"encoding/json"
	"github.com/invisi-lab/invisi-sdk-go/wallet"
	"math/big"
	"testing"
)

func TestDetachedWalletWireAuthorization(t *testing.T) {
	invisi, input, _ := explicitFixture(t)
	ctx := context.Background()
	p, err := invisi.PrepareTransfer(ctx, input)
	if err != nil {
		t.Fatal(err)
	}
	q, err := p.ExportWalletRequest(ctx, "external-1")
	if err != nil {
		t.Fatal(err)
	}
	response, err := testExecutor(input.Wallet).Execute(ctx, q)
	if err != nil {
		t.Fatal(err)
	}
	wrong := response
	wrong.ID = "other"
	if _, err = p.WithWalletResponse(ctx, q, wrong); err == nil {
		t.Fatal("accepted replayed authorization")
	}
	if _, err = p.WithWalletResponse(ctx, q, response); err != nil {
		t.Fatal(err)
	}
}

func TestWalletWireCannotAuthorizeDifferentStatement(t *testing.T) {
	invisi, input, _ := explicitFixture(t)
	ctx := context.Background()
	p, err := invisi.PrepareTransfer(ctx, input)
	if err != nil {
		t.Fatal(err)
	}
	q, err := p.ExportWalletRequest(ctx, "different-statement")
	if err != nil {
		t.Fatal(err)
	}
	var params wallet.AuthorizeParams
	if err = json.Unmarshal(q.Params, &params); err != nil {
		t.Fatal(err)
	}
	value, ok := new(big.Int).SetString(params.Words[2], 10)
	if !ok {
		t.Fatal("invalid fixture word")
	}
	params.Words[2] = value.Add(value, big.NewInt(1)).String()
	q.Params = wallet.Encode(params)
	response, err := testExecutor(input.Wallet).Execute(ctx, q)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = p.WithWalletResponse(ctx, q, response); err == nil {
		t.Fatal("accepted valid signature for different statement")
	}
}
