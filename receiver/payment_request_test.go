package receiver

import (
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-labs/invisi-sdk-go/core"
	"github.com/invisi-labs/invisi-sdk-go/ipp"
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
		TreeID:           big.NewInt(1),
		Operator:         common.HexToAddress("0x1234567890123456789012345678901234567890"),
		PaymentTypes:     []ipp.PaymentType{ipp.PaymentTypeOpen, ipp.PaymentTypeBlind},
	}
}

func TestCreatePaymentRequest_CorrectIdentityCommitment(t *testing.T) {
	params := makeBasePaymentRequestParams()
	params.Amount = big.NewInt(100_000_000)
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
	nk, err := core.DeriveNullifierKey(params.MasterKey, params.InvoiceID)
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

func TestCreatePaymentRequest_Mode1a_NonceAndPaymentCommitment(t *testing.T) {
	params := makeBasePaymentRequestParams()
	params.Amount = big.NewInt(100_000_000)
	// MultiPayer = false (default)

	request, err := CreatePaymentRequest(params)
	if err != nil {
		t.Fatal(err)
	}

	if request.Payload.Nonce == nil {
		t.Fatal("Nonce should be set in mode 1a")
	}
	if request.Payload.PaymentCommitment == nil {
		t.Fatal("PaymentCommitment should be set in mode 1a")
	}
	if request.Payload.Amount == nil || *request.Payload.Amount != "100000000" {
		t.Errorf("Amount = %v, want '100000000'", request.Payload.Amount)
	}

	// Verify paymentCommitment matches core computation
	nonceBigInt := new(big.Int)
	nonceBigInt.SetString((*request.Payload.Nonce)[2:], 16) // strip 0x prefix
	nonceElem := core.BigIntToElement(nonceBigInt)

	icElem, err := core.HexToElement(request.Payload.IdentityCommitment)
	if err != nil {
		t.Fatal(err)
	}
	tokenBigInt := new(big.Int).SetBytes(params.TokenAddress.Bytes())
	tokenElem := core.BigIntToElement(tokenBigInt)
	amountElem := core.BigIntToElement(params.Amount)
	chainElem := core.BigIntToElement(params.ChainID)

	expectedPC, err := core.ComputePaymentCommitment(chainElem, icElem, tokenElem, amountElem, nonceElem)
	if err != nil {
		t.Fatal(err)
	}
	expectedPCHex := core.ElementToHex(expectedPC)
	if *request.Payload.PaymentCommitment != expectedPCHex {
		t.Errorf("PaymentCommitment mismatch: got %s, want %s",
			*request.Payload.PaymentCommitment, expectedPCHex)
	}
}

func TestCreatePaymentRequest_Mode1b_MultiPayer(t *testing.T) {
	params := makeBasePaymentRequestParams()
	params.Amount = big.NewInt(100_000_000)
	params.MultiPayer = true
	endpoint := "https://pay.example.com"
	params.Endpoint = &endpoint

	request, err := CreatePaymentRequest(params)
	if err != nil {
		t.Fatal(err)
	}

	if request.Payload.Nonce != nil {
		t.Error("Nonce should not be set in multi-payer mode")
	}
	if request.Payload.PaymentCommitment != nil {
		t.Error("PaymentCommitment should not be set in multi-payer mode")
	}
	if request.Payload.Amount == nil || *request.Payload.Amount != "100000000" {
		t.Errorf("Amount = %v, want '100000000'", request.Payload.Amount)
	}
	if request.Payload.Endpoint == nil || *request.Payload.Endpoint != "https://pay.example.com" {
		t.Error("Endpoint should be set")
	}
}

func TestCreatePaymentRequest_Mode2_NoAmount(t *testing.T) {
	params := makeBasePaymentRequestParams()
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
	if request.Payload.Nonce != nil {
		t.Error("Nonce should not be set in mode 2")
	}
	if request.Payload.Endpoint == nil || *request.Payload.Endpoint != "https://pay.example.com" {
		t.Error("Endpoint should be set in mode 2")
	}
}

func TestCreatePaymentRequest_MetaFields(t *testing.T) {
	params := makeBasePaymentRequestParams()
	params.Amount = big.NewInt(100_000_000)
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
