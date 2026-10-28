// Package sdk composes local Invisi wallets, indexed Protocol RPC data and a caller-owned EVM provider.
package sdk

import (
	"context"
	"errors"
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/invisi-lab/invisi-sdk-go/core"
)

// Wallet holds permissions locally. Formatting and JSON never export private material.
type Wallet struct {
	identity common.Address
	address  string
	keys     core.ReceivingKeys
	material *core.InvisiWallet
}

func WalletFromPrivateKey(privateKey string) (*Wallet, error) {
	raw, e := core.DecodePrivateKey(privateKey)
	if e != nil {
		return nil, e
	}
	defer clear(raw)
	m, e := core.DeriveWallet(raw)
	if e != nil {
		return nil, e
	}
	a, e := core.EncodeLongAddress(core.ReceivingAddress{WalletIdentity: m.WalletIdentity, ViewingPublicKey: m.ViewingPublicKey, ClueKey: m.ClueKey})
	if e != nil {
		return nil, e
	}
	return &Wallet{common.BigToAddress(core.ElementToBigInt(m.WalletIdentity)), a, core.ReceivingKeys{ViewingPublicKey: append([]byte(nil), m.ViewingPublicKey...), ClueKey: append([]byte(nil), m.ClueKey...)}, m}, nil
}
func WalletFromAddress(address string) (*Wallet, error) {
	d, e := core.DecodeAddress(address)
	if e != nil {
		return nil, e
	}
	return &Wallet{identity: common.BigToAddress(core.ElementToBigInt(d.WalletIdentity)), address: address, keys: core.ReceivingKeys{ViewingPublicKey: d.ViewingPublicKey, ClueKey: d.ClueKey}}, nil
}
func (w *Wallet) WalletIdentity() common.Address { return w.identity }
func (w *Wallet) OwnerAddress() common.Address   { return w.identity }
func (w *Wallet) GetAddress() (string, error) {
	if len(w.keys.ViewingPublicKey) != 33 {
		return "", errors.New("resolved receiving keys required")
	}
	return w.address, nil
}
func (w *Wallet) GetShortAddress() string { return w.identity.Hex() }
func (w *Wallet) String() string          { return fmt.Sprintf("Wallet(%s)", w.identity.Hex()) }
func (w *Wallet) GoString() string        { return w.String() }
func (w *Wallet) MarshalJSON() ([]byte, error) {
	return []byte(`{"walletIdentity":"` + w.identity.Hex() + `"}`), nil
}
func (w *Wallet) GetViewer() (*ViewingCapability, error) {
	if w.material == nil {
		return nil, errors.New("viewing permission required")
	}
	return &ViewingCapability{w.identity, append([]byte(nil), w.material.ViewingPrivateKey...), append([]byte(nil), w.material.ViewingPublicKey...)}, nil
}
func (w *Wallet) GetDetector() (*Detector, error) {
	if w.material == nil {
		return nil, errors.New("detection permission required")
	}
	return &Detector{append([]byte(nil), w.material.DetectionKey...), append([]byte(nil), w.keys.ClueKey...)}, nil
}
func (w *Wallet) ExportAuthorization() (*WalletAuthorization, error) {
	if w.material == nil {
		return nil, errors.New("wallet authorization permission required")
	}
	return &WalletAuthorization{w.material}, nil
}
func (w *Wallet) ExportNullifier() *NullifierCapability { return &NullifierCapability{w.identity} }

type ViewingCapability struct {
	identity    common.Address
	key, public []byte
}

func NewViewingCapability(identity common.Address, key []byte, public []byte) (*ViewingCapability, error) {
	derived, e := core.ViewingPublicKey(key)
	if e != nil {
		return nil, e
	}
	if common.Bytes2Hex(derived) != common.Bytes2Hex(public) || identity == (common.Address{}) {
		return nil, errors.New("viewing key mismatch")
	}
	return &ViewingCapability{identity, append([]byte(nil), key...), append([]byte(nil), public...)}, nil
}
func (v *ViewingCapability) ExportKey() []byte    { return append([]byte(nil), v.key...) }
func (v *ViewingCapability) GetPublicKey() []byte { return append([]byte(nil), v.public...) }
func (v *ViewingCapability) String() string       { return "ViewingCapability([redacted])" }
func (v *ViewingCapability) GoString() string     { return v.String() }
func (v *ViewingCapability) MarshalJSON() ([]byte, error) {
	return []byte(`{"viewing":"[redacted]"}`), nil
}

type Detector struct{ key, clue []byte }

func (d *Detector) ExportKey() []byte            { return append([]byte(nil), d.key...) }
func (d *Detector) GetClueKey() []byte           { return append([]byte(nil), d.clue...) }
func (d *Detector) String() string               { return "Detector([redacted])" }
func (d *Detector) GoString() string             { return d.String() }
func (d *Detector) MarshalJSON() ([]byte, error) { return []byte(`{"detection":"[redacted]"}`), nil }

type WalletAuthorization struct{ material *core.InvisiWallet }

func (a *WalletAuthorization) Authorize(ctx context.Context, request core.WalletAuthorizationRequest) (core.WalletAuthorizationWitness, error) {
	if e := ctx.Err(); e != nil {
		return core.WalletAuthorizationWitness{}, e
	}
	return core.BuildWalletAuthorizationWitness(a.material, request)
}
func (a *WalletAuthorization) String() string   { return "WalletAuthorization([redacted])" }
func (a *WalletAuthorization) GoString() string { return a.String() }
func (a *WalletAuthorization) MarshalJSON() ([]byte, error) {
	return []byte(`{"authorization":"[redacted]"}`), nil
}

type NullifierCapability struct{ identity common.Address }

func (n *NullifierCapability) ForNote(note core.Note) (common.Hash, error) {
	if note.WalletIdentity != core.BigIntToElement(n.identity.Big()) {
		return common.Hash{}, errors.New("note owner mismatch")
	}
	id, e := core.ComputeNoteIdentity(note.WalletIdentity, note.Nonce)
	if e != nil {
		return common.Hash{}, e
	}
	leaf, e := core.ComputeNoteCommitment(note.ChainID, id, note.TokenAddress, note.Amount)
	if e != nil {
		return common.Hash{}, e
	}
	nf, e := core.ComputeNoteNullifier(note.ChainID, note.Nonce, leaf)
	return common.Hash(nf.Bytes()), e
}
func (v *ViewingCapability) decryptNote(data []byte, context core.NoteDecryptionContext) (core.Note, error) {
	context.WalletIdentity = core.BigIntToElement(v.identity.Big())
	return core.DecryptPrivateTransferNote(v.key, data, context)
}
func (v *ViewingCapability) decryptCredit(data []byte, context core.CreditNoteDecryptionContext) (core.CreditDeliveryState, error) {
	context.WalletIdentity = core.BigIntToElement(v.identity.Big())
	return core.DecryptCreditNote(v.key, data, context)
}
func fieldHash(e fr.Element) common.Hash { return common.Hash(e.Bytes()) }

func (w *Wallet) ViewingPublicKey() []byte { return append([]byte(nil), w.keys.ViewingPublicKey...) }
func (w *Wallet) ClueKey() []byte          { return append([]byte(nil), w.keys.ClueKey...) }
func (v *ViewingCapability) DecryptNote(envelope []byte, context core.NoteDecryptionContext) (core.Note, error) {
	return v.decryptNote(envelope, context)
}
func (v *ViewingCapability) DecryptCredit(envelope []byte, context core.CreditNoteDecryptionContext) (core.CreditDeliveryState, error) {
	return v.decryptCredit(envelope, context)
}

// RegistrationAuthorizationProvider signs the public registry EIP-712 message.
// ERC-1271 wallets may return arbitrary signature bytes.
type RegistrationAuthorizationProvider interface {
	WalletIdentity() common.Address
	SignTypedData(context.Context, apitypes.TypedData) ([]byte, error)
}

type RegistrationAuthorizationFunc struct {
	Owner common.Address
	Sign  func(context.Context, apitypes.TypedData) ([]byte, error)
}

func (a RegistrationAuthorizationFunc) WalletIdentity() common.Address { return a.Owner }
func (a RegistrationAuthorizationFunc) SignTypedData(ctx context.Context, data apitypes.TypedData) ([]byte, error) {
	if a.Sign == nil {
		return nil, errors.New("registration signing callback required")
	}
	return a.Sign(ctx, data)
}

// ExportRegistrationAuthorization grants ordinary registry signing separately
// from private circuit authorization. The key remains local.
func (w *Wallet) ExportRegistrationAuthorization() (*RegistrationAuthorization, error) {
	if w.material == nil {
		return nil, errors.New("registration authorization permission required")
	}
	return &RegistrationAuthorization{w.material}, nil
}

type RegistrationAuthorization struct{ material *core.InvisiWallet }

func (a *RegistrationAuthorization) WalletIdentity() common.Address {
	if a == nil || a.material == nil {
		return common.Address{}
	}
	return common.BigToAddress(core.ElementToBigInt(a.material.WalletIdentity))
}
func (a *RegistrationAuthorization) SignTypedData(ctx context.Context, data apitypes.TypedData) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if a == nil || a.material == nil {
		return nil, errors.New("registration authorization permission required")
	}
	if data.Domain.Name != "InvisiAddressRegistry" || data.Domain.Version != "1" || (data.PrimaryType != "RegisterReceivingKeys" && data.PrimaryType != "UpdateReceivingKeys") || data.Message["walletIdentity"] != common.BigToHash(a.WalletIdentity().Big()).Hex() {
		return nil, errors.New("invalid registration signing request")
	}
	digest, _, err := apitypes.TypedDataAndHash(data)
	if err != nil {
		return nil, err
	}
	key, err := crypto.ToECDSA(a.material.SpendingPrivateKey)
	if err != nil {
		return nil, err
	}
	signature, err := crypto.Sign(digest, key)
	if err != nil {
		return nil, err
	}
	signature[64] += 27
	return signature, nil
}
func (a *RegistrationAuthorization) String() string   { return "RegistrationAuthorization([redacted])" }
func (a *RegistrationAuthorization) GoString() string { return a.String() }
func (a *RegistrationAuthorization) MarshalJSON() ([]byte, error) {
	return []byte(`{"registrationAuthorization":"[redacted]"}`), nil
}
