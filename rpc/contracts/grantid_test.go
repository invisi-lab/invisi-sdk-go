package contracts

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
	"testing"
)

func TestGrantIDEventBindingsPreserveFullUint128(t *testing.T) {
	schema, err := InvisiCoreMetaData.GetAbi()
	if err != nil {
		t.Fatal(err)
	}
	filter, err := NewInvisiCoreFilterer(common.HexToAddress("11"), nil)
	if err != nil {
		t.Fatal(err)
	}
	id := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1))
	owner, token := common.HexToAddress("22"), common.HexToAddress("33")
	revoked := schema.Events["CreditRevoked"]
	data, err := revoked.Inputs.NonIndexed().Pack(common.Address{}, big.NewInt(0))
	if err != nil {
		t.Fatal(err)
	}
	topics, err := abi.MakeTopics([]any{owner}, []any{token}, []any{id})
	if err != nil {
		t.Fatal(err)
	}
	log := types.Log{Topics: []common.Hash{revoked.ID, topics[0][0], topics[1][0], topics[2][0]}, Data: data}
	parsed, err := filter.ParseCreditRevoked(log)
	if err != nil || parsed.GrantId.Cmp(id) != 0 || parsed.Owner != owner || parsed.Token != token {
		t.Fatalf("revocation binding lost grant identity: %v", err)
	}
	created := schema.Events["CreditCreated"]
	data, err = created.Inputs.NonIndexed().Pack(id, [32]byte{1}, uint32(0), []IInvisiCoreDelivery{}, big.NewInt(1), common.Address{}, big.NewInt(0))
	if err != nil {
		t.Fatal(err)
	}
	topics, err = abi.MakeTopics([]any{token}, []any{owner}, []any{big.NewInt(1)})
	if err != nil {
		t.Fatal(err)
	}
	log = types.Log{Topics: []common.Hash{created.ID, topics[0][0], topics[1][0], topics[2][0]}, Data: data}
	creation, err := filter.ParseCreditCreated(log)
	if err != nil || creation.GrantId.Cmp(id) != 0 {
		t.Fatalf("creation binding lost grant ID: %v", err)
	}
}
