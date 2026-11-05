package transaction

import (
	"context"
	"errors"
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/core"
)

// Wallet holds permissions locally. Formatting and JSON never export private material.
type PublicWallet struct {
	identity common.Address
	address  string
	keys     core.ReceivingKeys
}

// Wallet keeps private capabilities separate from address-only identity.
type Wallet struct {
	*PublicWallet
	material *core.InvisiWallet
}

// OwnerAddress is implemented only by SDK public and private wallet identities.
type WalletReference interface{ publicWallet() *PublicWallet }

func (w *PublicWallet) publicWallet() *PublicWallet { return w }

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
	a, e := core.EncodeLongAddress(core.ReceivingAddress{OwnerAddress: m.OwnerAddress, ViewingPublicKey: m.ViewingPublicKey, ClueKey: m.ClueKey})
	if e != nil {
		return nil, e
	}
	return &Wallet{&PublicWallet{common.BigToAddress(core.ElementToBigInt(m.OwnerAddress)), a, core.ReceivingKeys{ViewingPublicKey: append([]byte(nil), m.ViewingPublicKey...), ClueKey: append([]byte(nil), m.ClueKey...)}}, m}, nil
}

func PublicWalletFromAddress(address string) (*PublicWallet, error) {
	d, e := core.DecodeAddress(address)
	if e != nil {
		return nil, e
	}
	return &PublicWallet{identity: common.BigToAddress(core.ElementToBigInt(d.OwnerAddress)), address: address, keys: core.ReceivingKeys{ViewingPublicKey: d.ViewingPublicKey, ClueKey: d.ClueKey}}, nil
}

func (w *PublicWallet) OwnerAddress() common.Address { return w.identity }

func (w *PublicWallet) GetAddress() (string, error) {
	if len(w.keys.ViewingPublicKey) != 33 {
		return "", errors.New("resolved receiving keys required")
	}
	return w.address, nil
}

func (w *PublicWallet) String() string { return fmt.Sprintf("Wallet(%s)", w.identity.Hex()) }

func (w *PublicWallet) GoString() string { return w.String() }

func (w *PublicWallet) MarshalJSON() ([]byte, error) {
	return []byte(`{"ownerAddress":"` + w.identity.Hex() + `"}`), nil
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

func (w *PublicWallet) ExportNullifier() *NullifierCapability {
	return &NullifierCapability{w.identity}
}

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

func (v *ViewingCapability) ExportKey() []byte { return append([]byte(nil), v.key...) }

func (v *ViewingCapability) GetPublicKey() []byte { return append([]byte(nil), v.public...) }

func (v *ViewingCapability) String() string { return "ViewingCapability([redacted])" }

func (v *ViewingCapability) GoString() string { return v.String() }

func (v *ViewingCapability) MarshalJSON() ([]byte, error) {
	return []byte(`{"viewing":"[redacted]"}`), nil
}

type Detector struct{ key, clue []byte }

func (d *Detector) ExportKey() []byte { return append([]byte(nil), d.key...) }

func (d *Detector) GetClueKey() []byte { return append([]byte(nil), d.clue...) }

func (d *Detector) String() string { return "Detector([redacted])" }

func (d *Detector) GoString() string { return d.String() }

func (d *Detector) MarshalJSON() ([]byte, error) { return []byte(`{"detection":"[redacted]"}`), nil }

type WalletAuthorization struct{ material *core.InvisiWallet }

func (a *WalletAuthorization) Authorize(ctx context.Context, request core.WalletAuthorizationRequest) (core.WalletAuthorizationWitness, error) {
	if e := ctx.Err(); e != nil {
		return core.WalletAuthorizationWitness{}, e
	}
	return core.BuildWalletAuthorizationWitness(a.material, request)
}

func (a *WalletAuthorization) String() string { return "WalletAuthorization([redacted])" }

func (a *WalletAuthorization) GoString() string { return a.String() }

func (a *WalletAuthorization) MarshalJSON() ([]byte, error) {
	return []byte(`{"authorization":"[redacted]"}`), nil
}

type NullifierCapability struct{ identity common.Address }

func (n *NullifierCapability) ForNote(note core.Note) (common.Hash, error) {
	if note.OwnerAddress != core.BigIntToElement(n.identity.Big()) {
		return common.Hash{}, errors.New("note owner mismatch")
	}
	id, e := core.ComputeNoteIdentity(note.OwnerAddress, note.Nonce)
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
	context.OwnerAddress = core.BigIntToElement(v.identity.Big())
	return core.DecryptPrivateTransferNote(v.key, data, context)
}

func (v *ViewingCapability) decryptCredit(data []byte, context core.CreditNoteDecryptionContext) (core.CreditDeliveryState, error) {
	context.OwnerAddress = core.BigIntToElement(v.identity.Big())
	return core.DecryptCreditNote(v.key, data, context)
}

func FieldHash(e fr.Element) common.Hash { return common.Hash(e.Bytes()) }

func (w *PublicWallet) ViewingPublicKey() []byte {
	return append([]byte(nil), w.keys.ViewingPublicKey...)
}

func (w *PublicWallet) ClueKey() []byte { return append([]byte(nil), w.keys.ClueKey...) }

func (v *ViewingCapability) DecryptNote(envelope []byte, context core.NoteDecryptionContext) (core.Note, error) {
	return v.decryptNote(envelope, context)
}

func (v *ViewingCapability) DecryptCredit(envelope []byte, context core.CreditNoteDecryptionContext) (core.CreditDeliveryState, error) {
	return v.decryptCredit(envelope, context)
}
