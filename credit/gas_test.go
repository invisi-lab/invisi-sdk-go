package credit

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	"math/big"
	"reflect"
	"testing"
)

func TestBenchmarkAllCreditKindsBeforeProof(t *testing.T) {
	s, _ := fixture(t)
	c := Context{ProtocolContext: s.ProtocolContext, TokenAddress: common.HexToAddress("0x22"), Timestamp: 2000000000}
	owner := common.HexToAddress("0x2c")
	deadline := big.NewInt(2000000300)
	commitment, _ := s.Commitment()
	drawn, _ := s.Draw(primitives.ElementFromUint64(7), c.Timestamp)
	source := protocol.InputSource{TreeId: big.NewInt(1), MerkleRoot: common.HexToHash("0x01")}
	clue := make([]byte, 68)
	clue[64] = 24
	ds := []protocol.Delivery{{Ciphertext: make([]byte, 485), Clue: clue}, {Ciphertext: make([]byte, 485), Clue: clue}}
	cases := map[string]Intent{
		"create": {Context: c, Create: &CreateParams{Token: c.TokenAddress, Owner: owner, RevocationShard: 1, CreditCommitment: common.Hash(commitment.Bytes()), Deadline: deadline}, Deliveries: ds},
		"update": {Context: c, Update: &UpdateParams{Token: c.TokenAddress, Owner: owner, Source: source, Nullifier: common.Hash(drawn.Nullifier.Bytes()), CreditCommitment: common.Hash(drawn.StateCommitment.Bytes()), RevocationShard: 1, Deadline: deadline}, Deliveries: ds},
		"draw":   {Context: c, Draw: &DrawParams{Token: c.TokenAddress, Sources: []protocol.InputSource{source}, Payers: []Payer{{Payer: owner, Amount: big.NewInt(7)}}, GrantRefs: []GrantRef{{PayerIndex: big.NewInt(0), Shard: 1}}, Nullifiers: []common.Hash{common.Hash(drawn.Nullifier.Bytes())}, CreditCommitments: []common.Hash{common.Hash(drawn.StateCommitment.Bytes())}, RecoveryAmounts: []common.Hash{common.Hash(drawn.RecoveryAmount.Bytes())}, NoteCommitment: common.HexToHash("0x02"), TotalAmount: big.NewInt(7), Deadline: deadline}, Deliveries: append(append([]protocol.Delivery{}, ds...), protocol.Delivery{Ciphertext: make([]byte, 160), Clue: clue})},
		"revoke": {Context: c, Revoke: &RevokeParams{Token: c.TokenAddress, Owner: owner, Shard: 1, Deadline: deadline}},
	}
	for name, intent := range cases {
		t.Run(name, func(t *testing.T) {
			p := &Prepared{intent: intent}
			before := p.intent
			got, err := p.EstimateGas()
			if err != nil || got.Source != "benchmark" || name == "revoke" && (got.Available || got.GasUnits != nil || got.Reason == "") || name != "revoke" && (!got.Available || got.GasUnits == nil) {
				t.Fatal(got, err)
			}
			if name == "create" || name == "update" {
				originalCiphertext := p.intent.Deliveries[0].Ciphertext
				p.intent.Deliveries[0].Ciphertext = []byte{1}
				unsupported, _ := p.EstimateGas()
				if unsupported.Available || unsupported.Reason != "shape-not-covered" {
					t.Fatal(unsupported)
				}
				p.intent.Deliveries[0].Ciphertext = originalCiphertext
			}

			if !reflect.DeepEqual(before, p.intent) || p.proof != nil {
				t.Fatal("estimate mutated prepared operation")
			}
		})
	}
}
