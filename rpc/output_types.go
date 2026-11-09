package rpc

import "github.com/ethereum/go-ethereum/common"

// ProtocolNoteOutput is an ordinary Note delivery, never a Credit authority state.
type ProtocolNoteOutput struct {
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
type ProtocolNoteUpdates struct {
	Block CanonicalBlock
	Notes []ProtocolNoteOutput
}
type IndexedDelivery struct{ Ciphertext, Clue []byte }

// IndexedCreditState retains the two public owner/spender envelopes of one physical leaf.
type IndexedCreditState struct {
	Block           CanonicalBlock
	TreeID          uint64
	LeafIndex       uint32
	Commitment      common.Hash
	TokenAddress    common.Address
	Deliveries      []IndexedDelivery
	TransactionHash common.Hash
	LogIndex        uint32
	Kind            string
}
type ProtocolCreditCandidates struct {
	Block  CanonicalBlock
	States []IndexedCreditState
}
type ProtocolCreditUpdates struct {
	Block       CanonicalBlock
	Transitions []ProtocolCreditTransition
}
type RegisteredReceivingAddress struct {
	ReceivingAddress string
	Nonce            uint64
	Block            CanonicalBlock
}
