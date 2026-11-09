package credit

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"math/big"
	"testing"
)

func TestGrantAuditIsScopedAndPreservesUint128Usage(t *testing.T) {
	s, _ := fixture(t)
	wide := core.BigIntToElement(new(big.Int).Add(new(big.Int).Lsh(big.NewInt(1), 127), big.NewInt(3)))
	s.LastDay = core.ElementFromUint64(1)
	s.LastMonth = core.ElementFromUint64(1)
	s.DailySpent = wide
	s.MonthlySpent = wide
	s.TotalSpent = wide
	commitment, err := s.Commitment()
	if err != nil {
		t.Fatal(err)
	}
	disclosure, err := Disclose(s)
	if err != nil {
		t.Fatal(err)
	}
	verified, err := disclosure.Verify(common.Hash(commitment.Bytes()))
	if err != nil || verified.TotalSpent != wide {
		t.Fatal("wide grant audit rejected", err)
	}
	exported := disclosure.Export()
	if _, ok := exported.Fields["creditSharedKey"]; ok {
		t.Fatal("private grant key mixed into opening fields")
	}
	exported.Fields["totalSpent"] = new(big.Int).Lsh(big.NewInt(1), 128).String()
	raw, _ := json.Marshal(exported)
	bad, err := DisclosureFromExport(raw)
	if err == nil {
		_, err = bad.Verify(common.Hash(commitment.Bytes()))
	}
	if err == nil {
		t.Fatal("accepted uint128 overflow audit")
	}
	if _, err = disclosure.Verify(common.HexToHash("0x01")); err == nil {
		t.Fatal("audit accepted wrong canonical anchor")
	}
}
