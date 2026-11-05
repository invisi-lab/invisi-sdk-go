package sdk

import (
	"github.com/invisi-lab/invisi-sdk-go/broadcaster"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
)

type BroadcasterClient = broadcaster.Client
type BroadcasterOptions = protocol.BroadcasterOptions
type BroadcasterConfiguration = broadcaster.Configuration
type BroadcasterClientOptions = broadcaster.Options

const MaxUpdateBlocks = protocolrpc.MaxUpdateBlocks

func NewBroadcasterClient(config BroadcasterConfiguration, options ...BroadcasterClientOptions) (*BroadcasterClient, error) {
	return broadcaster.NewClient(config, options...)
}
