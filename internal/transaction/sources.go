package transaction

import (
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
	"math/big"
	"sort"
)

func SourcesFor(memberships []protocolrpc.Membership) ([]protocol.InputSource, error) {
	roots := map[uint64]common.Hash{}
	for _, m := range memberships {
		if root, ok := roots[m.TreeID]; ok && root != m.Root {
			return nil, errors.New("inconsistent roots for input tree")
		}
		roots[m.TreeID] = m.Root
	}
	ids := []uint64{}
	for id := range roots {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	result := []protocol.InputSource{}
	for _, id := range ids {
		result = append(result, protocol.InputSource{TreeId: new(big.Int).SetUint64(id), MerkleRoot: roots[id]})
	}
	return result, nil
}
