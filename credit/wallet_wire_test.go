package credit

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	wire "github.com/invisi-lab/invisi-sdk-go/wallet/protocol"
)

func TestCreditPreparedWalletWireAuthorization(t *testing.T) {
	s, f := fixture(t)
	raw := common.LeftPadBytes([]byte{1}, 32)
	key, _ := crypto.ToECDSA(raw)
	owner := crypto.PubkeyToAddress(key.PublicKey)
	s.OwnerWalletIdentity = primitives.BigIntToElement(owner.Big())
	s.SpenderWalletIdentity = s.OwnerWalletIdentity
	commitment, _ := s.Commitment()
	c := Context{ProtocolContext: s.ProtocolContext, TokenAddress: common.BigToAddress(big.NewInt(34)), Timestamp: 2000000000}
	d := f["delivery"].(map[string]any)
	clue := make([]byte, 68)
	clue[64] = 24
	deliveries := []protocol.Delivery{{Ciphertext: bytesHex(t, d["ownerCiphertext"].(string)), Clue: clue}, {Ciphertext: bytesHex(t, d["spenderCiphertext"].(string)), Clue: clue}}
	p, err := PrepareCreate(c, CreateParams{Token: c.TokenAddress, Owner: owner, RevocationShard: 1, CreditCommitment: common.Hash(commitment.Bytes()), Deadline: big.NewInt(2000000300)}, deliveries, s)
	if err != nil {
		t.Fatal(err)
	}
	input := LocatedState{State: s, TreeID: 1, Debit: primitives.ElementFromUint64(7)}
	root, err := input.root()
	if err != nil {
		t.Fatal(err)
	}
	drawn, err := s.Draw(primitives.ElementFromUint64(7), c.Timestamp)
	if err != nil {
		t.Fatal(err)
	}
	identity := primitives.ElementFromUint64(1)
	noteCommitment, err := primitives.ComputeNoteCommitment(s.ChainID, s.ProtocolAddress, identity, s.TokenAddress, primitives.ElementFromUint64(7))
	if err != nil {
		t.Fatal(err)
	}
	draw, err := PrepareDraw(c, DrawParams{Token: c.TokenAddress, Sources: []protocol.InputSource{{TreeId: big.NewInt(1), MerkleRoot: root}}, Payers: []Payer{{Payer: owner, Amount: big.NewInt(7)}}, GrantRefs: []GrantRef{{PayerIndex: big.NewInt(0), Shard: 1}}, Nullifiers: []common.Hash{common.Hash(drawn.Nullifier.Bytes())}, CreditCommitments: []common.Hash{common.Hash(drawn.StateCommitment.Bytes())}, RecoveryAmounts: []common.Hash{common.Hash(drawn.RecoveryAmount.Bytes())}, NoteCommitment: common.Hash(noteCommitment.Bytes()), TotalAmount: big.NewInt(7), Deadline: big.NewInt(2000000300)}, append(deliveries, protocol.Delivery{}), []LocatedState{input}, identity)
	if err != nil {
		t.Fatal(err)
	}
	change := Change{Policy: s.Policy, Expiry: s.Expiry, IncreaseAvailableCredit: primitives.ElementFromUint64(5)}
	updated, err := s.Update(change, 2000000300)
	if err != nil {
		t.Fatal(err)
	}
	update, err := PrepareUpdate(c, UpdateParams{Token: c.TokenAddress, Owner: owner, Source: protocol.InputSource{TreeId: big.NewInt(1), MerkleRoot: root}, Nullifier: common.Hash(updated.Nullifier.Bytes()), CreditCommitment: common.Hash(updated.StateCommitment.Bytes()), RevocationShard: 1, Deadline: big.NewInt(2000000300)}, deliveries, input, change)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range []*Prepared{p, draw, update} {
		q, err := p.ExportWalletRequest("credit-wire")

		if err != nil {
			t.Fatal(err)
		}
		var params wire.AuthorizeParams
		if err = wire.Decode(q.Params, &params); err != nil {
			t.Fatal(err)
		}
		request, err := params.Core()
		if err != nil {
			t.Fatal(err)
		}
		signature, _ := primitives.SignWalletAuthorization(raw, request)
		digest, _ := primitives.ComputeWalletAuthorizationDigest(request)
		response := wire.Response{Version: q.Version, ID: q.ID, OwnerAddress: q.OwnerAddress, Method: q.Method, Result: wire.Encode(wire.AuthorizeResult{SpendingPublicKey: wire.Hex(crypto.FromECDSAPub(&key.PublicKey)), Signature: wire.Hex(signature), Digest: digest.Hex()})}
		wrong := response
		wrong.ID = "other"
		if err = p.WithWalletResponse(q, wrong); err == nil {
			t.Fatal("accepted replayed Credit response")
		}
		if err = p.WithWalletResponse(q, response); err != nil {
			t.Fatal(err)
		}
	}
}
