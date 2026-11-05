package session

import (
	"context"
	"errors"
	"testing"

	"github.com/invisi-lab/invisi-sdk-go/core"
	transaction "github.com/invisi-lab/invisi-sdk-go/internal/transaction"
	sdk "github.com/invisi-lab/invisi-sdk-go/sdk"
)

type codedPreparationError interface {
	error
	ErrorCode() string
}

func TestPreparationInvalidRecipientHasStableCode(t *testing.T) {
	invisi, input, _ := explicitFixture(t)
	input.Outputs[0] = sdk.NoteOutput{}
	_, err := invisi.PrepareTransfer(context.Background(), input)
	var coded codedPreparationError
	if !errors.As(err, &coded) || coded.ErrorCode() != "PREPARATION_FAILED" {
		t.Fatalf("missing stable invalid recipient code: %v", err)
	}
}
func TestPreparationAmountFailureCarriesOutputPosition(t *testing.T) {
	invisi, input, _ := explicitFixture(t)
	input.Outputs[0] = sdk.NoteOutput{}
	_, err := invisi.PrepareTransfer(context.Background(), input)
	var coded codedPreparationError
	if !errors.As(err, &coded) || coded.ErrorCode() != "PREPARATION_FAILED" {
		t.Fatalf("missing stable amount failure code: %v", err)
	}
}

func TestRawPayloadExceedsMemoLimitWithoutInterpretation(t *testing.T) {
	invisi, input, _ := explicitFixture(t)
	original := input.Outputs[0]
	payload := make([]byte, 300)
	payload[0] = 255
	input.Outputs[0], _ = sdk.NewNoteOutput(sdk.PaymentRequest{ChainID: core.ElementToBigInt(transaction.OutputSnapshot(original).Context.ChainID), ProtocolAddress: input.Protocol.Context().ProtocolAddress, TokenAddress: input.TokenAddress, Amount: original.Amount(), Recipient: sdk.InvisiRecipient{Address: original.To()}, Payload: payload})
	prepared, err := invisi.PrepareTransfer(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	if len(transaction.PreparedExecution(prepared).Plan.Outputs[0].Note.Payload) != 300 {
		t.Fatal("raw payload was truncated")
	}
}

func TestNilPublicWalletReturnsCapabilityError(t *testing.T) {
	invisi, input, _ := explicitFixture(t)
	input.Wallet = nil
	_, err := invisi.PrepareTransfer(context.Background(), input)
	var typed *sdk.InvisiError
	if !errors.As(err, &typed) || typed.Code != sdk.MissingCapability {
		t.Fatalf("nil identity should return structured capability failure: %v", err)
	}
}
