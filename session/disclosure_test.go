package session

import (
	"context"
	"encoding/json"
	"math/big"
	"testing"

	"github.com/invisi-lab/invisi-sdk-go/core"
	transaction "github.com/invisi-lab/invisi-sdk-go/internal/transaction"
	sdk "github.com/invisi-lab/invisi-sdk-go/sdk"
)

func TestCreditOpeningSupportsFrozenUint128Usage(t *testing.T) {
	session, token, _ := testSession(t)
	wide := core.BigIntToElement(new(big.Int).Add(new(big.Int).Lsh(big.NewInt(1), 127), big.NewInt(3)))
	state := core.CreditSpendingState{DailyLimit: core.ElementFromUint64(1), LastDay: core.ElementFromUint64(1), DailySpent: wide, LastMonth: core.ElementFromUint64(1), MonthlySpent: wide}
	leaf := seedLimitedCredit(t, session, token, state)
	snapshot, err := session.options.Store.ReadSnapshot(context.Background(), session.scope)
	if err != nil {
		t.Fatal(err)
	}
	note := snapshot.Credits[transaction.CreditKey(leaf, "owner")].State.Note
	disclosure, err := sdk.DisclosureFromCredit(note)
	if err != nil {
		t.Fatal(err)
	}
	output := sdk.DisclosureOutput{ProtocolContext: note.ProtocolContext, TokenAddress: note.TokenAddress, OutputCommitment: leaf, Role: "owner"}
	result, err := disclosure.Verify(output)
	if err != nil || result.Credit.DailySpent != wide || result.Credit.MonthlySpent != wide {
		t.Fatal("frozen uint128 opening rejected", err)
	}
	material := disclosure.Export()
	material.Data["dailySpent"] = new(big.Int).Lsh(big.NewInt(1), 128).String()
	encoded, _ := json.Marshal(material)
	corrupt, err := sdk.DisclosureFromExport(encoded)
	if err == nil {
		_, err = corrupt.Verify(output)
	}
	if err == nil {
		t.Fatal("uint128 overflow opening accepted")
	}
}
