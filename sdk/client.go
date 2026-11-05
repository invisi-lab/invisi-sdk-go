package sdk

import (
	transaction "github.com/invisi-lab/invisi-sdk-go/internal/transaction"
)

type Invisi = transaction.Invisi

type InvisiOptions = transaction.InvisiOptions

func NewInvisi(options InvisiOptions) *Invisi { return transaction.NewInvisi(options) }

type Provider = transaction.Provider
