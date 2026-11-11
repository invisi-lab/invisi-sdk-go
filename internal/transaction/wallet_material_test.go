package transaction

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
	"reflect"
	"testing"
)

func TestWalletMaterialLayout(t *testing.T) {
	typ := reflect.TypeOf(Wallet{})
	if typ.NumField() != 2 {
		t.Fatalf("Wallet must store exactly ownerAddress and required receivingKeys; got %d fields", typ.NumField())
	}
	if typ.Field(0).Name != "ownerAddress" || typ.Field(1).Name != "receivingKeys" || typ.Field(1).Type.Kind() != reflect.Pointer {
		t.Fatal("Wallet material layout mismatch")
	}
}

func TestWalletRejectsOwnerOnlyMaterial(t *testing.T) {
	owner := common.HexToAddress("0x1111111111111111111111111111111111111111")
	if _, err := WalletFromAddress(owner.Hex()); err == nil {
		t.Fatal("accepted owner-only Wallet")
	}
	if _, err := newPublicWallet(owner, nil); err == nil {
		t.Fatal("accepted missing receiving keys")
	}
	if got, err := (&Wallet{}).GetAddress(); err == nil || got != "" {
		t.Fatal("invalid internal wallet exposed address")
	}
}

const materialAddressVector = "0irFjZ7r3u8t65TeAvH5uTqUJ3KWvsJ983QVcuv83jfKTUMy5Q36R88cst20rxt85TyTnS2k5WAlHS2t8NjykJCVxrK2zgKZ5S50U5uD0G7SuH8AdxYGj0t64Mkgr8qXc5q7gUFFqF"

func TestWalletMaterialVectorAndCopies(t *testing.T) {
	decoded, err := primitives.DecodeAddress(materialAddressVector)
	if err != nil {
		t.Fatal(err)
	}
	keys := &primitives.ReceivingKeys{ViewingPublicKey: decoded.ViewingPublicKey, ClueKey: decoded.ClueKey}
	owner := common.BigToAddress(primitives.ElementToBigInt(decoded.OwnerAddress))
	w, err := newPublicWallet(owner, keys)
	if err != nil {
		t.Fatal(err)
	}
	clone, err := SnapshotWallet(w)
	if err != nil {
		t.Fatal(err)
	}
	keys.ViewingPublicKey[0] ^= 1
	keys.ClueKey[0] ^= 1
	w.ViewingPublicKey()[0] ^= 1
	w.ClueKey()[0] ^= 1
	snapshot := ReceivingKeySnapshot(w)
	snapshot.ClueKey[0] ^= 1
	for _, v := range []*Wallet{w, clone} {
		got, err := v.GetAddress()
		if err != nil || got != materialAddressVector {
			t.Fatal("material aliases or vector changed", got, err)
		}
	}
	w.receivingKeys.ClueKey[0] ^= 1
	got, err := clone.GetAddress()
	if err != nil || got != materialAddressVector {
		t.Fatal("runtime snapshot aliases source", err)
	}
	if got, err := w.GetAddress(); err == nil && got == materialAddressVector {
		t.Fatal("GetAddress returned cached encoding")
	}
	parsed, err := WalletFromAddress(materialAddressVector)
	if err != nil {
		t.Fatal(err)
	}
	if got, err := parsed.GetAddress(); err != nil || got != materialAddressVector {
		t.Fatal("parse/reencode", err)
	}
}

func TestWalletRejectsPartialMaterial(t *testing.T) {
	decoded, _ := primitives.DecodeAddress(materialAddressVector)
	owner := common.BigToAddress(primitives.ElementToBigInt(decoded.OwnerAddress))
	for _, keys := range []*primitives.ReceivingKeys{{}, {ViewingPublicKey: decoded.ViewingPublicKey}, {ClueKey: decoded.ClueKey}, {ViewingPublicKey: make([]byte, 33), ClueKey: decoded.ClueKey}, {ViewingPublicKey: decoded.ViewingPublicKey, ClueKey: make([]byte, 32)}} {
		if _, err := newPublicWallet(owner, keys); err == nil {
			t.Fatal("accepted partial or invalid receiving material")
		}
	}
	if _, err := WalletFromAddress("0x52908400098527886E0F7030069857D2e4169EE7"); err == nil {
		t.Fatal("accepted invalid EVM checksum")
	}
}
