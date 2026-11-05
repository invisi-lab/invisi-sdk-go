package sdk

import (
	transaction "github.com/invisi-lab/invisi-sdk-go/internal/transaction"
)

type EVMRecipient = transaction.EVMRecipient

type IdentityRecipient = transaction.IdentityRecipient

type InvisiRecipient = transaction.InvisiRecipient

func NewNoteOutput(o NoteOutputOptions) (NoteOutput, error) { return transaction.NewNoteOutput(o) }

type NoteInput = transaction.NoteInput

type NoteOutput = transaction.NoteOutput

type NoteOutputContext = transaction.NoteOutputContext

type NoteOutputOptions = transaction.NoteOutputOptions

type NoteRecipient = transaction.NoteRecipient
