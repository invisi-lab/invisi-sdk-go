package core

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

func currentField(t *testing.T, s string) fr.Element {
	t.Helper()
	var n fr.Element
	if strings.HasPrefix(s, "0x") {
		v, e := HexToElement(s)
		if e != nil {
			t.Fatal(e)
		}
		return v
	}
	if _, e := n.SetString(s); e != nil {
		t.Fatal(e)
	}
	return n
}
func currentBytes(t *testing.T, s string) []byte {
	t.Helper()
	v, e := hex.DecodeString(strings.TrimPrefix(s, "0x"))
	if e != nil {
		t.Fatal(e)
	}
	return v
}
func currentJSON(t *testing.T, name string, v any) {
	t.Helper()
	raw, e := os.ReadFile("testdata/" + name)
	if e != nil {
		t.Fatal(e)
	}
	if e = json.Unmarshal(raw, v); e != nil {
		t.Fatal(e)
	}
}
func TestCurrentKDFAndAddress(t *testing.T) {
	var vectors struct {
		Vectors []struct{ PrivateKey, SpendingPrivateKey, ViewingPrivateKey, DetectionKey string }
	}
	currentJSON(t, "wallet-kdf-vectors.json", &vectors)
	for _, v := range vectors.Vectors {
		w, e := DeriveWallet(currentBytes(t, v.PrivateKey))
		if e != nil {
			t.Fatal(e)
		}
		if hex.EncodeToString(w.SpendingPrivateKey) != v.SpendingPrivateKey || hex.EncodeToString(w.ViewingPrivateKey) != v.ViewingPrivateKey || hex.EncodeToString(w.DetectionKey) != v.DetectionKey {
			t.Fatal("KDF differs from independent TS fixture")
		}
		a, e := EncodeLongAddress(ReceivingAddress{w.WalletIdentity, w.ViewingPublicKey, w.ClueKey})
		if e != nil {
			t.Fatal(e)
		}
		if len(a) != 138 {
			t.Fatal("address width")
		}
		decoded, e := DecodeAddress(a)
		if e != nil || decoded.WalletIdentity != w.WalletIdentity {
			t.Fatal("address round trip", e)
		}
		short, e := EncodeShortAddress(w.WalletIdentity)
		if e != nil || len(short) != 42 {
			t.Fatal("EVM short address", e)
		}
		if _, e := DecodeAddress(a[:137] + "!"); e == nil {
			t.Fatal("bad address accepted")
		}
	}
}
func TestCurrentNoteAndCreditDeliveryVectors(t *testing.T) {
	var v struct {
		Context map[string]string
		Owner   map[string]string
		Note    map[string]string
		Credit  struct {
			SpenderWalletIdentity, OwnerWalletIdentity, SpenderPayout, OwnerPayout, Amount, Nonce, Expiry, CreditSharedKey, Commitment, Nullifier, OwnerCiphertext string
			Remainder                                                                                                                                              struct{ Nullifier, RemainderCommitment, RecoveryAmount string }
		}
	}
	currentJSON(t, "wallet-note-vectors.json", &v)
	ctx := ProtocolContext{currentField(t, v.Context["chainId"]), currentField(t, v.Context["protocolAddress"])}
	token := currentField(t, v.Context["tokenAddress"])
	note := Note{ctx, currentField(t, v.Note["walletIdentity"]), token, currentField(t, v.Note["amount"]), currentField(t, v.Note["nonce"]), nil}
	identity, e := ComputeNoteIdentity(note.WalletIdentity, note.Nonce)
	if e != nil || identity != currentField(t, v.Note["identity"]) {
		t.Fatal("identity mismatch", e)
	}
	leaf, e := ComputeNoteCommitment(note.ChainID, identity, token, note.Amount)
	if e != nil || leaf != currentField(t, v.Note["commitment"]) {
		t.Fatal("note commitment mismatch", e)
	}
	nf, e := ComputeNoteNullifier(note.ChainID, note.Nonce, leaf)
	if e != nil || nf != currentField(t, v.Note["nullifier"]) {
		t.Fatal("nullifier mismatch", e)
	}
	enc := NoteDecryptionContext{NoteEncryptionContext{ctx, leaf}, token, note.WalletIdentity}
	plain, e := DecryptPrivateTransferNote(currentBytes(t, v.Owner["viewingPrivateKey"]), currentBytes(t, v.Note["ciphertext"]), enc)
	if e != nil || !reflect.DeepEqual(plain, note) {
		t.Fatal("TS note decryption failed", e)
	}
	ciphertext := currentBytes(t, v.Note["ciphertext"])
	ciphertext[len(ciphertext)-1] ^= 1
	if _, e := DecryptPrivateTransferNote(currentBytes(t, v.Owner["viewingPrivateKey"]), ciphertext, enc); e == nil {
		t.Fatal("modified envelope accepted")
	}
	c := v.Credit
	credit := CreditNote{ctx, currentField(t, c.SpenderWalletIdentity), currentField(t, c.OwnerWalletIdentity), currentField(t, c.SpenderPayout), currentField(t, c.OwnerPayout), token, currentField(t, c.Amount), currentField(t, c.Nonce), currentField(t, c.Expiry), currentField(t, c.CreditSharedKey), nil}
	keyHash, e := ComputeCreditSharedKeyHash(credit.CreditSharedKey)
	if e != nil {
		t.Fatal(e)
	}
	cl, e := ComputeCreditCommitment(ctx.ChainID, credit.SpenderWalletIdentity, credit.OwnerWalletIdentity, credit.SpenderPayout, credit.OwnerPayout, token, credit.Amount, credit.Nonce, credit.Expiry, keyHash)
	if e != nil || cl != currentField(t, c.Commitment) {
		t.Fatal("Credit commitment mismatch", e)
	}
	cn, e := ComputeCreditNullifier(ctx.ChainID, credit.CreditSharedKey, cl)
	if e != nil || cn != currentField(t, c.Nullifier) {
		t.Fatal("Credit nullifier mismatch", e)
	}
	state, e := DecryptCreditNote(currentBytes(t, v.Owner["viewingPrivateKey"]), currentBytes(t, c.OwnerCiphertext), CreditNoteDecryptionContext{NoteDecryptionContext{NoteEncryptionContext{ctx, cl}, token, credit.OwnerWalletIdentity}, CreditOwner})
	if e != nil || !reflect.DeepEqual(state.Note, credit) {
		t.Fatal("TS Credit decryption failed", e)
	}
	remainder, e := RecoverCreditRemainder(ctx, credit, CreditRemainderRecovery{currentField(t, c.Remainder.Nullifier), currentField(t, c.Remainder.RemainderCommitment), currentField(t, c.Remainder.RecoveryAmount)})
	if e != nil || remainder == nil || remainder.Amount != ElementFromUint64(70) {
		t.Fatal("Credit remainder mismatch", e)
	}
}
func TestCurrentAuthorizationVectors(t *testing.T) {
	var fixture struct {
		Rows []struct {
			Context   map[string]string
			Operation string
			Capacity  int
			Words     []string
			Digest    string
		}
	}
	currentJSON(t, "wallet-authorization-vectors.json", &fixture)
	for _, row := range fixture.Rows {
		words := make([]fr.Element, len(row.Words))
		for i, s := range row.Words {
			words[i] = currentField(t, s)
		}
		r := WalletAuthorizationRequest{ProtocolContext{currentField(t, row.Context["chainId"]), currentField(t, row.Context["protocolAddress"])}, WalletOperation(row.Operation), row.Capacity, words}
		digest, e := ComputeWalletAuthorizationDigest(r)
		if e != nil || digest.Hex() != row.Digest {
			t.Fatalf("%s authorization mismatch: %v", row.Operation, e)
		}
	}
}

func TestCurrentMemoDeliveryVectors(t *testing.T) {
	var vectors struct {
		Context        map[string]string
		Owner, Spender map[string]string
		Note           map[string]string
		Credit         map[string]json.RawMessage
		MemoVectors    []struct{ Name, Kind, Role, Memo, Ciphertext string }
	}
	currentJSON(t, "wallet-note-vectors.json", &vectors)
	if len(vectors.MemoVectors) != 12 {
		t.Fatalf("expected 12 cross-language memo vectors, got %d", len(vectors.MemoVectors))
	}
	context := ProtocolContext{currentField(t, vectors.Context["chainId"]), currentField(t, vectors.Context["protocolAddress"])}
	token := currentField(t, vectors.Context["tokenAddress"])
	creditField := func(name string) fr.Element {
		var value string
		if err := json.Unmarshal(vectors.Credit[name], &value); err != nil {
			t.Fatal(err)
		}
		return currentField(t, value)
	}
	for _, row := range vectors.MemoVectors {
		t.Run(row.Kind+"/"+row.Role+"/"+row.Name, func(t *testing.T) {
			recipient := vectors.Owner
			if row.Role == "spender" {
				recipient = vectors.Spender
			}
			private := currentBytes(t, recipient["viewingPrivateKey"])
			memo := currentBytes(t, row.Memo)
			ciphertext := currentBytes(t, row.Ciphertext)
			if row.Kind == "funds" {
				dec := NoteDecryptionContext{NoteEncryptionContext{context, currentField(t, vectors.Note["commitment"])}, token, currentField(t, recipient["walletIdentity"])}
				note, err := DecryptPrivateTransferNote(private, ciphertext, dec)
				if err != nil {
					t.Fatal(err)
				}
				expected := Note{context, currentField(t, vectors.Note["walletIdentity"]), token, currentField(t, vectors.Note["amount"]), currentField(t, vectors.Note["nonce"]), memo}
				if !reflect.DeepEqual(note, expected) {
					t.Fatal("memo vector changed note preimage or memo")
				}
			} else {
				dec := CreditNoteDecryptionContext{NoteDecryptionContext{NoteEncryptionContext{context, creditField("commitment")}, token, currentField(t, recipient["walletIdentity"])}, CreditDeliveryRole(row.Role)}
				state, err := DecryptCreditNote(private, ciphertext, dec)
				if err != nil {
					t.Fatal(err)
				}
				expected := CreditNote{context, creditField("spenderWalletIdentity"), creditField("ownerWalletIdentity"), creditField("spenderPayout"), creditField("ownerPayout"), token, creditField("amount"), creditField("nonce"), creditField("expiry"), creditField("creditSharedKey"), memo}
				if !reflect.DeepEqual(state.Note, expected) {
					t.Fatal("memo vector changed Credit preimage or memo")
				}
			}
		})
	}
}
