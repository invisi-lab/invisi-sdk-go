package settlement

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestDeferredCreditRoundTripAndUnsignedIntent(t *testing.T) {
	zero := big.NewInt(0)
	deadline := big.NewInt(2000000000)
	token := common.HexToAddress("0x2222222222222222222222222222222222222222")
	owner := common.HexToAddress("0x3333333333333333333333333333333333333333")
	requests := []*Request{
		{Method: "createCredit", Proof: []byte{1}, Creation: CreateCreditParams{Token: token, Owner: owner, RevocationShard: 1, CreditCommitment: [32]byte{1}, Deadline: deadline}},
		{Method: "creditDraw", Proof: []byte{1}, Draw: CreditDrawParams{Token: token, Payers: []CreditPayer{{owner, big.NewInt(7)}}, GrantRefs: []CreditGrantRef{{zero, 1, 0}}, Nullifiers: [][32]byte{{2}}, CreditCommitments: [][32]byte{{3}}, RecoveryAmounts: [][32]byte{{4}}, NoteCommitment: [32]byte{5}, TotalAmount: big.NewInt(7), Deadline: deadline}},
		{Method: "updateCredit", Proof: []byte{1}, Update: UpdateCreditParams{Token: token, Owner: owner, Source: InputSource{big.NewInt(1), [32]byte{6}}, Nullifier: [32]byte{2}, CreditCommitment: [32]byte{8}, RevocationShard: 1, Deadline: deadline}},
		{Method: "revokeCredit", Revocation: RevokeCreditParams{Token: token, Owner: owner, Shard: 1, Deadline: deadline, Signature: []byte{1}}},
	}
	for _, r := range requests {
		r.Creation.FeeAmount = zero
		r.Draw.FeeAmount = zero
		r.Update.FeeAmount = zero
		r.Revocation.FeeAmount = zero
		t.Run(r.Method, func(t *testing.T) {
			data, e := Encode(r)
			if e != nil {
				t.Fatal(e)
			}
			decoded, e := DecodeSingle(data)
			if e != nil {
				t.Fatal(e)
			}
			again, e := Encode(decoded)
			if e != nil || !bytes.Equal(data, again) {
				t.Fatal("noncanonical roundtrip", e)
			}
			s := decoded.Statement()
			if s.Token != token || s.FeeRecipient != (common.Address{}) || s.FeeAmount.Sign() != 0 {
				t.Fatal("default Credit fee", s)
			}
			hash, e := Hash(decoded)
			if e != nil {
				t.Fatal(e)
			}
			decoded.Proof = []byte{9}
			decoded.Revocation.Signature = []byte{9}
			unsigned, e := Hash(decoded)
			if e != nil || hash != unsigned {
				t.Fatal("proof/signature must not change intent", e)
			}
			switch decoded.Method {
			case "createCredit":
				decoded.Creation.RevocationSlot++
			case "creditDraw":
				decoded.Draw.GrantRefs[0].Slot++
			case "updateCredit":
				decoded.Update.RevocationSlot++
			case "revokeCredit":
				decoded.Revocation.Slot++
			}
			changed, e := Hash(decoded)
			if e != nil || changed == hash {
				t.Fatal("grant slot must bind business intent", e)
			}
		})
	}
}
