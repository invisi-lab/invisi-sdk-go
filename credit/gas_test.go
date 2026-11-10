package credit

import (
	"bytes"
	"context"
	"crypto/sha256"
	hexencoding "encoding/hex"
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	"github.com/invisi-lab/invisi-sdk-go/prover"
)

type creditEstimator struct {
	call  ethereum.CallMsg
	calls int
	err   error
}

func (e *creditEstimator) EstimateGas(_ context.Context, call ethereum.CallMsg) (uint64, error) {
	e.call = call
	e.calls++
	return 456789, e.err
}

func TestFinalGasEstimatesAllCreditKinds(t *testing.T) {
	s, _ := fixture(t)
	c := Context{ProtocolContext: s.ProtocolContext, TokenAddress: common.HexToAddress("0x22"), Timestamp: 2000000000}
	owner := common.HexToAddress("0x2c")
	deadline := big.NewInt(2000000300)
	commitment, _ := s.Commitment()
	drawn, _ := s.Draw(primitives.ElementFromUint64(7), c.Timestamp)
	source := protocol.InputSource{TreeId: big.NewInt(1), MerkleRoot: common.HexToHash("0x01")}
	clue := make([]byte, 68)
	clue[64] = 24
	ds := []protocol.Delivery{{Ciphertext: []byte{0x82}, Clue: clue}, {Ciphertext: []byte{0x82}, Clue: clue}}
	cases := map[string]Intent{
		"create": {Context: c, Create: &CreateParams{Token: c.TokenAddress, Owner: owner, RevocationShard: 1, CreditCommitment: common.Hash(commitment.Bytes()), Deadline: deadline}, Deliveries: ds},
		"update": {Context: c, Update: &UpdateParams{Token: c.TokenAddress, Owner: owner, Source: source, Nullifier: common.Hash(drawn.Nullifier.Bytes()), CreditCommitment: common.Hash(drawn.StateCommitment.Bytes()), RevocationShard: 1, Deadline: deadline}, Deliveries: ds},
		"draw":   {Context: c, Draw: &DrawParams{Token: c.TokenAddress, Sources: []protocol.InputSource{source}, Payers: []Payer{{Payer: owner, Amount: big.NewInt(7)}}, GrantRefs: []GrantRef{{PayerIndex: big.NewInt(0), Shard: 1}}, Nullifiers: []common.Hash{common.Hash(drawn.Nullifier.Bytes())}, CreditCommitments: []common.Hash{common.Hash(drawn.StateCommitment.Bytes())}, RecoveryAmounts: []common.Hash{common.Hash(drawn.RecoveryAmount.Bytes())}, NoteCommitment: common.HexToHash("0x02"), TotalAmount: big.NewInt(7), Deadline: deadline}, Deliveries: append(append([]protocol.Delivery{}, ds...), protocol.Delivery{})},
		"revoke": {Context: c, Revoke: &RevokeParams{Token: c.TokenAddress, Owner: owner, Shard: 1, Deadline: deadline}},
	}
	for name, intent := range cases {
		t.Run(name, func(t *testing.T) {
			p := &Prepared{intent: intent}
			estimator := &creditEstimator{}
			if name != "revoke" {
				if _, err := p.EstimateGas(context.Background(), estimator, owner); err == nil || estimator.calls != 0 {
					t.Fatal("unproved Credit estimated")
				}
				statement, err := p.expected()
				if err != nil {
					t.Fatal(err)
				}
				// This fixture tests orchestration; placeholder proof bytes do not establish cryptographic acceptance.
				proof := make([]byte, 32)
				digest := sha256.Sum256(proof)
				p.proof = &prover.ProofBundle{FormatVersion: 1, BBVersion: "6.0.0", CircuitName: statement.CircuitName, Target: statement.Target, ArtifactSHA256: statement.ArtifactSHA256, VKSHA256: statement.VKSHA256, Proof: proof, ProofSHA256: hexencoding.EncodeToString(digest[:]), PublicInputs: statement.ExpectedPublicInputs}
			}
			expected, err := p.Calldata(owner)
			if err != nil {
				t.Fatal(err)
			}
			got, err := p.EstimateGas(context.Background(), estimator, owner)
			if err != nil || got.Source != "rpc" || got.GasUnits != 456789 || estimator.call.From != owner || estimator.call.To == nil || *estimator.call.To != common.BigToAddress(primitives.ElementToBigInt(c.ProtocolAddress)) || !bytes.Equal(estimator.call.Data, expected) || estimator.call.Value.Sign() != 0 {
				t.Fatal("incorrect final Credit call", got, err)
			}
			estimator.err = errors.New("chain failure")
			if _, err = p.EstimateGas(context.Background(), estimator, owner); !errors.Is(err, estimator.err) {
				t.Fatal("provider error lost", err)
			}
		})
	}
}
