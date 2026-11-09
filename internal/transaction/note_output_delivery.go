package transaction

import (
	"context"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
)

func EncryptOutput(recipient core.ReceivingAddress, note core.Note, commitment common.Hash) (protocol.Delivery, error) {
	cipher, e := core.EncryptPrivateTransferNote(recipient.ViewingPublicKey, note, core.NoteEncryptionContext{ProtocolContext: note.ProtocolContext, OutputCommitment: core.BigIntToElement(commitment.Big())})
	if e != nil {
		return protocol.Delivery{}, e
	}
	clue, e := core.CreateClue(recipient.ClueKey)
	return protocol.Delivery{Ciphertext: cipher, Clue: clue}, e
}

func (s *Runtime) Output(ctx context.Context, token common.Address, n NoteOutput) (PreparedOutput, error) {
	if n.saved == nil {
		return PreparedOutput{}, errors.New("outputs must be created with NewNoteOutput")
	}
	p := n.saved
	if e := matchPaymentRequest(p); e != nil {
		return PreparedOutput{}, e
	}
	if p.Context != s.CoreContext() || p.Token != token {
		return PreparedOutput{}, errors.New("note output chain, protocol or token mismatch")
	}
	commitment, e := core.ComputeNoteCommitment(p.Context.ChainID, p.Context.ProtocolAddress, core.BigIntToElement(p.Identity.Big()), core.BigIntToElement(token.Big()), core.BigIntToElement(p.Amount))
	if e != nil || FieldHash(commitment) != p.Commitment {
		return PreparedOutput{}, errors.New("note output commitment mismatch")
	}
	d := n.Delivery()
	if p.Pending {
		recipient, e := s.ResolveWithLookup(ctx, p.Address, p.Lookup)
		if e != nil {
			return PreparedOutput{}, e
		}
		d, e = EncryptOutput(recipient, *p.Note, p.Commitment)
		if e != nil {
			return PreparedOutput{}, e
		}
	}
	return PreparedOutput{Address: p.Address, Note: n.Note(), Amount: n.Amount(), Identity: p.Identity, Commitment: p.Commitment, Delivery: d}, nil
}

func (s *Runtime) ChangeOutput(ctx context.Context, token common.Address, address string, amount *big.Int) (PreparedOutput, error) {
	var r NoteRecipient = InvisiRecipient{Address: address}
	if common.IsHexAddress(address) {
		r = EVMRecipient{Address: common.HexToAddress(address)}
	}
	n, e := NewNoteOutput(PaymentRequest{ChainID: core.ElementToBigInt(s.CoreContext().ChainID), ProtocolAddress: s.Scope.ProtocolAddress, TokenAddress: token, Amount: amount, Recipient: r})
	if e != nil {
		return PreparedOutput{}, e
	}
	return s.Output(ctx, token, n)
}
