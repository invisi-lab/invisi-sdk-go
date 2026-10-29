package receipts

import (
	"encoding/json"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
)

type Block struct {
	Timestamp  uint64      `json:"timestamp"`
	Number     uint64      `json:"number"`
	Hash       common.Hash `json:"hash"`
	ParentHash common.Hash `json:"parentHash"`
}
type Insertion struct {
	TreeID      uint64        `json:"treeId"`
	Start       uint32        `json:"startLeafIndex"`
	Commitments []common.Hash `json:"commitments"`
	LogIndex    uint          `json:"logIndex"`
}
type Output struct {
	Block           Block          `json:"block"`
	TreeID          uint64         `json:"treeId"`
	LeafIndex       uint32         `json:"leafIndex"`
	Commitment      common.Hash    `json:"commitment"`
	Token           common.Address `json:"token"`
	Ciphertext      hexutil.Bytes  `json:"ciphertext"`
	Clue            hexutil.Bytes  `json:"clue"`
	TransactionHash common.Hash    `json:"transactionHash"`
	LogIndex        uint           `json:"logIndex"`
	Kind            string         `json:"kind"`
	Role            string         `json:"role"`
}
type Event struct {
	Block           Block           `json:"block"`
	Name            string          `json:"name"`
	TransactionHash common.Hash     `json:"transactionHash"`
	LogIndex        uint            `json:"logIndex"`
	Data            json.RawMessage `json:"data"`
}
type Batch struct {
	Block        Block         `json:"block"`
	Logs         []types.Log   `json:"logs"`
	Events       []Event       `json:"events"`
	CreatedTrees []uint64      `json:"createdTrees"`
	Insertions   []Insertion   `json:"insertions"`
	Outputs      []Output      `json:"outputs"`
	Nullifiers   []common.Hash `json:"nullifiers"`
}
