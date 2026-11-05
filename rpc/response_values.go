package rpc

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

func wireObject(raw json.RawMessage) (map[string]json.RawMessage, error) {
	var m map[string]json.RawMessage
	e := json.Unmarshal(raw, &m)
	if e != nil || m == nil {
		return nil, errors.New("expected RPC object")
	}
	return m, nil
}

func wireUint(raw json.RawMessage, bits int) (uint64, error) {
	s := string(raw)
	quoted := strings.HasPrefix(s, "\"")
	if quoted {
		if json.Unmarshal(raw, &s) != nil {
			return 0, errors.New("invalid RPC integer")
		}
	}
	if s == "" || (len(s) > 1 && s[0] == '0') {
		return 0, errors.New("invalid RPC integer")
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, errors.New("invalid RPC integer")
		}
	}
	n, e := strconv.ParseUint(s, 10, bits)
	if e == nil && !quoted && n > 1<<53-1 {
		e = errors.New("unsafe RPC integer")
	}
	return n, e
}

func wireBytes(raw json.RawMessage, min, max int) ([]byte, error) {
	var s string
	if json.Unmarshal(raw, &s) != nil || !strings.HasPrefix(s, "0x") {
		return nil, errors.New("invalid RPC hex")
	}
	b, e := hex.DecodeString(s[2:])
	if e != nil || len(b) < min || len(b) > max {
		return nil, errors.New("invalid RPC bytes")
	}
	return b, nil
}

func wireHash(raw json.RawMessage) (common.Hash, error) {
	b, e := wireBytes(raw, 32, 32)
	return common.BytesToHash(b), e
}

func wireBlock(raw json.RawMessage) (CanonicalBlock, error) {
	m, e := wireObject(raw)
	if e != nil {
		return CanonicalBlock{}, e
	}
	var b CanonicalBlock
	if b.Number, e = wireUint(m["number"], 64); e != nil {
		return b, e
	}
	if b.Timestamp, e = wireUint(m["timestamp"], 64); e != nil {
		return b, e
	}
	if b.Hash, e = wireHash(m["hash"]); e != nil {
		return b, e
	}
	b.ParentHash, e = wireHash(m["parentHash"])
	return b, e
}
