package deployments_test

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/invisi-lab/invisi-sdk-go/contracts"
	"github.com/invisi-lab/invisi-sdk-go/deployments"
)

type factoryRPC struct {
	t                                  *testing.T
	factory, operator, token, instance common.Address
}

func (f *factoryRPC) Call(_ context.Context, args map[string]interface{}, _ string) (hexutil.Bytes, error) {
	f.t.Helper()
	if args["to"] != f.factory.Hex() {
		f.t.Fatalf("wrong factory destination: %v", args["to"])
	}
	encoded, ok := args["input"].(string)
	if !ok {
		encoded, _ = args["data"].(string)
	}
	data, err := hexutil.Decode(encoded)
	if err != nil {
		f.t.Fatal(err)
	}
	abi, err := contracts.InvisiFactoryMetaData.GetAbi()
	if err != nil {
		f.t.Fatal(err)
	}
	method := abi.Methods["getInstanceAddress"]
	if len(data) < 4 || string(data[:4]) != string(method.ID) {
		f.t.Fatal("wrong factory method")
	}
	inputs, err := method.Inputs.Unpack(data[4:])
	if err != nil {
		f.t.Fatal(err)
	}
	if inputs[0].(common.Address) != f.operator || inputs[1].(common.Address) != f.token {
		f.t.Fatal("operator/token argument mismatch")
	}
	return method.Outputs.Pack(f.instance)
}
func (f *factoryRPC) GetCode(context.Context, common.Address, *big.Int) (hexutil.Bytes, error) {
	return []byte{1}, nil
}
func TestResolveInstanceAddressCurrentFactoryQuery(t *testing.T) {
	service := &factoryRPC{t: t, factory: common.HexToAddress("0x1234"), operator: common.HexToAddress("0x2345"), token: common.HexToAddress("0x3456"), instance: common.HexToAddress("0x4567")}
	server := rpc.NewServer()
	if err := server.RegisterName("eth", service); err != nil {
		t.Fatal(err)
	}
	defer server.Stop()
	client := ethclient.NewClient(rpc.DialInProc(server))
	defer client.Close()
	got, err := deployments.ResolveInstanceAddress(client, service.factory, service.operator, service.token)
	if err != nil {
		t.Fatal(err)
	}
	if got != service.instance {
		t.Fatalf("got %s want %s", got, service.instance)
	}
	if _, err = deployments.ResolveInstanceAddress(client, common.Address{}, service.operator, service.token); err == nil {
		t.Fatal("accepted zero factory")
	}
	if _, err = deployments.ResolveInstanceAddress(nil, service.factory, service.operator, service.token); err == nil {
		t.Fatal("accepted nil client")
	}
}
func TestDeploymentRegistryContainsNoUnverifiedAddresses(t *testing.T) {
	for _, id := range deployments.GetSupportedChainIDs() {
		d, err := deployments.GetDeployment(id)
		if err != nil {
			t.Fatal(err)
		}
		if d.InvisiFactory.Address != (common.Address{}) || d.VerifierRouter.Address != (common.Address{}) || d.CreditVerifierRouter.Address != (common.Address{}) {
			t.Fatalf("unverified address for chain %d", id)
		}
		if len(d.TransactUserVerifiers) != 0 || len(d.TransactAggregateVerifiers) != 0 || len(d.CreditVerifiers) != 0 {
			t.Fatal("invented verifier deployment")
		}
		d.CreditVerifiers["create_credit"] = deployments.ContractDeployment{Address: common.HexToAddress("0x1234")}
		again, _ := deployments.GetDeployment(id)
		if len(again.CreditVerifiers) != 0 {
			t.Fatal("registry leaked mutable map")
		}
	}
}
