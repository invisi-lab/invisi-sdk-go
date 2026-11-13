package settlement

import (
	"crypto/rand"
	"encoding/hex"
	"github.com/ethereum/go-ethereum/crypto"
	"strconv"
	"strings"
)

type Schedule struct {
	PublishAfter    string `json:"publishAfter"`
	ExecuteBefore   string `json:"executeBefore"`
	CancelTokenHash string `json:"cancelTokenHash"`
}

var scheduledQuoteTypeHash = crypto.Keccak256Hash([]byte("ScheduledQuote(address token,address feeRecipient,uint256 feeAmount,uint256 deadline,uint256 maxGas,bytes32 requestHash,string execution,uint256 publishAfter,uint256 executeBefore,bytes32 cancelTokenHash)"))

func ScheduleSeconds(s string) (uint64, error) {
	n, e := strconv.ParseUint(s, 10, 64)
	if e != nil || n == 0 || n > 9007199254740991 || strconv.FormatUint(n, 10) != s {
		return 0, ErrQuote
	}
	return n, nil
}
func (s Schedule) Validate() error {
	a, e := ScheduleSeconds(s.PublishAfter)
	if e != nil {
		return e
	}
	b, e := ScheduleSeconds(s.ExecuteBefore)
	if e != nil || b <= a {
		return ErrQuote
	}
	h, e := hex.DecodeString(strings.TrimPrefix(s.CancelTokenHash, "0x"))
	if e != nil || len(h) != 32 || s.CancelTokenHash != "0x"+hex.EncodeToString(h) {
		return ErrQuote
	}
	for _, v := range h {
		if v != 0 {
			return nil
		}
	}
	return ErrQuote
}
func NewCancelToken() (string, string, error) {
	b := make([]byte, 32)
	if _, e := rand.Read(b); e != nil {
		return "", "", e
	}
	return "0x" + hex.EncodeToString(b), crypto.Keccak256Hash(b).Hex(), nil
}
