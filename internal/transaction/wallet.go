package transaction

import (
	"errors"
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-lab/invisi-sdk-go/credit"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
)

// Wallet contains public identity and receiving material only.
type Wallet struct {
	ownerAddress  common.Address
	receivingKeys *primitives.ReceivingKeys
}
type WalletReference interface{ publicWallet() *Wallet }

func (w *Wallet) publicWallet() *Wallet { return w }

func WalletFromAddress(address string) (*Wallet, error) {
	d, e := primitives.DecodeAddress(address)
	if e != nil {
		return nil, e
	}
	owner := common.BigToAddress(primitives.ElementToBigInt(d.OwnerAddress))
	if d.Kind == "short" {
		return nil, errors.New("complete receiving address required")
	}
	return newPublicWallet(owner, &primitives.ReceivingKeys{ViewingPublicKey: d.ViewingPublicKey, ClueKey: d.ClueKey})
}

func (w *Wallet) OwnerAddress() common.Address { return w.ownerAddress }

func (w *Wallet) GetAddress() (string, error) {
	if w.receivingKeys == nil {
		return "", errors.New("resolved receiving keys required")
	}
	return primitives.EncodeLongAddress(primitives.ReceivingAddress{OwnerAddress: primitives.BigIntToElement(w.ownerAddress.Big()), ViewingPublicKey: w.receivingKeys.ViewingPublicKey, ClueKey: w.receivingKeys.ClueKey})
}

func (w *Wallet) String() string { return fmt.Sprintf("Wallet(%s)", w.ownerAddress.Hex()) }

func (w *Wallet) GoString() string { return w.String() }

func (w *Wallet) MarshalJSON() ([]byte, error) {
	return []byte(`{"ownerAddress":"` + w.ownerAddress.Hex() + `"}`), nil
}

func (w *Wallet) ExportNullifier() *NullifierCapability {
	return &NullifierCapability{w.ownerAddress}
}

type ViewingCapability struct {
	identity    common.Address
	key, public []byte
}

func NewViewingCapability(identity common.Address, key []byte, public []byte) (*ViewingCapability, error) {
	derived, e := primitives.ViewingPublicKey(key)
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

func (n *NullifierCapability) ForNote(note primitives.Note) (common.Hash, error) {
	if note.OwnerAddress != primitives.BigIntToElement(n.identity.Big()) {
		return common.Hash{}, errors.New("note owner mismatch")
	}
	id, e := primitives.ComputeNoteIdentity(note.OwnerAddress, note.Nonce)
	if e != nil {
		return common.Hash{}, e
	}
	leaf, e := primitives.ComputeNoteCommitment(note.ChainID, note.ProtocolAddress, id, note.TokenAddress, note.Amount)
	if e != nil {
		return common.Hash{}, e
	}
	nf, e := primitives.ComputeNoteNullifier(note.ChainID, note.ProtocolAddress, note.Nonce, leaf)
	return common.Hash(nf.Bytes()), e
}

func (v *ViewingCapability) decryptNote(data []byte, context primitives.NoteDecryptionContext) (primitives.Note, error) {
	context.OwnerAddress = primitives.BigIntToElement(v.identity.Big())
	return primitives.DecryptPrivateTransferNote(v.key, data, context)
}

func FieldHash(e fr.Element) common.Hash { return common.Hash(e.Bytes()) }

func (w *Wallet) ViewingPublicKey() []byte {
	if w.receivingKeys == nil {
		return nil
	}
	return append([]byte(nil), w.receivingKeys.ViewingPublicKey...)
}

func (w *Wallet) ClueKey() []byte {
	if w.receivingKeys == nil {
		return nil
	}
	return append([]byte(nil), w.receivingKeys.ClueKey...)
}

func (v *ViewingCapability) DecryptNote(envelope []byte, context primitives.NoteDecryptionContext) (primitives.Note, error) {
	return v.decryptNote(envelope, context)
}

func (v *ViewingCapability) DecryptCreditState(envelope []byte, context credit.DeliveryContext, role credit.Role) (credit.DeliveryState, error) {
	if v == nil || len(v.key) == 0 {
		return credit.DeliveryState{}, errors.New("viewing capability required")
	}
	return credit.Decrypt(v.key, envelope, context, role)
}

func newPublicWallet(owner common.Address, keys *primitives.ReceivingKeys) (*Wallet, error) {
	if _, err := primitives.EncodeShortAddress(primitives.BigIntToElement(owner.Big())); err != nil {
		return nil, err
	}
	if keys == nil {
		return nil, errors.New("complete receiving keys required")
	}
	if _, err := primitives.EncodeLongAddress(primitives.ReceivingAddress{OwnerAddress: primitives.BigIntToElement(owner.Big()), ViewingPublicKey: keys.ViewingPublicKey, ClueKey: keys.ClueKey}); err != nil {
		return nil, err
	}
	return &Wallet{ownerAddress: owner, receivingKeys: &primitives.ReceivingKeys{ViewingPublicKey: append([]byte(nil), keys.ViewingPublicKey...), ClueKey: append([]byte(nil), keys.ClueKey...)}}, nil
}
