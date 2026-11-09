package credit

import (
	hexcodec "encoding/hex"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"strings"
	"testing"
)

func bytesHex(t *testing.T, s string) []byte {
	t.Helper()
	b, e := hexcodec.DecodeString(strings.TrimPrefix(s, "0x"))
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func TestCanonicalDeferredDelivery(t *testing.T) {
	s, f := fixture(t)
	d := f["delivery"].(map[string]any)
	commitment, e := s.Commitment()
	if e != nil {
		t.Fatal(e)
	}
	for _, name := range []string{"owner", "spender"} {
		role := Owner
		if name == "spender" {
			role = Spender
		}
		owner := s.OwnerWalletIdentity
		if role == Spender {
			owner = s.SpenderWalletIdentity
		}
		context := DeliveryContext{ProtocolContext: s.ProtocolContext, TokenAddress: s.TokenAddress, Commitment: commitment, RecipientIdentity: owner}
		cipher := bytesHex(t, d[name+"Ciphertext"].(string))
		private := bytesHex(t, d[name+"PrivateKey"].(string))
		if len(cipher) != 485 || cipher[0] != 0x82 {
			t.Fatal("wrong canonical size")
		}
		got, e := Decrypt(private, cipher, context, role)
		if e != nil {
			t.Fatal(e)
		}
		if c, e := got.State.Commitment(); e != nil || c != commitment {
			t.Fatal("state mismatch", e)
		}
		roundtrip, e := Encrypt(got, context, role)
		if e != nil {
			t.Fatal(e)
		}
		if _, e := Decrypt(private, roundtrip, context, role); e != nil {
			t.Fatal(e)
		}
		cipher[0] = 0xff
		if _, e := Decrypt(private, cipher, context, role); e == nil {
			t.Fatal("accepted unknown envelope format")
		}
		context.RecipientIdentity = core.ElementFromUint64(999)
		if _, e := Decrypt(private, roundtrip, context, role); e == nil {
			t.Fatal("accepted wrong identity")
		}
	}
}

func TestQuotaEnvelopeBoundsAndStrictBoolean(t *testing.T) {
	s, f := fixture(t)
	material := f["delivery"].(map[string]any)
	keys := func(name string) core.ReceivingKeys {
		values := material[name+"ReceivingKeys"].(map[string]any)
		return core.ReceivingKeys{ViewingPublicKey: bytesHex(t, values["viewingPublicKey"].(string)), ClueKey: bytesHex(t, values["clueKey"].(string))}
	}
	d := DeliveryState{State: s, OwnerReceivingKeys: keys("owner"), SpenderReceivingKeys: keys("spender")}
	commitment, err := s.Commitment()
	if err != nil {
		t.Fatal(err)
	}
	c := DeliveryContext{ProtocolContext: s.ProtocolContext, TokenAddress: s.TokenAddress, Commitment: commitment, RecipientIdentity: s.OwnerWalletIdentity}
	private := bytesHex(t, material["ownerPrivateKey"].(string))
	for _, size := range []int{0, 3611, 3612} {
		d.State.Payload = make([]byte, size)
		envelope, err := Encrypt(d, c, Owner)
		if size == 3612 {
			if err == nil {
				t.Fatal("accepted oversized quota envelope")
			}
			continue
		}
		if err != nil || len(envelope) != 485+size {
			t.Fatal("incorrect quota envelope bound", len(envelope), err)
		}
		opened, err := Decrypt(private, envelope, c, Owner)
		if err != nil || opened.State.AvailableCredit != s.AvailableCredit || opened.State.Unlimited {
			t.Fatal("quota wire roundtrip", err)
		}
	}
	d.State.Payload = nil
	ephemeral := make([]byte, 32)
	ephemeral[31] = 99
	key, err := crypto.ToECDSA(ephemeral)
	if err != nil {
		t.Fatal(err)
	}
	header := append([]byte{0x82}, crypto.CompressPubkey(&key.PublicKey)...)
	header = append(header, make([]byte, 28)...)
	aead, err := encryption(ephemeral, d.OwnerReceivingKeys.ViewingPublicKey, d.OwnerReceivingKeys.ViewingPublicKey, header)
	if err != nil {
		t.Fatal(err)
	}
	plain := encodeDelivery(d)
	if len(plain) != 407 {
		t.Fatal("incorrect fixed plaintext size", len(plain))
	}
	plain[210] = 2
	envelope := aead.Seal(header, header[50:62], plain, aad(c, header, Owner))
	if _, err = Decrypt(private, envelope, c, Owner); err == nil {
		t.Fatal("accepted nonboolean quota flag in authenticated envelope")
	}
	d.State.AvailableCredit = core.ElementFromUint64(0)
	d.State.Unlimited = true
	commitment, err = d.State.Commitment()
	if err != nil {
		t.Fatal(err)
	}
	c.Commitment = commitment
	envelope, err = Encrypt(d, c, Owner)
	if err != nil {
		t.Fatal(err)
	}
	opened, err := Decrypt(private, envelope, c, Owner)
	if err != nil || !opened.State.Unlimited || !opened.State.AvailableCredit.IsZero() {
		t.Fatal("unlimited wire roundtrip", err)
	}
}
