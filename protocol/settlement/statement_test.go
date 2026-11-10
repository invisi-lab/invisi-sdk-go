package settlement

import (
	"encoding/json"
	"math/big"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

func TestAuthenticatedStatementVectors(t *testing.T) {
	raw, e := os.ReadFile("testdata/statements.json")
	if e != nil {
		t.Fatal(e)
	}
	var f struct {
		ChainID  string         `json:"chainId"`
		Protocol common.Address `json:"protocol"`
		Vectors  []struct {
			Data       hexutil.Bytes `json:"data"`
			Operations []struct {
				Capacity uint32        `json:"mainCircuitCapacity"`
				Inputs   [][32]byte    `json:"-"`
				Words    []common.Hash `json:"publicInputs"`
				Digest   common.Hash   `json:"operationDigest"`
			} `json:"operations"`
		} `json:"vectors"`
	}
	if e = json.Unmarshal(raw, &f); e != nil {
		t.Fatal(e)
	}
	chain, _ := new(big.Int).SetString(f.ChainID, 10)
	for i, v := range f.Vectors {
		calls, e := DecodeCalls(v.Data)
		if e != nil {
			t.Fatal(e)
		}
		if len(calls) != len(v.Operations) {
			t.Fatal("member count")
		}
		for j, r := range calls {
			c, w, d, e := TransactStatement(chain, f.Protocol, &r)
			if e != nil {
				t.Fatalf("%d/%d: %v", i, j, e)
			}
			want := v.Operations[j]
			if c != want.Capacity || d != want.Digest || len(w) != len(want.Words) {
				t.Fatalf("statement %d/%d mismatch", i, j)
			}
			for k := range w {
				if common.Hash(w[k]) != want.Words[k] {
					t.Fatalf("word %d mismatch", k)
				}
			}
		}
		if _, e = DecodeCalls(append(append([]byte{}, v.Data...), 0)); e == nil {
			t.Fatal("accepted trailing calldata")
		}
	}
}
