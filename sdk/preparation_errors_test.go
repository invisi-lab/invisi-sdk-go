package sdk

import (
	"context"
	"errors"
	"math/big"
	"testing"
)

type codedPreparationError interface {
	error
	ErrorCode() string
}

func TestPreparationInvalidRecipientHasStableCode(t *testing.T) {
	invisi, input, _ := explicitFixture(t)
	input.Outputs[0].To = "invalid"
	_, err := invisi.PrepareTransfer(context.Background(), input)
	var coded codedPreparationError
	if !errors.As(err, &coded) || coded.ErrorCode() != "INVALID_RECIPIENT" {
		t.Fatalf("missing stable invalid recipient code: %v", err)
	}
}
func TestPreparationAmountFailureCarriesOutputPosition(t *testing.T) {
	invisi, input, _ := explicitFixture(t)
	input.Outputs[0].Amount = big.NewInt(-1)
	_, err := invisi.PrepareTransfer(context.Background(), input)
	var coded codedPreparationError
	if !errors.As(err, &coded) || coded.ErrorCode() != "PREPARATION_FAILED" {
		t.Fatalf("missing stable amount failure code: %v", err)
	}
}

func TestRawPayloadExceedsMemoLimitWithoutInterpretation(t *testing.T) {
	invisi, input, _ := explicitFixture(t)
	input.Outputs[0].Payload = make([]byte, 300)
	input.Outputs[0].Payload[0] = 255
	prepared, err := invisi.PrepareTransfer(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	if len(prepared.operation.plan.Outputs[0].Note.Payload) != 300 {
		t.Fatal("raw payload was truncated")
	}
}

func TestNilPublicWalletReturnsCapabilityError(t *testing.T) {
	invisi, input, _ := explicitFixture(t)
	input.Wallet = nil
	_, err := invisi.PrepareTransfer(context.Background(), input)
	var typed *InvisiError
	if !errors.As(err, &typed) || typed.Code != MissingCapability {
		t.Fatalf("nil identity should return structured capability failure: %v", err)
	}
}
