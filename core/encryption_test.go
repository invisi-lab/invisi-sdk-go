package core

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"
)

type envelopeFixture struct {
	Kind, PrivateKey, PublicKey, Entropy, Envelope string
	Context, Note                                  map[string]string
}

func envelopeFixtures(t *testing.T) []envelopeFixture {
	t.Helper()
	data, err := os.ReadFile("testdata/envelopes.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct{ Envelopes []envelopeFixture }
	if err = json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Envelopes) != 6 {
		t.Fatal("expected six envelope fixtures")
	}
	return fixture.Envelopes
}
func fixtureBytes(t *testing.T, s string) []byte {
	t.Helper()
	value, err := hex.DecodeString(s[2:])
	if err != nil {
		t.Fatal(err)
	}
	return value
}
func noteFixture(t *testing.T, m map[string]string) Note {
	return Note{NoteSecret: vectorElement(t, m["noteSecret"]), ReceivingAddress: vectorElement(t, m["receivingAddress"]), TokenAddress: vectorElement(t, m["tokenAddress"]), Amount: vectorElement(t, m["amount"]), Nonce: vectorElement(t, m["nonce"])}
}
func TestProtocolEnvelopeInteroperability(t *testing.T) {
	for _, v := range envelopeFixtures(t) {
		t.Run(v.Kind, func(t *testing.T) {
			key := fixtureBytes(t, v.PrivateKey)
			public := fixtureBytes(t, v.PublicKey)
			envelope := fixtureBytes(t, v.Envelope)
			gotPublic, err := ViewingPublicKey(key)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(gotPublic, public) {
				t.Fatal("viewing key differs from protocol")
			}
			context := NoteEncryptionContext{ProtocolContext: ProtocolContext{ChainID: vectorElement(t, v.Context["chainId"]), ProtocolAddress: vectorElement(t, v.Context["protocolAddress"])}, OutputCommitment: vectorElement(t, v.Context["outputCommitment"])}
			decryptContext := NoteDecryptionContext{NoteEncryptionContext: context, TokenAddress: vectorElement(t, v.Context["tokenAddress"])}
			var plaintext []byte
			version := byte(1)
			if v.Kind == "ordinary" {
				want := noteFixture(t, v.Note)
				got, err := DecryptPrivateTransferNote(key, envelope, decryptContext)
				if err != nil {
					t.Fatal(err)
				}
				if got != want {
					t.Fatal("ordinary note differs from protocol")
				}
				plaintext = encodeNote(want)
				randomEnvelope, err := EncryptPrivateTransferNote(public, want, context)
				if err != nil {
					t.Fatal(err)
				}
				if len(randomEnvelope) != 178 {
					t.Fatal("wrong ordinary envelope length")
				}
				if got, err := DecryptPrivateTransferNote(key, randomEnvelope, decryptContext); err != nil || got != want {
					t.Fatalf("ordinary roundtrip: %v", err)
				}
			} else {
				version = 2
				want := creditFixture(t, v.Note)
				got, err := DecryptCreditNote(key, envelope, decryptContext)
				if err != nil {
					t.Fatal(err)
				}
				if got != want {
					t.Fatal("credit note differs from protocol")
				}
				plaintext = encodeCredit(want)
				randomEnvelope, err := EncryptCreditNote(public, want, context)
				if err != nil {
					t.Fatal(err)
				}
				if len(randomEnvelope) != 254 {
					t.Fatal("wrong Credit envelope length")
				}
				if got, err := DecryptCreditNote(key, randomEnvelope, decryptContext); err != nil || got != want {
					t.Fatalf("Credit roundtrip: %v", err)
				}
			}
			got, err := sealEnvelope(public, plaintext, context, decryptContext.TokenAddress, version, bytes.NewReader(fixtureBytes(t, v.Entropy)))
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, envelope) {
				t.Fatalf("encrypted bytes differ from protocol: %x", got)
			}
			decrypt := func(data []byte, ctx NoteDecryptionContext) error {
				if version == 1 {
					_, err := DecryptPrivateTransferNote(key, data, ctx)
					return err
				}
				_, err := DecryptCreditNote(key, data, ctx)
				return err
			}
			for _, offset := range []int{0, 2, 34, 50, 62, len(envelope) - 1} {
				tampered := bytes.Clone(envelope)
				tampered[offset] ^= 1
				if err := decrypt(tampered, decryptContext); err == nil {
					t.Fatalf("accepted tampered byte %d", offset)
				}
			}
			// v2 authenticates the compressed point sign in the header; v1 uses
			// the shared X coordinate and has the protocol's original AAD layout.
			if version == 2 {
				tampered := bytes.Clone(envelope)
				tampered[1] ^= 1
				if err := decrypt(tampered, decryptContext); err == nil {
					t.Fatal("accepted modified v2 point sign")
				}
			}
			wrong := decryptContext
			wrong.ChainID = ElementFromUint64(42)
			if err := decrypt(envelope, wrong); err == nil {
				t.Fatal("accepted wrong chain")
			}
			wrong = decryptContext
			wrong.ProtocolAddress = ElementFromUint64(42)
			if err := decrypt(envelope, wrong); err == nil {
				t.Fatal("accepted wrong deployment")
			}
			wrong = decryptContext
			wrong.TokenAddress = ElementFromUint64(42)
			if err := decrypt(envelope, wrong); err == nil {
				t.Fatal("accepted wrong token")
			}
			wrong = decryptContext
			wrong.OutputCommitment = ElementFromUint64(42)
			if err := decrypt(envelope, wrong); err == nil {
				t.Fatal("accepted wrong commitment")
			}
			if err := decrypt(envelope[:len(envelope)-1], decryptContext); err == nil {
				t.Fatal("accepted truncated envelope")
			}
			// Seal authenticated but noncanonical plaintext. Parsing must reject it without modular reduction.
			malformed := bytes.Clone(plaintext)
			offset := 16
			if version == 2 {
				offset = 80
			}
			modulus := FieldModulus().FillBytes(make([]byte, 32))
			copy(malformed[offset:offset+32], modulus)
			data, err := sealEnvelope(public, malformed, context, decryptContext.TokenAddress, version, bytes.NewReader(fixtureBytes(t, v.Entropy)))
			if err != nil {
				t.Fatal(err)
			}
			if err := decrypt(data, decryptContext); err == nil {
				t.Fatal("accepted noncanonical plaintext field")
			}
		})
	}
}

func TestEnvelopeRejectsInvalidKeysAndRanges(t *testing.T) {
	for _, key := range [][]byte{nil, make([]byte, 31), make([]byte, 32), bytes.Repeat([]byte{255}, 32)} {
		if _, err := ViewingPublicKey(key); err == nil {
			t.Fatal("accepted invalid viewing private key")
		}
	}
	fixtures := envelopeFixtures(t)
	for _, v := range fixtures {
		context := NoteEncryptionContext{ProtocolContext: ProtocolContext{ChainID: vectorElement(t, v.Context["chainId"]), ProtocolAddress: vectorElement(t, v.Context["protocolAddress"])}, OutputCommitment: vectorElement(t, v.Context["outputCommitment"])}
		public := fixtureBytes(t, v.PublicKey)
		if v.Kind == "ordinary" {
			note := noteFixture(t, v.Note)
			for _, badKey := range [][]byte{nil, make([]byte, 33), append([]byte{4}, public[1:]...)} {
				if _, err := EncryptPrivateTransferNote(badKey, note, context); err == nil {
					t.Fatal("accepted invalid viewing public key")
				}
			}
			bad := note
			bad.Amount = ElementFromUint64(0)
			if _, err := EncryptPrivateTransferNote(public, bad, context); err == nil {
				t.Fatal("accepted zero amount")
			}
			bad = note
			bad.Amount = vectorElement(t, "340282366920938463463374607431768211456")
			if _, err := EncryptPrivateTransferNote(public, bad, context); err == nil {
				t.Fatal("accepted uint128 overflow")
			}
			bad = note
			bad.ReceivingAddress = vectorElement(t, "1461501637330902918203684832716283019655932542976")
			if _, err := EncryptPrivateTransferNote(public, bad, context); err == nil {
				t.Fatal("accepted uint160 overflow")
			}
		} else {
			note := creditFixture(t, v.Note)
			bad := note
			bad.Expiry = ElementFromUint64(0)
			if _, err := EncryptCreditNote(public, bad, context); err == nil {
				t.Fatal("accepted zero Credit expiry")
			}
			bad = note
			bad.CreditSharedKey = ElementFromUint64(0)
			if _, err := EncryptCreditNote(public, bad, context); err == nil {
				t.Fatal("accepted zero shared key")
			}
			bad = note
			bad.Amount = vectorElement(t, "85070591730234615865843651857942052864")
			if _, err := EncryptCreditNote(public, bad, context); err == nil {
				t.Fatal("accepted uint126 overflow")
			}
		}
	}
}
