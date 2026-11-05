package sdk

import transaction "github.com/invisi-lab/invisi-sdk-go/internal/transaction"

type PaymentNoteData = transaction.PaymentNoteData
type PaymentNote = transaction.PaymentNote
type VerifiedPaymentNote = transaction.VerifiedPaymentNote
type SpendablePaymentNote = transaction.SpendablePaymentNote
type PaymentNoteVerificationOptions = transaction.PaymentNoteVerificationOptions
type TransactionOutput = transaction.TransactionOutput

func EncodePaymentNote(data PaymentNoteData) (string, error) {
	return transaction.EncodePaymentNote(data)
}
func DecodePaymentNote(encoded string) (PaymentNoteData, error) {
	return transaction.DecodePaymentNote(encoded)
}
func PaymentNoteFromIPN(encoded string) (*PaymentNote, error) {
	return transaction.PaymentNoteFromIPN(encoded)
}
