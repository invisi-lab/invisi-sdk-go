package credit

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	"math/big"
	"os"
	"testing"
)

func TestCanonicalRequestStatements(t *testing.T) {
	s, _ := fixture(t)
	raw, e := os.ReadFile("testdata/requests.json")
	if e != nil {
		t.Fatal(e)
	}
	var f map[string]json.RawMessage
	if e = json.Unmarshal(raw, &f); e != nil {
		t.Fatal(e)
	}
	var ds []struct{ Ciphertext, Clue string }
	if e = json.Unmarshal(f["deliveries"], &ds); e != nil {
		t.Fatal(e)
	}
	deliveries := []protocol.Delivery{}
	for _, d := range ds {
		deliveries = append(deliveries, protocol.Delivery{Ciphertext: bytesHex(t, d.Ciphertext), Clue: bytesHex(t, d.Clue)})
	}
	ctx := Context{ProtocolContext: s.ProtocolContext, TokenAddress: common.BigToAddress(big.NewInt(34)), Timestamp: 2000000000}
	c, _ := s.Commitment()
	nf, _ := s.Nullifier()
	next, _ := s.Draw(s.SingleLimit, 2000000000)
	_ = next
	created := CreateParams{Token: ctx.TokenAddress, Owner: common.BigToAddress(big.NewInt(44)), RevocationShard: 1, CreditCommitment: common.HexToHash(hex(c)), Deadline: big.NewInt(2000000300)}
	drawn, _ := s.Draw(core.ElementFromUint64(7), 2000000000)
	updated := UpdateParams{Token: ctx.TokenAddress, Owner: created.Owner, Source: protocol.InputSource{TreeId: big.NewInt(1), MerkleRoot: common.HexToHash("0x01")}, Nullifier: common.HexToHash(hex(nf)), CreditCommitment: common.HexToHash(hex(drawn.StateCommitment)), RevocationShard: 1, Deadline: created.Deadline}
	draw := DrawParams{Token: ctx.TokenAddress, Sources: []protocol.InputSource{updated.Source}, Payers: []Payer{{Payer: created.Owner, Amount: big.NewInt(7)}}, GrantRefs: []GrantRef{{PayerIndex: big.NewInt(0), Shard: 1}}, Nullifiers: []common.Hash{updated.Nullifier}, CreditCommitments: []common.Hash{updated.CreditCommitment}, RecoveryAmounts: []common.Hash{common.HexToHash(hex(drawn.RecoveryAmount))}, NoteCommitment: common.HexToHash("0x02"), TotalAmount: big.NewInt(7), Deadline: created.Deadline}
	cases := []struct {
		name string
		run  func() ([]common.Hash, apitypes.TypedData, error)
	}{
		{"create", func() ([]common.Hash, apitypes.TypedData, error) { return created.Statement(ctx, deliveries) }},
		{"update", func() ([]common.Hash, apitypes.TypedData, error) { return updated.Statement(ctx, deliveries) }},
		{"draw", func() ([]common.Hash, apitypes.TypedData, error) {
			return draw.Statement(ctx, append(deliveries, protocol.Delivery{}))
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var expected struct {
				PublicInputs []common.Hash
				TypedData    apitypes.TypedData
			}
			if e = json.Unmarshal(f[tc.name], &expected); e != nil {
				t.Fatal(e)
			}
			actual, typed, e := tc.run()
			if e != nil {
				t.Fatal(e)
			}
			if len(actual) != len(expected.PublicInputs) {
				t.Fatalf("width %d != %d", len(actual), len(expected.PublicInputs))
			}
			for i, h := range actual {
				if h != expected.PublicInputs[i] {
					t.Fatalf("public input %d mismatch %s != %s", i, h, expected.PublicInputs[i])
				}
			}
			a, _, e := apitypes.TypedDataAndHash(typed)
			if e != nil {
				t.Fatal(e)
			}
			expected.TypedData.Types["EIP712Domain"] = typed.Types["EIP712Domain"]
			b, _, e := apitypes.TypedDataAndHash(expected.TypedData)
			if e != nil {
				t.Fatal(e)
			}
			if string(a) != string(b) {
				t.Fatal("authorization digest mismatch")
			}
		})
	}
}

func TestDrawCapacityBoundaries(t *testing.T) {
	for _, tc := range []struct{ n, c int }{{1, 1}, {2, 2}, {3, 3}, {4, 4}, {5, 5}, {6, 10}, {10, 10}, {11, 20}, {79, 80}, {80, 80}} {
		actual, e := DrawCapacity(tc.n)
		if e != nil || actual != tc.c {
			t.Fatalf("count %d capacity %d error %v", tc.n, actual, e)
		}
	}
	for _, n := range []int{0, 81} {
		if _, e := DrawCapacity(n); e == nil {
			t.Fatalf("accepted %d grants", n)
		}
	}
}
