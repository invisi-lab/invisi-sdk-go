package core

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"unicode/utf8"
)

// MaxMemoBytes limits authenticated output metadata, measured in bytes.
const MaxMemoBytes = 256

// ValidateMemo accepts arbitrary bytes, including invalid UTF-8.
func ValidateMemo(memo []byte) error {
	if len(memo) > MaxMemoBytes {
		return fmt.Errorf("memo exceeds %d bytes", MaxMemoBytes)
	}
	return nil
}

// MemoFromText encodes valid UTF-8 text, enforcing the byte limit.
func MemoFromText(text string) ([]byte, error) {
	if !utf8.ValidString(text) {
		return nil, fmt.Errorf("memo text must be valid UTF-8")
	}
	memo := []byte(text)
	if err := ValidateMemo(memo); err != nil {
		return nil, err
	}
	return memo, nil
}

// MemoText explicitly decodes a memo as UTF-8; binary memos remain valid notes.
func MemoText(memo []byte) (string, error) {
	if err := ValidateMemo(memo); err != nil {
		return "", err
	}
	if !utf8.Valid(memo) {
		return "", fmt.Errorf("memo is not valid UTF-8")
	}
	return string(memo), nil
}

// CloneMemo copies optional metadata, preserving nil versus explicitly empty input.
func CloneMemo(memo []byte) []byte {
	if memo == nil {
		return nil
	}
	return append(make([]byte, 0, len(memo)), memo...)
}

// Clone returns an independent note including its memo.
func (n Note) Clone() Note { n.Memo = CloneMemo(n.Memo); return n }

// Clone returns an independent Credit note including its memo.
func (n CreditNote) Clone() CreditNote { n.Memo = CloneMemo(n.Memo); return n }

// Equal compares the preimage and optional delivery metadata.
func (n Note) Equal(other Note) bool {
	return n.ProtocolContext == other.ProtocolContext && n.WalletIdentity == other.WalletIdentity && n.TokenAddress == other.TokenAddress && n.Amount == other.Amount && n.Nonce == other.Nonce && bytes.Equal(n.Memo, other.Memo)
}
func appendMemo(plain, memo []byte) []byte {
	plain = binary.BigEndian.AppendUint16(plain, uint16(len(memo)))
	return append(plain, memo...)
}
func readMemo(plain []byte, base int) ([]byte, error) {
	if len(plain) < base+2 {
		return nil, fmt.Errorf("missing memo length")
	}
	length := int(binary.BigEndian.Uint16(plain[base : base+2]))
	if length > MaxMemoBytes || len(plain) != base+2+length {
		return nil, fmt.Errorf("invalid memo length")
	}
	if length == 0 {
		return nil, nil
	}
	return CloneMemo(plain[base+2:]), nil
}
