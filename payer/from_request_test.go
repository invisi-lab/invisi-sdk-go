package payer

import (
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/ipp"
)

func makeTestRequest() ipp.IPPRequest {
	return ipp.IPPRequest{
		V:       1,
		Network: "eip155:84532",
		Payload: ipp.IPPPayloadV1{
			IdentityCommitment: "0x00000000000000000000000000000000000000000000000000000000000000ab",
			TokenAddress:       "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",
			Operator:           "0x0000000000000000000000000000000000000001",
			PaymentTypes:       []ipp.PaymentType{ipp.PaymentTypeOpen},
		},
	}
}

func TestBuildPaymentFromRequest_Valid(t *testing.T) {
	req := makeTestRequest()

	// Set up known nonce and compute expected commitment
	var nonce fr.Element
	nonce.SetUint64(42)
	nonceHash := common.HexToHash(core.ElementToHex(nonce))

	chainID := big.NewInt(84532)
	amount := big.NewInt(100_000_000)
	icHash := common.HexToHash(req.Payload.IdentityCommitment)
	tokenAddr := common.HexToAddress(req.Payload.TokenAddress)

	// Compute valid paymentCommitment
	chainElem := core.BigIntToElement(chainID)
	icElem, _ := core.HexToElement(icHash.Hex())
	tokenHash := common.BytesToHash(tokenAddr.Bytes())
	tokenElem, _ := core.HexToElement(tokenHash.Hex())
	amountElem := core.BigIntToElement(amount)

	commitment, _ := core.ComputePaymentCommitment(chainElem, icElem, tokenElem, amountElem, nonce)
	commitmentHash := common.HexToHash(core.ElementToHex(commitment))

	result, err := BuildPaymentFromRequest(BuildFromRequestParams{
		Request:           req,
		Nonce:             nonceHash,
		PaymentCommitment: commitmentHash,
		Amount:            amount,
		TreeID:            big.NewInt(1),
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Calldata) < 4 {
		t.Fatal("calldata too short")
	}
	if result.ApproveAmount.Cmp(amount) != 0 {
		t.Errorf("approveAmount = %s, want %s", result.ApproveAmount, amount)
	}
}

func TestBuildPaymentFromRequest_InvalidCommitment(t *testing.T) {
	req := makeTestRequest()

	var nonce fr.Element
	nonce.SetUint64(42)
	nonceHash := common.HexToHash(core.ElementToHex(nonce))

	_, err := BuildPaymentFromRequest(BuildFromRequestParams{
		Request:           req,
		Nonce:             nonceHash,
		PaymentCommitment: common.HexToHash("0xdeadbeef"), // wrong
		Amount:            big.NewInt(100_000_000),
	})
	if err == nil {
		t.Error("expected error for invalid commitment")
	}
}

func TestBuildPaymentFromRequest_InvalidNetwork(t *testing.T) {
	req := makeTestRequest()
	req.Network = "invalid:chain"

	_, err := BuildPaymentFromRequest(BuildFromRequestParams{
		Request:           req,
		Nonce:             common.Hash{},
		PaymentCommitment: common.Hash{},
		Amount:            big.NewInt(100),
	})
	if err == nil {
		t.Error("expected error for invalid CAIP-2 network")
	}
}
