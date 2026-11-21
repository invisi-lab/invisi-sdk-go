package credit

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-labs/invisi-sdk-go/primitives"
	"math/big"
	"strings"
	"testing"
)

func TestGrantIDBoundsAndImmutableTransitions(t *testing.T) {
	s, _ := fixture(t)
	max := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1))
	for _, n := range []*big.Int{new(big.Int), new(big.Int).Lsh(big.NewInt(1), 128)} {
		bad := s
		bad.GrantID = primitives.BigIntToElement(n)
		if bad.Validate() == nil {
			t.Fatalf("accepted invalid grant ID %s", n)
		}
	}
	s.GrantID = primitives.BigIntToElement(max)
	if err := s.Validate(); err != nil {
		t.Fatal(err)
	}
	next, err := s.Draw(primitives.ElementFromUint64(7), 2000000000)
	if err != nil {
		t.Fatal(err)
	}
	if next.State.GrantID != s.GrantID {
		t.Fatal("draw changed grant ID")
	}
	recovered, err := s.Recover(next.Nullifier, next.StateCommitment, next.RecoveryAmount, 2000000000)
	if err != nil || recovered.GrantID != s.GrantID {
		t.Fatalf("recovery changed grant ID: %v", err)
	}
}

func TestGrantIDCanonicalDecimalAndGeneration(t *testing.T) {
	for _, raw := range []string{"0", "-1", "+1", "01", " 1", "1.0", "0x1", "340282366920938463463374607431768211456"} {
		if _, err := ParseGrantID(raw); err == nil {
			t.Fatalf("accepted %q", raw)
		}
	}
	max := "340282366920938463463374607431768211455"
	if id, err := ParseGrantID(max); err != nil || id.String() != max {
		t.Fatal(id, err)
	}
	seen := map[string]bool{}
	for range 32 {
		id, err := GenerateGrantID()
		if err != nil || !ValidGrantID(id) || seen[id.String()] {
			t.Fatal("invalid/repeated random ID", id, err)
		}
		seen[id.String()] = true
	}
}

func TestCreationChecksScopedStatusAndClonesID(t *testing.T) {
	b, s, _ := requestBuilder(t)
	id := new(big.Int).Add(new(big.Int).Lsh(big.NewInt(1), 100), big.NewInt(7))
	want := new(big.Int).Set(id)
	token := common.BigToAddress(primitives.ElementToBigInt(s.TokenAddress))
	owner := common.BigToAddress(primitives.ElementToBigInt(s.OwnerWalletIdentity))
	p := CreatePreparation{TokenAddress: token, Spender: "spender", Expiry: 2100000000, AvailableCredit: big.NewInt(400), GrantID: id}
	b.Status = func(_ context.Context, c Context, gotOwner common.Address, gotID *big.Int) (uint8, error) {
		if gotOwner != owner || c.TokenAddress != token || gotID.Cmp(want) != 0 {
			t.Fatal("status scope differs from frozen ID")
		}
		id.SetInt64(1)
		return 0, nil
	}
	prepared, err := b.Create(context.Background(), p)
	if err != nil {
		t.Fatal(err)
	}
	intent, _ := prepared.ExportIntent()
	if intent.Create.GrantID.Cmp(want) != 0 || primitives.ElementToBigInt(intent.CreatedState.GrantID).Cmp(want) != 0 {
		t.Fatal("caller mutation changed frozen grant ID")
	}
	for _, status := range []uint8{1, 2, 3} {
		b.Status = func(context.Context, Context, common.Address, *big.Int) (uint8, error) { return status, nil }
		if _, err = b.Create(context.Background(), p); err == nil {
			t.Fatalf("creation accepted status %d", status)
		}
	}
}

func TestPreRevokeAllowsUnknownAndRejectsDisabledBeforeFees(t *testing.T) {
	b, s, _ := requestBuilder(t)
	p := RevokePreparation{TokenAddress: common.BigToAddress(primitives.ElementToBigInt(s.TokenAddress)), Owner: common.BigToAddress(primitives.ElementToBigInt(s.OwnerWalletIdentity)), GrantID: big.NewInt(77)}
	for _, status := range []uint8{0, 1} {
		b.Status = func(context.Context, Context, common.Address, *big.Int) (uint8, error) { return status, nil }
		prepared, err := b.Revoke(context.Background(), p)
		if err != nil {
			t.Fatal(err)
		}
		intent, _ := prepared.ExportIntent()
		if len(intent.Inputs) != 0 {
			t.Fatal("pre-revoke needs private state")
		}
	}
	b.Status = func(context.Context, Context, common.Address, *big.Int) (uint8, error) { return 2, nil }
	p.Fee = &Fee{Amount: big.NewInt(-1)}
	if _, err := b.Revoke(context.Background(), p); err == nil || !strings.Contains(err.Error(), "already disabled") {
		t.Fatalf("disabled ID did not fail before fee validation: %v", err)
	}
}
