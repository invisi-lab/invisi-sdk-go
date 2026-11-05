package sdk

import (
	transaction "github.com/invisi-lab/invisi-sdk-go/internal/transaction"
)

type InvisiError = transaction.InvisiError

type InvisiErrorCode = transaction.InvisiErrorCode

type MissingCapabilityError = transaction.MissingCapabilityError

const PreparationFailed = transaction.PreparationFailed

const MissingCapability = transaction.MissingCapability

const InvalidRecipient = transaction.InvalidRecipient

const EVMAddressNotRegistered = transaction.EVMAddressNotRegistered

const RecipientLookupFailed = transaction.RecipientLookupFailed
