package credit

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
)

func requestBuilder(t *testing.T) (PreparationBuilder, State, map[string]any) {
	t.Helper()
	s, f := fixture(t)
	material := f["delivery"].(map[string]any)
	keys := func(name string) primitives.ReceivingKeys {
		v := material[name+"ReceivingKeys"].(map[string]any)
		return primitives.ReceivingKeys{ViewingPublicKey: bytesHex(t, v["viewingPublicKey"].(string)), ClueKey: bytesHex(t, v["clueKey"].(string))}
	}
	owner, spender := keys("owner"), keys("spender")
	token := common.BigToAddress(primitives.ElementToBigInt(s.TokenAddress))
	b := PreparationBuilder{
		Frame: func(_ context.Context, token common.Address, deadline uint64) (Context, uint64, error) {
			if deadline == 0 {
				deadline = 2000000300
			}
			return Context{ProtocolContext: s.ProtocolContext, TokenAddress: token, Timestamp: 2000000000}, deadline, nil
		},
		Owner: func() (string, error) { return "owner", nil },
		Resolve: func(_ context.Context, address string) (primitives.ReceivingAddress, error) {
			k := owner
			identity := s.OwnerWalletIdentity
			if address == "spender" {
				k = spender
				identity = s.SpenderWalletIdentity
			}
			return primitives.ReceivingAddress{OwnerAddress: identity, ViewingPublicKey: k.ViewingPublicKey, ClueKey: k.ClueKey}, nil
		},
		Allocation: func(context.Context, Context) (uint64, uint16, error) { return 1, 2, nil },
		Input: func(_ context.Context, _ Context, _ common.Hash, _ Role) (PreparationInput, error) {
			return PreparationInput{Located: LocatedState{State: s, TreeID: 1}, Delivery: DeliveryState{State: s, OwnerReceivingKeys: owner, SpenderReceivingKeys: spender}, Source: protocol.InputSource{TreeId: big.NewInt(1), MerkleRoot: common.HexToHash("0x1")}}, nil
		},
	}
	_ = token
	return b, s, f
}
func TestRequestCreateBuildsPrivateStateAndSealedDeliveries(t *testing.T) {
	b, s, _ := requestBuilder(t)
	p, err := b.Create(context.Background(), CreatePreparation{TokenAddress: common.BigToAddress(primitives.ElementToBigInt(s.TokenAddress)), Spender: "spender", SingleLimit: s.SingleLimit, DailyLimit: s.DailyLimit, MonthlyLimit: s.MonthlyLimit, Expiry: 2100000000, AvailableCredit: big.NewInt(90), Payload: []byte("memo")})
	if err != nil {
		t.Fatal(err)
	}
	intent, err := p.ExportIntent()
	if err != nil {
		t.Fatal(err)
	}
	state := intent.CreatedState
	if state.CreditSharedKey.IsZero() || state.Nonce.IsZero() || state.AvailableCredit != primitives.ElementFromUint64(90) || len(intent.Deliveries) != 2 {
		t.Fatal("request did not construct independent opening")
	}
	if state.Usage != (Usage{}) {
		t.Fatal("new grant retained usage")
	}
	if _, err = b.Create(context.Background(), CreatePreparation{TokenAddress: intent.Context.TokenAddress, Spender: "spender", Expiry: 2100000000, Unlimited: true, AvailableCredit: big.NewInt(1)}); err == nil {
		t.Fatal("unlimited accepted finite principal")
	}
}
func TestRequestDrawPayloadOverrideAndGrossFee(t *testing.T) {
	b, s, f := requestBuilder(t)
	s.Payload = []byte("preserved")
	original := b.Input
	b.Input = func(ctx context.Context, c Context, h common.Hash, r Role) (PreparationInput, error) {
		input, err := original(ctx, c, h, r)
		input.Located.State.Payload = s.Payload
		input.Delivery.State.Payload = s.Payload
		return input, err
	}
	token := common.BigToAddress(primitives.ElementToBigInt(s.TokenAddress))
	for _, payload := range [][]byte{nil, {}} {
		p, err := b.Draw(context.Background(), DrawPreparation{TokenAddress: token, Credits: []DrawMember{{Credit: common.HexToHash("0x1"), Amount: big.NewInt(10), Payload: payload}}, Fee: &Fee{Recipient: common.HexToAddress("0x123"), Amount: big.NewInt(2)}})
		if err != nil {
			t.Fatal(err)
		}
		i, err := p.ExportIntent()
		if err != nil {
			t.Fatal(err)
		}
		if primitives.ElementToBigInt(i.OutputNote.Amount).Int64() != 8 || i.Draw.TotalAmount.Int64() != 10 {
			t.Fatal("fee changed gross debit")
		}
		next, err := s.Draw(primitives.ElementFromUint64(10), 2000000000)
		if err != nil {
			t.Fatal(err)
		}
		private := bytesHex(t, f["delivery"].(map[string]any)["ownerPrivateKey"].(string))
		got, err := Decrypt(private, i.Deliveries[0].Ciphertext, DeliveryContext{ProtocolContext: s.ProtocolContext, TokenAddress: s.TokenAddress, Commitment: next.StateCommitment, RecipientIdentity: s.OwnerWalletIdentity}, Owner)
		if err != nil {
			t.Fatal(err)
		}
		expected := "preserved"
		if payload != nil {
			expected = ""
		}
		if string(got.State.Payload) != expected {
			t.Fatal("member payload override lost")
		}
	}
}
func TestRequestRevokeNeedsOnlyOriginalReference(t *testing.T) {
	b, s, _ := requestBuilder(t)
	b.Input = nil
	p, err := b.Revoke(context.Background(), RevokePreparation{TokenAddress: common.BigToAddress(primitives.ElementToBigInt(s.TokenAddress)), Shard: 1, Slot: 2})
	if err != nil {
		t.Fatal(err)
	}
	i, _ := p.ExportIntent()
	if i.Revoke.Shard != 1 || len(i.Inputs) != 0 {
		t.Fatal("revocation invented a spend")
	}
}
