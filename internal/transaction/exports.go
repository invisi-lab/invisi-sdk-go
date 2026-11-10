package transaction

import (
	"context"

	"github.com/invisi-lab/invisi-sdk-go/rpc/chain"

	"github.com/invisi-lab/invisi-sdk-go/prover"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
)

type Protocol = chain.Protocol

type ProtocolRpcClient = protocolrpc.ProtocolRpcClient

type ProtocolTransactionNoteUpdates = protocolrpc.ProtocolTransactionNoteUpdates
type ProtocolTransactionCreditUpdates = protocolrpc.ProtocolTransactionCreditUpdates

type TransactionUpdateStatus = protocolrpc.TransactionUpdateStatus

const (
	TransactionNotFound     = protocolrpc.TransactionNotFound
	TransactionPendingIndex = protocolrpc.TransactionPendingIndex
	TransactionIndexed      = protocolrpc.TransactionIndexed
)

type Prover = prover.Prover

type ProverOptions = prover.Options

type ProofBundle = prover.ProofBundle

type ProtocolDeployment = protocolrpc.ProtocolDeployment

func NewProtocolRpcClient(endpoint string, options ...protocolrpc.ClientOptions) (*ProtocolRpcClient, error) {
	return protocolrpc.NewProtocolRpcClient(endpoint, options...)
}

func ProtocolFromRPC(ctx context.Context, rpc *ProtocolRpcClient, provider protocolrpc.ChainProvider) (*Protocol, error) {
	return chain.ProtocolFromRPC(ctx, rpc, provider)
}

func NewProtocol(deployment ProtocolDeployment) (*Protocol, error) {
	return chain.NewProtocol(deployment)
}
