package core

import (
	"bytes"
	"crypto/rand"
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

type CreditDeliveryRole string

const (
	CreditOwner   CreditDeliveryRole = "owner"
	CreditSpender CreditDeliveryRole = "spender"
)

type CreditNoteEncryptionContext struct {
	NoteEncryptionContext
	Role CreditDeliveryRole
}

type CreditNoteDecryptionContext struct {
	NoteDecryptionContext
	Role CreditDeliveryRole
}

func creditRole(role CreditDeliveryRole) (byte, error) {
	switch role {
	case CreditOwner:
		return 1, nil
	case CreditSpender:
		return 2, nil
	default:
		return 0, fmt.Errorf("Credit role must be owner or spender")
	}
}

func encodeCredit(note CreditNote) []byte {
	out := appendElement(nil, note.SpenderWalletIdentity, 32)
	out = appendElement(out, note.OwnerWalletIdentity, 32)
	out = appendElement(out, note.SpenderPayout, 20)
	out = appendElement(out, note.OwnerPayout, 20)
	out = appendElement(out, note.Amount, 16)
	out = appendElement(out, note.Nonce, 32)
	out = appendElement(out, note.Expiry, 8)
	return appendElement(out, note.CreditSharedKey, 32)
}

// NormalizeCreditDeliveryState validates and copies the opaque payload and receiving snapshots.
func NormalizeCreditDeliveryState(state CreditDeliveryState) (CreditDeliveryState, error) {
	if err := ValidatePayload(state.Note.Payload); err != nil {
		return CreditDeliveryState{}, err
	}
	owner, err := walletAddressKeys(state.OwnerReceivingKeys)
	if err != nil {
		return CreditDeliveryState{}, err
	}
	spender, err := walletAddressKeys(state.SpenderReceivingKeys)
	if err != nil {
		return CreditDeliveryState{}, err
	}
	return CreditDeliveryState{Note: state.Note.Clone(), OwnerReceivingKeys: owner, SpenderReceivingKeys: spender}, nil
}

func matchCreditRecipient(state CreditDeliveryState, role byte, public []byte) error {
	keys := state.OwnerReceivingKeys
	if role == 2 {
		keys = state.SpenderReceivingKeys
	}
	if !bytes.Equal(keys.ViewingPublicKey, public) {
		return fmt.Errorf("Credit recipient key differs from its snapshot")
	}
	return nil
}

func encodeCreditState(state CreditDeliveryState) []byte {
	out := encodeCredit(state.Note)
	out = append(out, state.OwnerReceivingKeys.ViewingPublicKey...)
	out = append(out, state.OwnerReceivingKeys.ClueKey...)
	out = append(out, state.SpenderReceivingKeys.ViewingPublicKey...)
	out = append(out, state.SpenderReceivingKeys.ClueKey...)
	for i, value := range state.Note.CreditSpendingState.Fields() {
		size := 16
		if i == 3 || i == 5 {
			size = 8
		}
		out = appendElement(out, value, size)
	}
	return appendPayload(out, state.Note.Payload)
}

// EncryptCreditNote creates a v1 role-bound credit envelope with complete receiving snapshots.
func EncryptCreditNote(viewingKey []byte, state CreditDeliveryState, context CreditNoteEncryptionContext) ([]byte, error) {
	role, err := creditRole(context.Role)
	if err != nil {
		return nil, err
	}
	state, err = NormalizeCreditDeliveryState(state)
	if err != nil {
		return nil, err
	}
	if err = matchCreditRecipient(state, role, viewingKey); err != nil {
		return nil, err
	}
	if err = verifyCredit(context.NoteEncryptionContext, state.Note); err != nil {
		return nil, err
	}
	plain := encodeCreditState(state)
	if len(plain)+78 > MaxCiphertextBytes {
		return nil, fmt.Errorf("envelope exceeds resource bound")
	}
	defer clear(plain)
	return sealEnvelope(viewingKey, plain, context.NoteEncryptionContext, state.Note.TokenAddress, 0x81, role, rand.Reader)
}

func DecryptCreditNote(private, envelope []byte, context CreditNoteDecryptionContext) (CreditDeliveryState, error) {
	role, err := creditRole(context.Role)
	if err != nil {
		return CreditDeliveryState{}, err
	}
	plain, err := openEnvelope(private, envelope, context.NoteDecryptionContext, 0x81, role)
	if err != nil {
		return CreditDeliveryState{}, err
	}
	defer clear(plain)
	public, err := ViewingPublicKey(private)
	if err != nil {
		return CreditDeliveryState{}, err
	}
	return readCreditPlain(plain, context, public)
}

func readCreditPlain(plain []byte, context CreditNoteDecryptionContext, public []byte) (CreditDeliveryState, error) {
	memo, err := readPayload(plain, 418)
	if err != nil {
		return CreditDeliveryState{}, fmt.Errorf("invalid Credit plaintext: %w", err)
	}
	role, err := creditRole(context.Role)
	if err != nil {
		return CreditDeliveryState{}, err
	}
	note := CreditNote{ProtocolContext: context.ProtocolContext, TokenAddress: context.TokenAddress, Payload: memo}
	fields := []*fr.Element{&note.SpenderWalletIdentity, &note.OwnerWalletIdentity, &note.SpenderPayout, &note.OwnerPayout, &note.Amount, &note.Nonce, &note.Expiry, &note.CreditSharedKey}
	offsets := []int{0, 32, 64, 84, 104, 120, 152, 160, 192}
	for i, target := range fields {
		value, err := decodeElement(plain[offsets[i]:offsets[i+1]])
		if err != nil {
			return CreditDeliveryState{}, err
		}
		*target = value
	}
	stateFields := []*fr.Element{&note.SingleLimit, &note.DailyLimit, &note.MonthlyLimit, &note.LastDay, &note.DailySpent, &note.LastMonth, &note.MonthlySpent}
	offset := 322
	for i, target := range stateFields {
		size := 16
		if i == 3 || i == 5 {
			size = 8
		}
		value, e := decodeElement(plain[offset : offset+size])
		if e != nil {
			return CreditDeliveryState{}, e
		}
		*target = value
		offset += size
	}
	expected := note.OwnerWalletIdentity
	if role == 2 {
		expected = note.SpenderWalletIdentity
	}
	if expected != context.OwnerAddress {
		return CreditDeliveryState{}, fmt.Errorf("recipient owner address mismatch")
	}
	if err = verifyCredit(context.NoteEncryptionContext, note); err != nil {
		return CreditDeliveryState{}, err
	}
	state, err := NormalizeCreditDeliveryState(CreditDeliveryState{Note: note,
		OwnerReceivingKeys:   ReceivingKeys{ViewingPublicKey: plain[192:225], ClueKey: plain[225:257]},
		SpenderReceivingKeys: ReceivingKeys{ViewingPublicKey: plain[257:290], ClueKey: plain[290:322]}})
	if err != nil {
		return CreditDeliveryState{}, err
	}
	if public != nil {
		if err = matchCreditRecipient(state, role, public); err != nil {
			return CreditDeliveryState{}, err
		}
	}
	return state, nil
}
