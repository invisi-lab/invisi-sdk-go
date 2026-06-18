package receiver

import (
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/ipp"
)

func makeVoucher(t *testing.T, recv common.Address) ipp.IPPVoucher {
	t.Helper()
	var mk, inv fr.Element // payer masterKey, random invoiceId
	mk.SetUint64(999)
	inv.SetUint64(7)
	nk, _ := core.DeriveNullifierKey(mk, inv)
	ic, _ := core.ComputeIdentityCommitment(nk, core.BigIntToElement(recv.Big()))
	var amount fr.Element
	amount.SetUint64(100_000_000)
	var nonce fr.Element
	nonce.SetUint64(12345)
	token := common.HexToAddress("0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913")
	chainID := core.ElementFromUint64(84532)
	pc, _ := core.ComputePaymentCommitment(chainID, ic, core.BigIntToElement(token.Big()), amount, nonce)
	return ipp.IPPVoucher{
		V: 1, Network: "eip155:84532", Type: ipp.MessageTypeVoucher,
		Payload: ipp.IPPVoucherV1{
			IdentityCommitment: core.ElementToHex(ic),
			PaymentCommitment:  core.ElementToHex(pc),
			TokenAddress:       token.Hex(),
			Operator:           "0x1234567890123456789012345678901234567890",
			Amount:             "100000000",
			Nonce:              core.ElementToHex(nonce),
			NullifierKey:       core.ElementToHex(nk),
			TreeID:             "1",
		},
	}
}

func TestVerifyVoucher_OK(t *testing.T) {
	recv := common.HexToAddress("0x00000000000000000000000000000000DeaD0000")
	ok, err := VerifyVoucher(makeVoucher(t, recv), recv, big.NewInt(84532))
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Error("expected valid voucher")
	}
}

func TestVerifyVoucher_WrongAddress(t *testing.T) {
	recv := common.HexToAddress("0x00000000000000000000000000000000DeaD0000")
	other := common.HexToAddress("0x00000000000000000000000000000000bEEf0000")
	ok, _ := VerifyVoucher(makeVoucher(t, recv), other, big.NewInt(84532))
	if ok {
		t.Error("expected rejection for wrong receiving address")
	}
}

func TestVerifyVoucher_TamperedAmount(t *testing.T) {
	recv := common.HexToAddress("0x00000000000000000000000000000000DeaD0000")
	v := makeVoucher(t, recv)
	v.Payload.Amount = "999"
	ok, _ := VerifyVoucher(v, recv, big.NewInt(84532))
	if ok {
		t.Error("expected rejection for tampered amount")
	}
}

func TestVoucherToWithdrawRequest_NoSecret(t *testing.T) {
	recv := common.HexToAddress("0x00000000000000000000000000000000DeaD0000")
	v := makeVoucher(t, recv)
	req := VoucherToWithdrawRequest(v)
	if len(req.PaymentCommitments) != 1 || req.PaymentCommitments[0] != v.Payload.PaymentCommitment {
		t.Error("paymentCommitments mismatch")
	}
	if req.IdentityCommitment != v.Payload.IdentityCommitment {
		t.Error("ic mismatch")
	}
	// nullifierKey must not appear in the withdraw request
}
