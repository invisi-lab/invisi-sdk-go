package credit

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/invisi-labs/invisi-sdk-go/primitives"
	"github.com/invisi-labs/invisi-sdk-go/protocol"
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

func TestExplicitCreditSharedKeyPreservedAndValidated(t *testing.T) {
	b, s, f := requestBuilder(t)
	key := big.NewInt(123456789)
	input := CreatePreparation{TokenAddress: common.BigToAddress(primitives.ElementToBigInt(s.TokenAddress)), Spender: "spender", Expiry: 2100000000, AvailableCredit: big.NewInt(90), CreditSharedKey: key}
	copied := input.Clone()
	key.SetInt64(777)
	p, err := b.Create(context.Background(), copied)
	if err != nil {
		t.Fatal(err)
	}
	intent, err := p.ExportIntent()
	if err != nil {
		t.Fatal(err)
	}
	if primitives.ElementToBigInt(intent.CreatedState.CreditSharedKey).Cmp(big.NewInt(123456789)) != 0 {
		t.Fatal("explicit key changed")
	}
	commitment, err := intent.CreatedState.Commitment()
	if err != nil {
		t.Fatal(err)
	}
	for i, role := range []Role{Owner, Spender} {
		name := "owner"
		identity := intent.CreatedState.OwnerWalletIdentity
		if role == Spender {
			name = "spender"
			identity = intent.CreatedState.SpenderWalletIdentity
		}
		private := bytesHex(t, f["delivery"].(map[string]any)[name+"PrivateKey"].(string))
		got, err := Decrypt(private, intent.Deliveries[i].Ciphertext, DeliveryContext{ProtocolContext: s.ProtocolContext, TokenAddress: s.TokenAddress, Commitment: commitment, RecipientIdentity: identity}, role)
		if err != nil {
			t.Fatal(err)
		}
		if got.State.CreditSharedKey != intent.CreatedState.CreditSharedKey {
			t.Fatal("ciphertext key changed")
		}
	}

	signer, err := ethcrypto.ToECDSA(common.LeftPadBytes([]byte{9}, 32))
	if err != nil {
		t.Fatal(err)
	}
	client := creditBroadcaster(t, intent.Context, signer, common.HexToAddress("0x1234"))
	if err = client.Quote(context.Background(), p, BroadcasterQuoteOptions{MaxFeeAmount: big.NewInt(2)}); err != nil {
		t.Fatal(err)
	}
	quoted, err := p.ExportIntent()
	if err != nil {
		t.Fatal(err)
	}
	if quoted.CreatedState.CreditSharedKey != intent.CreatedState.CreditSharedKey || quoted.Create.CreditCommitment != intent.Create.CreditCommitment {
		t.Fatal("quote changed explicit key")
	}
	restored, err := Restore(context.Background(), quoted, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	resumed, err := restored.ExportIntent()
	if err != nil {
		t.Fatal(err)
	}
	if resumed.CreatedState.CreditSharedKey != intent.CreatedState.CreditSharedKey {
		t.Fatal("restored intent changed key")
	}
	for _, bad := range []*big.Int{new(big.Int), big.NewInt(-1), new(big.Int).Set(primitives.ElementToBigInt(primitives.ElementFromUint64(1)))} {
		if bad.Sign() == 1 {
			bad.SetString("21888242871839275222246405745257275088548364400416034343698204186575808495617", 10)
		}
		copied.CreditSharedKey = bad
		if _, err = b.Create(context.Background(), copied); err == nil {
			t.Fatal("noncanonical key accepted")
		}
	}
}

func TestCreateInitialPaymentBuildsPostDebitGrantAndSpenderNote(t *testing.T) {
	b, s, _ := requestBuilder(t)
	p := CreatePreparation{TokenAddress: common.BigToAddress(primitives.ElementToBigInt(s.TokenAddress)), Spender: "spender", SingleLimit: s.SingleLimit, DailyLimit: s.DailyLimit, MonthlyLimit: s.MonthlyLimit, Expiry: 2100000000, AvailableCredit: big.NewInt(90), InitialPayment: &InitialPayment{Amount: big.NewInt(7), Payload: []byte("merchant")}, Fee: &Fee{Recipient: common.HexToAddress("44"), Amount: big.NewInt(1)}}
	prepared, e := b.Create(context.Background(), p)
	if e != nil {
		t.Fatal(e)
	}
	p.InitialPayment.Amount.SetInt64(999)
	i, e := prepared.ExportIntent()
	if e != nil {
		t.Fatal(e)
	}
	if i.CreatedState.AvailableCredit != primitives.ElementFromUint64(83) || i.CreatedState.TotalSpent != primitives.ElementFromUint64(7) || i.OutputNote.Amount != primitives.ElementFromUint64(6) || i.OutputNote.OwnerAddress != s.SpenderWalletIdentity || len(i.Deliveries) != 3 {
		t.Fatal("incorrect initial debit, merchant amount, recipient or deliveries")
	}
	request, e := prepared.AuthorizationRequest()
	if e != nil || request.Operation != primitives.WalletCreateCreditInitialPayment || len(request.Words) != 16 {
		t.Fatal(request, e)
	}
	if e = prepared.CheckTimestamp(2000000000 + 86400); e == nil {
		t.Fatal("accepted next UTC period")
	}
	preview := prepared.Preview()
	if preview.NoteOutputCount != 1 || preview.GrossAmount.Cmp(big.NewInt(7)) != 0 || preview.NoteAmount.Cmp(big.NewInt(6)) != 0 {
		t.Fatal(preview)
	}
	estimate, e := prepared.EstimateGas()
	if e != nil || !estimate.Available || *estimate.GasUnits == 0 {
		t.Fatal(estimate, e)
	}
	p.InitialPayment.Amount = big.NewInt(100)
	if _, e = b.Create(context.Background(), p); e == nil {
		t.Fatal("accepted exhausted quota")
	}
}
