// Package rpc communicates with the versioned Invisi Protocol data service.
// Indexed responses are independently checked against the caller's Chain RPC.
package rpc

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type ContractDeployment struct {
	Address         common.Address `json:"address"`
	RuntimeCodeHash common.Hash    `json:"runtimeCodeHash"`
	FromBlock       uint64         `json:"fromBlock"`
}
type ProtocolDeployment struct {
	ChainID             string              `json:"chainId"`
	Protocol            ContractDeployment  `json:"protocol"`
	Registry            *ContractDeployment `json:"registry,omitempty"`
	BroadcasterRegistry *ContractDeployment `json:"broadcasterRegistry,omitempty"`
}

func (d ProtocolDeployment) ChainNumber() (*big.Int, error) {
	n, ok := new(big.Int).SetString(d.ChainID, 10)
	if !ok || n.Sign() <= 0 || n.BitLen() > 64 {
		return nil, fmt.Errorf("invalid deployment chainId")
	}
	return n, nil
}
func (d ProtocolDeployment) Validate() error {
	if _, e := d.ChainNumber(); e != nil {
		return e
	}
	for _, c := range []*ContractDeployment{&d.Protocol, d.Registry, d.BroadcasterRegistry} {
		if c != nil && (c.Address == (common.Address{}) || c.RuntimeCodeHash == (common.Hash{}) || c.FromBlock > 1<<53-1) {
			return fmt.Errorf("invalid contract deployment")
		}
	}
	return nil
}

type DeploymentContext struct {
	ChainID         uint64
	ProtocolAddress common.Address
}
type CanonicalBlock struct {
	Number     uint64      `json:"number"`
	Hash       common.Hash `json:"hash"`
	ParentHash common.Hash `json:"parentHash"`
	Timestamp  uint64      `json:"timestamp"`
}
type ChainProvider interface {
	ChainID(context.Context) (*big.Int, error)
	HeaderByNumber(context.Context, *big.Int) (*types.Header, error)
	CodeAt(context.Context, common.Address, *big.Int) ([]byte, error)
	CallContract(context.Context, ethereum.CallMsg, *big.Int) ([]byte, error)
}
type ProtocolDataStatus struct {
	IndexedHead      *CanonicalBlock
	DetectionEnabled bool
	CheckedBits      int
}
type ProtocolUpdateRange struct {
	FromBlock, ToBlock uint64
	DetectionKey       []byte
}
type TransactionReference struct {
	TransactionHash common.Hash    `json:"transactionHash"`
	LogIndex        uint64         `json:"logIndex"`
	Block           CanonicalBlock `json:"block"`
}
type MembershipInput struct {
	TreeID     uint64
	LeafIndex  uint32
	Commitment common.Hash
	Root       *common.Hash
}
type Membership struct {
	TreeID     uint64
	LeafIndex  uint32
	Leaf, Root common.Hash
	Path       []common.Hash
	Block      CanonicalBlock
}
type Chain struct {
	ChainID         string `json:"chainId"`
	ProtocolVersion string `json:"protocolVersion"`
	RPCPath         string `json:"rpcPath"`
	RPCURL          string `json:"-"`
}
type rootAnchor struct {
	TreeID    uint64         `json:"treeId"`
	Root      common.Hash    `json:"root"`
	LeafCount uint32         `json:"leafCount"`
	Block     CanonicalBlock `json:"block"`
}
