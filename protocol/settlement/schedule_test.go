package settlement

import (
	"bytes"
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
	"os"
	"strings"
	"testing"
)

func TestScheduledQuoteVector(t *testing.T) {
	key, _ := crypto.HexToECDSA(strings.Repeat("0", 63) + "1")
	hash := crypto.Keccak256Hash(bytes.Repeat([]byte{0x55}, 32)).Hex()
	q := Quote{Token: "0x" + strings.Repeat("22", 20), FeeRecipient: "0x" + strings.Repeat("33", 20), FeeAmount: "123", Deadline: "2000000000", MaxGas: "1000000", RequestHash: "0x" + strings.Repeat("44", 32), Execution: "direct", Schedule: &Schedule{"2000000100", "2000000400", hash}}
	d := QuoteDomain{big.NewInt(31337), common.HexToAddress("0x" + strings.Repeat("11", 20))}
	digest, e := QuoteDigest(d, q)
	if e != nil {
		t.Fatal(e)
	}
	sig, e := SignQuote(d, q, key)
	if e != nil {
		t.Fatal(e)
	}
	if e = VerifyQuote(d, q, sig, crypto.PubkeyToAddress(key.PublicKey)); e != nil {
		t.Fatal(e)
	}
	raw, _ := json.MarshalIndent(map[string]any{"quote": q, "digest": digest.Hex(), "signature": "0x" + common.Bytes2Hex(sig), "tokenHash": hash, "chainId": "31337", "protocol": d.Protocol.Hex()}, "", "  ")
	if e = os.WriteFile("/tmp/schedule-go-vector.json", raw, 0600); e != nil {
		t.Fatal(e)
	}
	ordinary := q
	ordinary.Schedule = nil
	h, _ := QuoteDigest(d, ordinary)
	if h == digest {
		t.Fatal("scheduled domain type not separated")
	}
}
func TestScheduleCanonicalBounds(t *testing.T) {
	hash := "0x" + strings.Repeat("ab", 32)
	for _, v := range []Schedule{{"01", "2", hash}, {"0", "2", hash}, {"2", "2", hash}, {"9007199254740992", "9007199254740993", hash}, {"1", "2", "0x" + strings.Repeat("00", 32)}} {
		if v.Validate() == nil {
			t.Fatalf("accepted %+v", v)
		}
	}
	token, hash, e := NewCancelToken()
	if e != nil || len(token) != 66 || len(hash) != 66 {
		t.Fatal(token, hash, e)
	}
}
