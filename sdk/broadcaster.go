package sdk

import (
	transaction "github.com/invisi-lab/invisi-sdk-go/internal/transaction"
)

type BroadcasterClient = transaction.BroadcasterClient

type BroadcasterClientOptions = transaction.BroadcasterClientOptions

type BroadcasterConfiguration = transaction.BroadcasterConfiguration

type BroadcasterFeeSource = transaction.BroadcasterFeeSource

type BroadcasterOptions = transaction.BroadcasterOptions

func NewBroadcasterClient(config BroadcasterConfiguration, options ...BroadcasterClientOptions) (*BroadcasterClient, error) {
	return transaction.NewBroadcasterClient(config, options...)
}

type WithBroadcasterOptions = transaction.WithBroadcasterOptions

const BroadcasterFeePublicFunding = transaction.BroadcasterFeePublicFunding

const BroadcasterFeePrivateInputs = transaction.BroadcasterFeePrivateInputs

const BroadcasterFeePayout = transaction.BroadcasterFeePayout
