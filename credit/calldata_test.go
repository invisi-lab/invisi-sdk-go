package credit

import (
	"bytes"
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/invisi-labs/invisi-sdk-go/protocol/contracts"
	"math/big"
	"testing"
)

func TestRevokeHasNoProofAndBindsOriginalNamespace(t *testing.T) {
	s, _ := fixture(t)
	ctx := Context{ProtocolContext: s.ProtocolContext, TokenAddress: common.BigToAddress(big.NewInt(34)), Timestamp: 2000000000}
	p := RevokeParams{Token: ctx.TokenAddress, Owner: common.BigToAddress(big.NewInt(44)), GrantID: big.NewInt(1), Deadline: big.NewInt(2000000300)}
	data, err := p.Calldata(ctx, nil, p.Owner)
	if err != nil {
		t.Fatal(err)
	}
	id := crypto.Keccak256([]byte("revokeCredit(address,address,uint128,uint256,address,uint256,bytes)"))[:4]
	if !bytes.Equal(data[:4], id) {
		t.Fatal("wrong revoke ABI")
	}
	if _, err := p.Calldata(ctx, nil, common.Address{}); err == nil {
		t.Fatal("accepted direct revoke by another caller")
	}
	request, err := p.TypedData(ctx)
	if err != nil {
		t.Fatal(err)
	}
	before, _, err := apitypes.TypedDataAndHash(request)
	if err != nil {
		t.Fatal(err)
	}
	p.GrantID = big.NewInt(2)
	request, err = p.TypedData(ctx)
	if err != nil {
		t.Fatal(err)
	}
	after, _, err := apitypes.TypedDataAndHash(request)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(before, after) {
		t.Fatal("slot not bound to owner authorization")
	}
}

func TestLifecycleCalldataUsesCurrentSignatures(t *testing.T) {
	var compact bytes.Buffer
	if err := json.Compact(&compact, []byte(currentABIJSON)); err != nil {
		t.Fatal(err)
	}
	if compact.String() != contracts.InvisiCoreABI {
		t.Fatal("current Credit ABI differs from public Core ABI; run go generate ./protocol/contracts")
	}
	a, err := ContractABI()
	if err != nil {
		t.Fatal(err)
	}
	if common.Bytes2Hex(a.Methods["createCredit"].ID) != "4aa2aed8" {
		t.Fatal("default ABI uses legacy selector")
	}
	for _, name := range []string{"createCredit", "creditDraw", "updateCredit", "revokeCredit", "creditStatus"} {
		if _, ok := a.Methods[name]; !ok {
			t.Fatalf("missing %s", name)
		}
	}

	if len(a.Methods["createCredit"].Inputs) != 3 || len(a.Methods["updateCredit"].Inputs) != 3 {
		t.Fatal("unexpected Credit proof/params/delivery signature")
	}
	for _, name := range []string{"CreditCreated", "CreditDraw", "CreditUpdated", "CreditRevoked"} {
		if _, ok := a.Events[name]; !ok {
			t.Fatalf("missing event %s", name)
		}
	}
}
