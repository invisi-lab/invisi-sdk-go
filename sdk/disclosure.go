package sdk

import (
	"github.com/invisi-lab/invisi-sdk-go/core"
	transaction "github.com/invisi-lab/invisi-sdk-go/internal/transaction"
)

type CreditOpening = transaction.CreditOpening

type Disclosure = transaction.Disclosure

type DisclosureContext = transaction.DisclosureContext

type DisclosureEvidence = transaction.DisclosureEvidence

type DisclosureExport = transaction.DisclosureExport

func DisclosureFromCredit(note core.CreditNote) (*Disclosure, error) {
	return transaction.DisclosureFromCredit(note)
}

func DisclosureFromExport(data []byte) (*Disclosure, error) {
	return transaction.DisclosureFromExport(data)
}

func DisclosureFromNote(note core.Note) (*Disclosure, error) {
	return transaction.DisclosureFromNote(note)
}

type DisclosureOutput = transaction.DisclosureOutput

type DisclosureReference = transaction.DisclosureReference

type DisclosureResult = transaction.DisclosureResult

type DisclosureScope = transaction.DisclosureScope

type DisclosureVerifier = transaction.DisclosureVerifier

type DisclosureVerifierOptions = transaction.DisclosureVerifierOptions

func NewDisclosureVerifier(options DisclosureVerifierOptions) (*DisclosureVerifier, error) {
	return transaction.NewDisclosureVerifier(options)
}

type VerifiedDisclosure = transaction.VerifiedDisclosure
