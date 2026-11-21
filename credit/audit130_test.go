package credit

import (
	"context"
	encodinghex "encoding/hex"
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/invisi-labs/invisi-sdk-go/primitives"
	"math/big"
	"testing"
)

func TestAudit130DefaultCreateSelector(t *testing.T) {
	b, s, _ := requestBuilder(t)
	p, e := b.Create(context.Background(), CreatePreparation{TokenAddress: common.BigToAddress(primitives.ElementToBigInt(s.TokenAddress)), Spender: "spender", Expiry: 2100000000, AvailableCredit: big.NewInt(90)})
	if e != nil {
		t.Fatal(e)
	}
	data, e := p.ExportBroadcasterIntent()
	if e != nil {
		t.Fatal(e)
	}
	if encodinghex.EncodeToString(data[:4]) != "4aa2aed8" {
		t.Fatalf("default selector %x", data[:4])
	}
}
func TestAudit130PinnedNetRepricesGrantAndRecovery(t *testing.T) {
	b, s, _ := requestBuilder(t)
	p, e := b.Create(context.Background(), CreatePreparation{TokenAddress: common.BigToAddress(primitives.ElementToBigInt(s.TokenAddress)), Spender: "spender", Expiry: 2100000000, AvailableCredit: big.NewInt(90), InitialPayment: &InitialPayment{Amount: big.NewInt(5), NetAmount: big.NewInt(5)}})
	if e != nil {
		t.Fatal(e)
	}
	i, _ := p.ExportIntent()
	quoteKey, _ := ethcrypto.ToECDSA(common.LeftPadBytes([]byte{9}, 32))
	client := creditBroadcaster(t, i.Context, quoteKey, common.HexToAddress("44"))
	if e = client.Quote(context.Background(), p, BroadcasterQuoteOptions{MaxFeeAmount: big.NewInt(2)}); e != nil {
		t.Fatal(e)
	}
	quoted, _ := p.ExportIntent()
	if quoted.Create.TotalAmount.Int64() != 6 || primitives.ElementToBigInt(quoted.OutputNote.Amount).Int64() != 5 {
		t.Fatal("quote did not preserve frozen net")
	}
	next, e := reprice(i, common.HexToAddress("44"), big.NewInt(1))
	if e != nil {
		t.Fatal(e)
	}
	if next.Create.TotalAmount.Int64() != 6 || next.OutputNote.Amount != primitives.ElementFromUint64(5) || next.CreatedState.AvailableCredit != primitives.ElementFromUint64(84) || next.CreatedState.TotalSpent != primitives.ElementFromUint64(6) {
		t.Fatal("net, debit or grant usage changed incorrectly")
	}
	restored, e := prepare(next)
	if e != nil {
		t.Fatal(e)
	}
	j, _ := restored.ExportIntent()
	again, e := reprice(j, common.HexToAddress("44"), big.NewInt(2))
	if e != nil {
		t.Fatal(e)
	}
	if again.Create.TotalAmount.Int64() != 7 || again.OutputNote.Amount != primitives.ElementFromUint64(5) || again.CreatedState.AvailableCredit != primitives.ElementFromUint64(83) {
		t.Fatal("recovery lost original net or principal")
	}
	if _, e = reprice(i, common.HexToAddress("44"), big.NewInt(90)); e == nil {
		t.Fatal("accepted exhausted principal")
	}
}

func TestAudit130LowCreateDefaultsCurrentAndExplicitLegacy(t *testing.T) {
	b, s, _ := requestBuilder(t)
	p, e := b.Create(context.Background(), CreatePreparation{TokenAddress: common.BigToAddress(primitives.ElementToBigInt(s.TokenAddress)), Spender: "spender", Expiry: 2100000000, AvailableCredit: big.NewInt(90)})
	if e != nil {
		t.Fatal(e)
	}
	i, _ := p.ExportIntent()
	params := *i.Create
	params.HasInitialPayment = nil
	data, e := params.Calldata(i.Context, []byte{1}, i.Deliveries)
	if e != nil {
		t.Fatal(e)
	}
	if common.Bytes2Hex(data[:4]) != "4aa2aed8" {
		t.Fatalf("low default selector %x", data[:4])
	}
	params.CreationABI = "legacy"
	if _, e = params.Calldata(i.Context, []byte{1}, i.Deliveries); e == nil {
		t.Fatal("historical ABI accepted current Credit identity")
	}

}
func TestAudit130NetInvariantRejectsTamperedRecovery(t *testing.T) {
	b, s, _ := requestBuilder(t)
	req := CreatePreparation{TokenAddress: common.BigToAddress(primitives.ElementToBigInt(s.TokenAddress)), Spender: "spender", Expiry: 2100000000, AvailableCredit: big.NewInt(90), InitialPayment: &InitialPayment{Amount: big.NewInt(5), NetAmount: big.NewInt(4)}}
	if _, e := b.Create(context.Background(), req); e == nil {
		t.Fatal("accepted inconsistent gross and net")
	}
	req.InitialPayment.NetAmount = big.NewInt(5)
	p, e := b.Create(context.Background(), req)
	if e != nil {
		t.Fatal(e)
	}
	req.InitialPayment.NetAmount.SetInt64(6)
	i, _ := p.ExportIntent()
	raw, e := json.Marshal(i)
	if e != nil {
		t.Fatal(e)
	}
	var restored Intent
	if e = json.Unmarshal(raw, &restored); e != nil {
		t.Fatal(e)
	}
	if _, e = prepare(restored); e != nil {
		t.Fatal(e)
	}
	restored.InitialPaymentCreationDelivery.State.AvailableCredit = primitives.ElementFromUint64(100)
	if _, e = prepare(restored); e == nil {
		t.Fatal("accepted forged original principal")
	}
}
