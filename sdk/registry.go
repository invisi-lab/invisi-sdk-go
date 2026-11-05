package sdk

import (
	transaction "github.com/invisi-lab/invisi-sdk-go/internal/transaction"
)

type AddressRegistry = transaction.AddressRegistry

func NewAddressRegistry(pool *Protocol, provider Provider) (*AddressRegistry, error) {
	return transaction.NewAddressRegistry(pool, provider)
}

type RegistrationOperation = transaction.RegistrationOperation

type RegistrationPreparation = transaction.RegistrationPreparation

type RegistrationPreview = transaction.RegistrationPreview
