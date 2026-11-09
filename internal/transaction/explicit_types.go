package transaction

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
)

// NoteInput is an application-selected private note and its authenticated tree position.
// The application owns reservations and persistence; the SDK checks its preimage and membership.
type NoteInput struct {
	Note       core.Note
	Commitment common.Hash
	TreeID     uint64
	LeafIndex  uint32
}

type TransactionIdentity struct {
	Wallet   *Wallet
	Protocol *protocol.Protocol
}

type ExplicitFunds struct {
	Inputs        []NoteInput
	ChangeAddress string
}

type ExplicitTransferPreparation struct {
	TransactionIdentity
	TransferPreparation
	ExplicitFunds
}

type ExplicitShieldPreparation struct {
	TransactionIdentity
	ShieldPreparation
}
