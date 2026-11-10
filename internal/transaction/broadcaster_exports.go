package transaction

import (
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
	"github.com/invisi-lab/invisi-sdk-go/rpc/broadcaster"
	"github.com/invisi-lab/invisi-sdk-go/rpc/chain"
)

type BroadcasterClient = broadcaster.Client

type BroadcasterOptions = chain.BroadcasterOptions

type BroadcasterConfiguration = broadcaster.Configuration

type BroadcasterClientOptions = broadcaster.Options

const MaxUpdateBlocks = protocolrpc.MaxUpdateBlocks

func NewBroadcasterClient(config BroadcasterConfiguration, options ...BroadcasterClientOptions) (*BroadcasterClient, error) {
	return broadcaster.NewClient(config, options...)
}
