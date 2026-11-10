package transaction

import (
	"errors"
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/credit"
)

// Wallet contains public identity and receiving material only.
type Wallet struct {
	identity common.Address
	address  string
	keys     core.ReceivingKeys
}
type WalletReference interface{ publicWallet() *Wallet }

func (w *Wallet) publicWallet() *Wallet { return w }

func WalletFromAddress(address string) (*Wallet, error) {
	d, e := core.DecodeAddress(address)
	if e != nil {
		return nil, e
	}
	return &Wallet{identity: common.BigToAddress(core.ElementToBigInt(d.OwnerAddress)), address: address, keys: core.ReceivingKeys{ViewingPublicKey: d.ViewingPublicKey, ClueKey: d.ClueKey}}, nil
}

func (w *Wallet) OwnerAddress() common.Address { return w.identity }

func (w *Wallet) GetAddress() (string, error) {
	if len(w.keys.ViewingPublicKey) != 33 {
		return "", errors.New("resolved receiving keys required")
	}
	return w.address, nil
}

func (w *Wallet) String() string { return fmt.Sprintf("Wallet(%s)", w.identity.Hex()) }

func (w *Wallet) GoString() string { return w.String() }

func (w *Wallet) MarshalJSON() ([]byte, error) {
	return []byte(`{"ownerAddress":"` + w.identity.Hex() + `"}`), nil
}

func (w *Wallet) ExportNullifier() *NullifierCapability {
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
	return &ViewingCapability{identity: identity, key: append([]byte(nil), key...), public: append([]byte(nil), public...)}, nil
}

func (v *ViewingCapability) ExportKey() []byte { return append([]byte(nil), v.key...) }

func (v *ViewingCapability) GetPublicKey() []byte { return append([]byte(nil), v.public...) }

func (v *ViewingCapability) String() string { return "ViewingCapability([redacted])" }

func (v *ViewingCapability) GoString() string { return v.String() }

func (v *ViewingCapability) MarshalJSON() ([]byte, error) {
	return []byte(`{"viewing":"[redacted]"}`), nil
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
	leaf, e := core.ComputeNoteCommitment(note.ChainID, note.ProtocolAddress, id, note.TokenAddress, note.Amount)
	if e != nil {
		return common.Hash{}, e
	}
	nf, e := core.ComputeNoteNullifier(note.ChainID, note.ProtocolAddress, note.Nonce, leaf)
	return common.Hash(nf.Bytes()), e
}

func (v *ViewingCapability) decryptNote(data []byte, context core.NoteDecryptionContext) (core.Note, error) {
	context.OwnerAddress = core.BigIntToElement(v.identity.Big())
	return core.DecryptPrivateTransferNote(v.key, data, context)
}

func FieldHash(e fr.Element) common.Hash { return common.Hash(e.Bytes()) }

func (w *Wallet) ViewingPublicKey() []byte {
	return append([]byte(nil), w.keys.ViewingPublicKey...)
}

func (w *Wallet) ClueKey() []byte { return append([]byte(nil), w.keys.ClueKey...) }

func (v *ViewingCapability) DecryptNote(envelope []byte, context core.NoteDecryptionContext) (core.Note, error) {
	return v.decryptNote(envelope, context)
}

func (v *ViewingCapability) DecryptCreditState(envelope []byte, context credit.DeliveryContext, role credit.Role) (credit.DeliveryState, error) {
	if v == nil || len(v.key) == 0 {
		return credit.DeliveryState{}, errors.New("viewing capability required")
	}
	return credit.Decrypt(v.key, envelope, context, role)
}
