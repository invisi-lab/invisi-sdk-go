package transaction

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-lab/invisi-sdk-go/core"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
)

type WalletScope struct {
	ChainID                       uint64
	ProtocolAddress, OwnerAddress common.Address
	Namespace                     string
}

func (s WalletScope) Key() string {
	b, _ := json.Marshal(s)
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

type NoteRecord struct {
	TransactionHash       common.Hash
	Block                 *protocolrpc.CanonicalBlock
	Note                  core.Note
	Commitment, Nullifier common.Hash
	TreeID                uint64
	LeafIndex             uint32
	Spent                 bool
	ReservedBy            string
}
