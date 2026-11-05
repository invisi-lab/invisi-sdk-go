package sdk

import (
	"context"

	transaction "github.com/invisi-lab/invisi-sdk-go/internal/transaction"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
)

const MaxUpdateBlocks = transaction.MaxUpdateBlocks

func NewProtocol(deployment ProtocolDeployment) (*Protocol, error) {
	return transaction.NewProtocol(deployment)
}

func NewProtocolRpcClient(endpoint string, options ...protocolrpc.ClientOptions) (*ProtocolRpcClient, error) {
	return transaction.NewProtocolRpcClient(endpoint, options...)
}

type Protocol = transaction.Protocol

type ProtocolCreditTransition = transaction.ProtocolCreditTransition

type ProtocolDeployment = transaction.ProtocolDeployment

func ProtocolFromRPC(ctx context.Context, rpc *ProtocolRpcClient, provider protocolrpc.ChainProvider) (*Protocol, error) {
	return transaction.ProtocolFromRPC(ctx, rpc, provider)
}

type ProtocolRpcClient = transaction.ProtocolRpcClient

type ProtocolTransactionUpdates = transaction.ProtocolTransactionUpdates

type TransactionUpdateStatus = transaction.TransactionUpdateStatus

const TransactionNotFound = transaction.TransactionNotFound

const TransactionPendingIndex = transaction.TransactionPendingIndex

const TransactionIndexed = transaction.TransactionIndexed
