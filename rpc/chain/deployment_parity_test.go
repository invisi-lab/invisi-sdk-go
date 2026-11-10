package chain

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-lab/invisi-sdk-go/rpc"
)

func TestBroadcasterRegistryDeploymentSnapshot(t *testing.T) {
	registry := rpc.ContractDeployment{Address: common.HexToAddress("0x1"), RuntimeCodeHash: common.HexToHash("0x1"), FromBlock: 1}
	d := rpc.ProtocolDeployment{ChainID: "1", Protocol: registry, BroadcasterRegistry: &registry}
	p, err := NewProtocol(d)
	if err != nil {
		t.Fatal(err)
	}
	registry.Address = common.HexToAddress("0x2")
	snapshot := p.Deployment()
	if snapshot.BroadcasterRegistry.Address != common.HexToAddress("0x1") {
		t.Fatal("constructor shares registry pointer")
	}
	snapshot.BroadcasterRegistry.Address = common.HexToAddress("0x3")
	if p.Deployment().BroadcasterRegistry.Address != common.HexToAddress("0x1") {
		t.Fatal("getter shares registry pointer")
	}
}
