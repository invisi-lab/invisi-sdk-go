package core

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

type ReceivingKeys struct{ ViewingPublicKey, ClueKey []byte }
type ReceivingAddress struct {
	WalletIdentity            fr.Element
	ViewingPublicKey, ClueKey []byte
}
type DecodedAddress struct {
	Kind string
	ReceivingAddress
}
type ReceivingKeyLookup func(fr.Element) (*ReceivingKeys, error)

const walletAddressAlphabet = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"

func walletAddressKeys(keys ReceivingKeys) (ReceivingKeys, error) {
	if len(keys.ViewingPublicKey) != 33 || (keys.ViewingPublicKey[0] != 2 && keys.ViewingPublicKey[0] != 3) {
		return ReceivingKeys{}, fmt.Errorf("invalid viewing public key")
	}
	if _, err := ethcrypto.DecompressPubkey(keys.ViewingPublicKey); err != nil {
		return ReceivingKeys{}, err
	}
	if err := ValidateClueKey(keys.ClueKey); err != nil {
		return ReceivingKeys{}, err
	}
	return ReceivingKeys{append([]byte(nil), keys.ViewingPublicKey...), append([]byte(nil), keys.ClueKey...)}, nil
}
func walletAddressChecksum(prefix, data string) string {
	digest := ethcrypto.Keccak256([]byte("Invisi.AddressCaseChecksum.v1:" + prefix + data))
	out := []byte(data)
	for i, c := range out {
		if digest[i>>3]>>uint(7-(i&7))&1 != 0 && c >= 'a' && c <= 'z' {
			out[i] = c - 'a' + 'A'
		}
	}
	return prefix + string(out)
}
func EncodeLongAddress(recipient ReceivingAddress) (string, error) {
	if recipient.WalletIdentity.IsZero() || ElementToBigInt(recipient.WalletIdentity).BitLen() > 160 {
		return "", fmt.Errorf("invalid wallet identity")
	}
	keys, err := walletAddressKeys(ReceivingKeys{recipient.ViewingPublicKey, recipient.ClueKey})
	if err != nil {
		return "", err
	}
	identity := recipient.WalletIdentity.Bytes()
	payload := append(append(append([]byte{}, identity[12:]...), keys.ViewingPublicKey...), keys.ClueKey...)
	var buffer uint32
	bits := uint(0)
	var data strings.Builder
	for _, b := range payload {
		buffer = buffer<<8 | uint32(b)
		bits += 8
		for bits >= 5 {
			bits -= 5
			data.WriteByte(walletAddressAlphabet[buffer>>bits&31])
		}
		buffer &= (1 << bits) - 1
	}
	if bits != 0 {
		data.WriteByte(walletAddressAlphabet[buffer<<(5-bits)&31])
	}
	return walletAddressChecksum("0i", data.String()), nil
}
func EncodeShortAddress(identity fr.Element) (string, error) {
	if identity.IsZero() || ElementToBigInt(identity).BitLen() > 160 {
		return "", fmt.Errorf("invalid wallet identity")
	}
	return common.BigToAddress(ElementToBigInt(identity)).Hex(), nil
}
func DecodeAddress(address string) (DecodedAddress, error) {
	var result DecodedAddress
	if strings.HasPrefix(address, "0x") {
		if len(address) != 42 || !common.IsHexAddress(address) {
			return result, fmt.Errorf("invalid EVM address")
		}
		raw := address[2:]
		if raw != strings.ToLower(raw) && raw != strings.ToUpper(raw) && common.HexToAddress(address).Hex() != address {
			return result, fmt.Errorf("invalid EVM address checksum")
		}
		result.Kind = "short"
		result.WalletIdentity = BigIntToElement(common.HexToAddress(address).Big())
		if result.WalletIdentity.IsZero() {
			return DecodedAddress{}, fmt.Errorf("zero owner")
		}
		return result, nil
	}
	if len(address) != 138 || !strings.HasPrefix(address, "0i") {
		return result, fmt.Errorf("invalid Invisi address")
	}
	data := strings.ToLower(address[2:])
	if walletAddressChecksum("0i", data) != address {
		return result, fmt.Errorf("invalid address checksum")
	}
	var buffer uint32
	bits := uint(0)
	payload := []byte{}
	for _, r := range data {
		v := strings.IndexRune(walletAddressAlphabet, r)
		if v < 0 {
			return result, fmt.Errorf("invalid address alphabet")
		}
		buffer = buffer<<5 | uint32(v)
		bits += 5
		if bits >= 8 {
			bits -= 8
			payload = append(payload, byte(buffer>>bits))
		}
		buffer &= (1 << bits) - 1
	}
	if bits != 0 || buffer != 0 || len(payload) != 85 {
		return result, fmt.Errorf("noncanonical address payload")
	}
	keys, err := walletAddressKeys(ReceivingKeys{payload[20:53], payload[53:]})
	if err != nil {
		return result, err
	}
	result.Kind = "long"
	result.ReceivingAddress = ReceivingAddress{BigIntToElement(common.BytesToAddress(payload[:20]).Big()), keys.ViewingPublicKey, keys.ClueKey}
	if result.WalletIdentity.IsZero() {
		return DecodedAddress{}, fmt.Errorf("zero owner")
	}
	return result, nil
}
func ResolveAddress(address string, lookup ReceivingKeyLookup) (ReceivingAddress, error) {
	decoded, err := DecodeAddress(address)
	if err != nil {
		return ReceivingAddress{}, err
	}
	if decoded.Kind == "long" {
		return decoded.ReceivingAddress, nil
	}
	if lookup == nil {
		return ReceivingAddress{}, fmt.Errorf("registry lookup required")
	}
	keys, err := lookup(decoded.WalletIdentity)
	if err != nil {
		return ReceivingAddress{}, err
	}
	if keys == nil {
		return ReceivingAddress{}, fmt.Errorf("owner is not registered")
	}
	validated, err := walletAddressKeys(*keys)
	if err != nil {
		return ReceivingAddress{}, err
	}
	return ReceivingAddress{decoded.WalletIdentity, validated.ViewingPublicKey, validated.ClueKey}, nil
}
func DecodePrivateKey(text string) ([]byte, error) {
	if !strings.HasPrefix(text, "0x") || len(text) != 66 {
		return nil, fmt.Errorf("expected a 32-byte hexadecimal private key")
	}
	return hex.DecodeString(text[2:])
}
