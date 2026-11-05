package sdk

import (
	transaction "github.com/invisi-lab/invisi-sdk-go/internal/transaction"
)

type EVMRecipient = transaction.EVMRecipient

type IdentityRecipient = transaction.IdentityRecipient

type InvisiRecipient = transaction.InvisiRecipient

func NewNoteOutput(input any, options ...NoteOutputExecutionOptions) (NoteOutput, error) {
	return transaction.NewNoteOutput(input, options...)
}

type NoteOutputExecutionOptions = transaction.NoteOutputExecutionOptions

func EncodePaymentRequest(request PaymentRequest) (string, error) {
	return transaction.EncodePaymentRequest(request)
}
func DecodePaymentRequest(encoded string) (PaymentRequest, error) {
	return transaction.DecodePaymentRequest(encoded)
}

type NoteInput = transaction.NoteInput

type NoteOutput = transaction.NoteOutput

type NoteOutputContext = transaction.NoteOutputContext

type PaymentRequest = transaction.PaymentRequest

type NoteRecipient = transaction.NoteRecipient
