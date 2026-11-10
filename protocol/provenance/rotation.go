package provenance

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"regexp"
	"strings"
)

type RotationRecord struct {
	Version              string `json:"version"`
	ProviderID           string `json:"providerId"`
	OldKeyID             string `json:"oldKeyId"`
	NewKeyID             string `json:"newKeyId"`
	NewPublicKey         string `json:"newPublicKey"`
	ActivatedAt          string `json:"activatedAt"`
	PreviousRotationHash string `json:"previousRotationHash"`
	Signature            string `json:"signature"`
}
type KeyValidity struct {
	ActivatedAt string
	RetiredAt   string
}
type RotationCheckpoint struct {
	ProviderID           string
	Keys                 map[string]ed25519.PublicKey
	Validity             map[string]KeyValidity
	PreviousRotationHash string
	LastActivatedAt      string
	Now                  string
}
type TrustedKeyring struct {
	ProviderID      string
	Keys            map[string]ed25519.PublicKey
	Validity        map[string]KeyValidity
	RotationHash    string
	LastActivatedAt string
}

func decodeFixed(value string, size int) ([]byte, error) {
	if !strings.HasPrefix(value, "0x") || len(value) != 2+size*2 {
		return nil, errors.New("invalid fixed-width hex")
	}
	return hex.DecodeString(value[2:])
}

var decimalTime = regexp.MustCompile(`^(0|[1-9][0-9]*)$`)

func rotationTime(value string) (uint64, error) {
	if !decimalTime.MatchString(value) {
		return 0, errors.New("rotation time must be decimal")
	}
	n, e := Uint(value, 64)
	if e != nil {
		return 0, e
	}
	return n.Uint64(), nil
}
func RotationPayload(r RotationRecord) ([]byte, error) {
	if r.Version != "1" {
		return nil, errors.New("unsupported rotation version")
	}
	out := []byte{1}
	var newID, newKey []byte
	for i, value := range []string{r.ProviderID, r.OldKeyID, r.NewKeyID, r.NewPublicKey} {
		b, e := decodeFixed(value, 32)
		if e != nil {
			return nil, e
		}
		out = append(out, b...)
		if i == 2 {
			newID = b
		}
		if i == 3 {
			newKey = b
		}
	}
	digest := sha256.Sum256(newKey)
	if !bytes.Equal(digest[:], newID) {
		return nil, errors.New("rotation key ID mismatch")
	}
	at, e := rotationTime(r.ActivatedAt)
	if e != nil {
		return nil, e
	}
	out = binary.BigEndian.AppendUint64(out, at)
	previous, e := decodeFixed(r.PreviousRotationHash, 32)
	if e != nil {
		return nil, e
	}
	return append(out, previous...), nil
}
func RotationMessage(r RotationRecord) ([]byte, error) {
	payload, e := RotationPayload(r)
	if e != nil {
		return nil, e
	}
	h := sha256.Sum256(append([]byte("Invisi.Provenance.KeyRotation.v1"), payload...))
	return h[:], nil
}
func RotationHash(r RotationRecord) (string, error) {
	payload, e := RotationPayload(r)
	if e != nil {
		return "", e
	}
	h := sha256.Sum256(payload)
	return "0x" + hex.EncodeToString(h[:]), nil
}
func VerifyRotationChain(initial RotationCheckpoint, records []RotationRecord) (TrustedKeyring, error) {
	provider, e := decodeFixed(initial.ProviderID, 32)
	if e != nil {
		return TrustedKeyring{}, e
	}
	previous, e := decodeFixed(initial.PreviousRotationHash, 32)
	if e != nil {
		return TrustedKeyring{}, e
	}
	now, e := rotationTime(initial.Now)
	if e != nil {
		return TrustedKeyring{}, e
	}
	last, e := rotationTime(initial.LastActivatedAt)
	if e != nil || last > now {
		return TrustedKeyring{}, errors.New("invalid rotation checkpoint time")
	}
	result := TrustedKeyring{ProviderID: "0x" + hex.EncodeToString(provider), Keys: map[string]ed25519.PublicKey{}, Validity: map[string]KeyValidity{}, RotationHash: "0x" + hex.EncodeToString(previous), LastActivatedAt: initial.LastActivatedAt}
	for id, key := range initial.Keys {
		raw, e := decodeFixed(id, 32)
		if e != nil || len(key) != 32 {
			return TrustedKeyring{}, errors.New("invalid trusted key")
		}
		canonical := "0x" + hex.EncodeToString(raw)
		result.Keys[canonical] = append(ed25519.PublicKey(nil), key...)
		interval, ok := initial.Validity[canonical]
		if !ok {
			interval = KeyValidity{ActivatedAt: "0"}
		}
		result.Validity[canonical] = interval
	}
	for _, r := range records {
		message, e := RotationMessage(r)
		if e != nil {
			return TrustedKeyring{}, e
		}
		at, e := rotationTime(r.ActivatedAt)
		if e != nil {
			return TrustedKeyring{}, e
		}
		old := strings.ToLower(r.OldKeyID)
		next := strings.ToLower(r.NewKeyID)
		key := result.Keys[old]
		interval, ok := result.Validity[old]
		activated, e := rotationTime(interval.ActivatedAt)
		if strings.ToLower(r.ProviderID) != result.ProviderID || strings.ToLower(r.PreviousRotationHash) != result.RotationHash || at <= last || at > now || len(key) != 32 || !ok || interval.RetiredAt != "" || e != nil || at < activated || result.Keys[next] != nil {
			return TrustedKeyring{}, errors.New("untrusted or out-of-order key rotation")
		}
		signature, e := decodeFixed(r.Signature, 64)
		if e != nil || !ed25519.Verify(key, message, signature) {
			return TrustedKeyring{}, errors.New("invalid key rotation signature")
		}
		publicKey, _ := decodeFixed(r.NewPublicKey, 32)
		interval.RetiredAt = r.ActivatedAt
		result.Validity[old] = interval
		result.Keys[next] = publicKey
		result.Validity[next] = KeyValidity{ActivatedAt: r.ActivatedAt}
		result.RotationHash, _ = RotationHash(r)
		result.LastActivatedAt = r.ActivatedAt
		last = at
	}
	return result, nil
}
func WithRotationHistory(trust TrustContext, keys TrustedKeyring) (TrustContext, error) {
	if !equalNumber(trust.ProviderID, keys.ProviderID) {
		return TrustContext{}, errors.New("rotation provider mismatch")
	}
	trust.SigningKeys = keys.Keys
	trust.SigningKeyValidity = keys.Validity
	return trust, nil
}
