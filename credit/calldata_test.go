package credit

import (
	"bytes"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"math/big"
	"testing"
)

func TestRevokeHasNoProofAndBindsOriginalNamespace(t *testing.T) {
	s, _ := fixture(t)
	ctx := Context{ProtocolContext: s.ProtocolContext, TokenAddress: common.BigToAddress(big.NewInt(34)), Timestamp: 2000000000}
	p := RevokeParams{Token: ctx.TokenAddress, Owner: common.BigToAddress(big.NewInt(44)), Shard: 1, Slot: 0, Deadline: big.NewInt(2000000300)}
	data, err := p.Calldata(ctx, nil, p.Owner)
	if err != nil {
		t.Fatal(err)
	}
	id := crypto.Keccak256([]byte("revokeCredit(address,address,uint64,uint16,uint256,address,uint256,bytes)"))[:4]
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
	p.Slot = 1
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
	a, err := ContractABI()
	if err != nil {
		t.Fatal(err)
	}
	if len(a.Methods) != 6 {
		t.Fatal("unexpected Credit lifecycle method count", len(a.Methods))
	}
	for _, name := range []string{"createCredit", "creditDraw", "updateCredit", "revokeCredit", "creditAllocation", "creditRevocationBitmap"} {
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
