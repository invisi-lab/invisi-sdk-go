package receiver

import (
	"math/big"
	"strings"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/ipp"
)

func makeBasePaymentRequestParams() CreatePaymentRequestParams {
	var mk fr.Element
	mk.SetUint64(42)
	var invID fr.Element
	invID.SetUint64(1)

	return CreatePaymentRequestParams{
		MasterKey:        mk,
		InvoiceID:        invID,
		ReceivingAddress: common.HexToAddress("0x000000000000000000000000000000000000dead"),
		TokenAddress:     common.HexToAddress("0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913"),
		ChainID:          big.NewInt(84532),
		Operator:         common.HexToAddress("0x1234567890123456789012345678901234567890"),
		PaymentTypes:     []ipp.PaymentType{ipp.PaymentTypeOpen, ipp.PaymentTypeBlind},
	}
}

func TestCreatePaymentRequest_CorrectIdentityCommitment(t *testing.T) {
	params := makeBasePaymentRequestParams()
	params.Amount = big.NewInt(100_000_000)
	params.Mode = "single"
	endpoint := "https://pay.example.com"
	params.Endpoint = &endpoint

	request, err := CreatePaymentRequest(params)
	if err != nil {
		t.Fatal(err)
	}

	// Validate the request itself
	if err := ipp.ValidateIPPRequest(*request); err != nil {
		t.Fatalf("invalid IPP request: %v", err)
	}

	// Verify identityCommitment matches core computation
	nk, err := core.DeriveNoteSecret(params.MasterKey, params.InvoiceID)
	if err != nil {
		t.Fatal(err)
	}
	recvAddrBigInt := new(big.Int).SetBytes(params.ReceivingAddress.Bytes())
	recvAddrElem := core.BigIntToElement(recvAddrBigInt)
	expectedIC, err := core.ComputeIdentityCommitment(nk, recvAddrElem)
	if err != nil {
		t.Fatal(err)
	}
	expectedICHex := core.ElementToHex(expectedIC)
	if request.Payload.IdentityCommitment != expectedICHex {
		t.Errorf("IdentityCommitment mismatch: got %s, want %s",
			request.Payload.IdentityCommitment, expectedICHex)
	}
}

func TestCreatePaymentRequest_CAIP2Network(t *testing.T) {
	params := makeBasePaymentRequestParams()
	params.Amount = big.NewInt(100_000_000)
	params.Mode = "single"
	endpoint := "https://pay.example.com"
	params.Endpoint = &endpoint

	request, err := CreatePaymentRequest(params)
	if err != nil {
		t.Fatal(err)
	}

	if request.Network != "eip155:84532" {
		t.Errorf("Network = %s, want eip155:84532", request.Network)
	}

	chainID, err := ipp.CAIP2ToChainID(request.Network)
	if err != nil {
		t.Fatal(err)
	}
	if chainID.Cmp(big.NewInt(84532)) != 0 {
		t.Errorf("parsed chainID = %s, want 84532", chainID)
	}
}

func TestCreatePaymentRequest_Single(t *testing.T) {
	var masterKey, invoiceID fr.Element
	masterKey.SetUint64(999)
	invoiceID.SetUint64(1)
	endpoint := "https://pay.merchant.com"

	req, err := CreatePaymentRequest(CreatePaymentRequestParams{
		MasterKey:        masterKey,
		InvoiceID:        invoiceID,
		ReceivingAddress: common.HexToAddress("0xdead"),
		TokenAddress:     common.HexToAddress("0x" + strings.Repeat("aa", 20)),
		ChainID:          big.NewInt(84532),
		Operator:         common.HexToAddress("0x1234567890123456789012345678901234567890"),
		PaymentTypes:     []ipp.PaymentType{ipp.PaymentTypeOpen, ipp.PaymentTypeBlind},
		Mode:             "single",
		Amount:           big.NewInt(100_000_000),
		Endpoint:         &endpoint,
	})
	if err != nil {
		t.Fatal(err)
	}
	if req.Payload.Mode != "single" {
		t.Errorf("Mode = %q, want single", req.Payload.Mode)
	}
	if req.Payload.Amount == nil || *req.Payload.Amount != "100000000" {
		t.Errorf("Amount = %v", req.Payload.Amount)
	}
}

func TestCreatePaymentRequest_Multi(t *testing.T) {
	var masterKey, invoiceID fr.Element
	masterKey.SetUint64(999)
	invoiceID.SetUint64(2)
	endpoint := "https://pay.merchant.com"

	req, err := CreatePaymentRequest(CreatePaymentRequestParams{
		MasterKey:        masterKey,
		InvoiceID:        invoiceID,
		ReceivingAddress: common.HexToAddress("0xdead"),
		TokenAddress:     common.HexToAddress("0x" + strings.Repeat("aa", 20)),
		ChainID:          big.NewInt(84532),
		Operator:         common.HexToAddress("0x1234567890123456789012345678901234567890"),
		PaymentTypes:     []ipp.PaymentType{ipp.PaymentTypeOpen},
		Mode:             "multi",
		Endpoint:         &endpoint,
	})
	if err != nil {
		t.Fatal(err)
	}
	if req.Payload.Mode != "multi" {
		t.Errorf("Mode = %q, want multi", req.Payload.Mode)
	}
	if req.Payload.Amount != nil {
		t.Errorf("Amount should be nil for open-amount multi, got %v", req.Payload.Amount)
	}
}

func TestCreatePaymentRequest_Mode2_NoAmount(t *testing.T) {
	params := makeBasePaymentRequestParams()
	params.Mode = "multi"
	// No Amount set
	endpoint := "https://pay.example.com"
	params.Endpoint = &endpoint

	request, err := CreatePaymentRequest(params)
	if err != nil {
		t.Fatal(err)
	}

	if request.Payload.Amount != nil {
		t.Error("Amount should not be set in mode 2")
	}
	if request.Payload.Endpoint == nil || *request.Payload.Endpoint != "https://pay.example.com" {
		t.Error("Endpoint should be set in mode 2")
	}
}

func TestCreatePaymentRequest_MetaFields(t *testing.T) {
	params := makeBasePaymentRequestParams()
	params.Amount = big.NewInt(100_000_000)
	params.Mode = "single"
	params.Meta = map[string]string{
		"merchantName": "Coffee Shop",
		"orderId":      "ORD-001",
	}

	request, err := CreatePaymentRequest(params)
	if err != nil {
		t.Fatal(err)
	}

	if request.Meta == nil {
		t.Fatal("Meta should be set")
	}
	if request.Meta["merchantName"] != "Coffee Shop" {
		t.Errorf("merchantName = %s, want 'Coffee Shop'", request.Meta["merchantName"])
	}
	if request.Meta["orderId"] != "ORD-001" {
		t.Errorf("orderId = %s, want 'ORD-001'", request.Meta["orderId"])
	}
}

func TestCreatePaymentRequest_PaymentTypes(t *testing.T) {
	params := makeBasePaymentRequestParams()
	params.Amount = big.NewInt(100_000_000)
	params.Mode = "single"
	params.PaymentTypes = []ipp.PaymentType{ipp.PaymentTypeOpen}

	request, err := CreatePaymentRequest(params)
	if err != nil {
		t.Fatal(err)
	}

	if len(request.Payload.PaymentTypes) != 1 {
		t.Fatalf("PaymentTypes length = %d, want 1", len(request.Payload.PaymentTypes))
	}
	if request.Payload.PaymentTypes[0] != ipp.PaymentTypeOpen {
		t.Errorf("PaymentTypes[0] = %s, want 'open'", request.Payload.PaymentTypes[0])
	}
}
