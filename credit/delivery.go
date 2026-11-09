package credit

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hkdf"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/invisi-lab/invisi-sdk-go/core"
)

const deliveryDomain = "Invisi.CreditDelivery.v2"

type Role byte

const (
	Owner   Role = 1
	Spender Role = 2
)

type DeliveryState struct {
	State                                    State
	OwnerReceivingKeys, SpenderReceivingKeys core.ReceivingKeys
}
type DeliveryContext struct {
	core.ProtocolContext
	TokenAddress, Commitment, RecipientIdentity fr.Element
}

func put(out []byte, e fr.Element, n int) []byte { b := e.Bytes(); return append(out, b[32-n:]...) }
func read(raw []byte) (fr.Element, error) {
	var out fr.Element
	var b [32]byte
	if len(raw) > 32 {
		return out, fmt.Errorf("oversized field")
	}
	copy(b[32-len(raw):], raw)
	return out, out.SetBytesCanonical(b[:])
}
func roleKey(d DeliveryState, role Role) (core.ReceivingKeys, error) {
	switch role {
	case Owner:
		return d.OwnerReceivingKeys, nil
	case Spender:
		return d.SpenderReceivingKeys, nil
	default:
		return core.ReceivingKeys{}, fmt.Errorf("invalid Credit delivery role")
	}
}
func aad(c DeliveryContext, header []byte, role Role) []byte {
	out := append([]byte(deliveryDomain), header...)
	out = append(out, byte(role))
	out = put(out, c.ChainID, 8)
	out = put(out, c.ProtocolAddress, 20)
	out = put(out, c.TokenAddress, 20)
	return put(out, c.Commitment, 32)
}
func encryption(private, peer, recipient, header []byte) (cipher.AEAD, error) {
	k, err := ethcrypto.ToECDSA(private)
	if err != nil {
		return nil, err
	}
	pub, err := ethcrypto.DecompressPubkey(peer)
	if err != nil {
		return nil, err
	}
	x, _ := ethcrypto.S256().ScalarMult(pub.X, pub.Y, k.D.Bytes())
	if x == nil {
		return nil, fmt.Errorf("invalid ECDH shared point")
	}
	secret := x.FillBytes(make([]byte, 32))
	defer clear(secret)
	info := append([]byte(deliveryDomain), header[1:34]...)
	info = append(info, recipient...)
	key, err := hkdf.Key(sha256.New, secret, header[34:50], string(info), 32)
	if err != nil {
		return nil, err
	}
	defer clear(key)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}
func validateDelivery(d DeliveryState, c DeliveryContext, role Role) error {
	if err := d.State.Validate(); err != nil {
		return err
	}
	if d.State.ProtocolContext != c.ProtocolContext || d.State.TokenAddress != c.TokenAddress {
		return fmt.Errorf("Credit delivery context mismatch")
	}
	commitment, err := d.State.Commitment()
	if err != nil {
		return err
	}
	if commitment != c.Commitment {
		return fmt.Errorf("Credit commitment mismatch")
	}
	_, err = roleKey(d, role)
	if err != nil {
		return err
	}
	for _, keys := range []core.ReceivingKeys{d.OwnerReceivingKeys, d.SpenderReceivingKeys} {
		if _, err := ethcrypto.DecompressPubkey(keys.ViewingPublicKey); err != nil {
			return fmt.Errorf("invalid receiving public key")
		}
		if core.ValidateClueKey(keys.ClueKey) != nil {
			return fmt.Errorf("invalid clue key length")
		}
	}
	identity := d.State.OwnerWalletIdentity
	if role == Spender {
		identity = d.State.SpenderWalletIdentity
	}
	if identity != c.RecipientIdentity {
		return fmt.Errorf("recipient identity mismatch")
	}
	return nil
}
func encodeDelivery(d DeliveryState) []byte {
	s := d.State
	out := put(nil, s.SpenderWalletIdentity, 32)
	out = put(out, s.OwnerWalletIdentity, 32)
	out = put(out, s.RevocationShard, 8)
	out = put(out, s.RevocationSlot, 2)
	out = put(out, s.Nonce, 32)
	out = put(out, s.Expiry, 8)
	out = put(out, s.SharedKey, 32)
	for _, v := range s.Policy.Fields() {
		out = put(out, v, 16)
	}
	out = put(out, s.AvailableCredit, 16)
	if s.Unlimited {
		out = append(out, 1)
	} else {
		out = append(out, 0)
	}
	for i, v := range s.Usage.Fields() {
		n := 16
		if i == 0 || i == 2 {
			n = 8
		}
		out = put(out, v, n)
	}
	for _, keys := range []core.ReceivingKeys{d.OwnerReceivingKeys, d.SpenderReceivingKeys} {
		out = append(out, keys.ViewingPublicKey...)
		out = append(out, keys.ClueKey...)
	}
	out = binary.BigEndian.AppendUint16(out, uint16(len(s.Payload)))
	return append(out, s.Payload...)
}
func Encrypt(d DeliveryState, c DeliveryContext, role Role) ([]byte, error) {
	if err := validateDelivery(d, c, role); err != nil {
		return nil, err
	}
	keys, _ := roleKey(d, role)
	if len(d.State.Payload) > 4096-485 {
		return nil, fmt.Errorf("Credit envelope exceeds resource bound")
	}
	private, err := ethcrypto.GenerateKey()
	if err != nil {
		return nil, err
	}
	raw := ethcrypto.FromECDSA(private)
	defer clear(raw)
	header := append([]byte{0x82}, ethcrypto.CompressPubkey(&private.PublicKey)...)
	entropy := make([]byte, 28)
	if _, err = rand.Read(entropy); err != nil {
		return nil, err
	}
	header = append(header, entropy...)
	cipher, err := encryption(raw, keys.ViewingPublicKey, keys.ViewingPublicKey, header)
	if err != nil {
		return nil, err
	}
	plain := encodeDelivery(d)
	defer clear(plain)
	return cipher.Seal(header, header[50:62], plain, aad(c, header, role)), nil
}
func Decrypt(private, envelope []byte, c DeliveryContext, role Role) (DeliveryState, error) {
	if role != Owner && role != Spender {
		return DeliveryState{}, fmt.Errorf("invalid Credit delivery role")
	}
	if len(envelope) < 485 || len(envelope) > 4096 || envelope[0] != 0x82 {
		return DeliveryState{}, fmt.Errorf("unsupported Credit envelope")
	}
	recipient, err := core.ViewingPublicKey(private)
	if err != nil {
		return DeliveryState{}, err
	}
	cipher, err := encryption(private, envelope[1:34], recipient, envelope[:62])
	if err != nil {
		return DeliveryState{}, err
	}
	plain, err := cipher.Open(nil, envelope[50:62], envelope[62:], aad(c, envelope[:62], role))
	if err != nil {
		return DeliveryState{}, fmt.Errorf("Credit authentication failed")
	}
	defer clear(plain)
	return DecodeDeliveryPlaintext(plain, c, role, recipient)
}

// DecodeDeliveryPlaintext verifies native wallet output independently of the executor.
func DecodeDeliveryPlaintext(plain []byte, c DeliveryContext, role Role, recipient []byte) (DeliveryState, error) {
	var err error
	if len(plain) < 407 || int(binary.BigEndian.Uint16(plain[405:407])) != len(plain)-407 {
		return DeliveryState{}, fmt.Errorf("invalid Credit plaintext length")
	}
	s := State{ProtocolContext: c.ProtocolContext, TokenAddress: c.TokenAddress, Payload: append([]byte(nil), plain[407:]...)}
	fields := []*fr.Element{&s.SpenderWalletIdentity, &s.OwnerWalletIdentity, &s.RevocationShard, &s.RevocationSlot, &s.Nonce, &s.Expiry, &s.SharedKey, &s.SingleLimit, &s.DailyLimit, &s.MonthlyLimit, &s.AvailableCredit, &s.LastDay, &s.DailySpent, &s.LastMonth, &s.MonthlySpent, &s.TotalSpent}
	sizes := []int{32, 32, 8, 2, 32, 8, 32, 16, 16, 16, 16, 8, 16, 8, 16, 16}
	at := 0
	for i, target := range fields {
		if i == 11 {
			if plain[at] > 1 {
				return DeliveryState{}, fmt.Errorf("invalid unlimited flag")
			}
			s.Unlimited = plain[at] == 1
			at++
		}
		*target, err = read(plain[at : at+sizes[i]])
		if err != nil {
			return DeliveryState{}, err
		}
		at += sizes[i]
	}
	keys := func() core.ReceivingKeys {
		v := core.ReceivingKeys{ViewingPublicKey: append([]byte(nil), plain[at:at+33]...), ClueKey: append([]byte(nil), plain[at+33:at+65]...)}
		at += 65
		return v
	}
	d := DeliveryState{State: s, OwnerReceivingKeys: keys(), SpenderReceivingKeys: keys()}
	if err = validateDelivery(d, c, role); err != nil {
		return DeliveryState{}, err
	}
	expected, _ := roleKey(d, role)
	if !bytes.Equal(expected.ViewingPublicKey, recipient) {
		return DeliveryState{}, fmt.Errorf("receiving key snapshot mismatch")
	}
	return d, nil
}

func (d DeliveryState) Validate() error {
	commitment, err := d.State.Commitment()
	if err != nil {
		return err
	}
	return validateDelivery(d, DeliveryContext{ProtocolContext: d.State.ProtocolContext, TokenAddress: d.State.TokenAddress, Commitment: commitment, RecipientIdentity: d.State.OwnerWalletIdentity}, Owner)
}

// EncodeDeliveryPlaintext returns the canonical public transport plaintext.
func EncodeDeliveryPlaintext(d DeliveryState) []byte { return encodeDelivery(d) }
