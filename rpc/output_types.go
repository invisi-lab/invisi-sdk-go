package rpc

import (
	"github.com/ethereum/go-ethereum/common"
)

// ProtocolOutput is an indexed opaque delivery. Authenticate decryption and membership before spending.
type ProtocolOutput struct {
	Block            CanonicalBlock
	TreeID           uint64
	LeafIndex        uint32
	Commitment       common.Hash
	TokenAddress     common.Address
	Ciphertext, Clue []byte
	TransactionHash  common.Hash
	LogIndex         uint32
	Kind             string
}

type ProtocolOutputs struct {
	Block   CanonicalBlock
	Outputs []ProtocolOutput
}

type RegisteredReceivingAddress struct {
	ReceivingAddress string
	Nonce            uint64
	Block            CanonicalBlock
}
