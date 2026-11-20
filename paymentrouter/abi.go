package paymentrouter

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/invisi-labs/invisi-sdk-go/protocol"
)

var gettersABI = makeGetters()

func makeGetters() abi.ABI {
	items := []map[string]any{}
	for _, n := range []string{"CHAIN_ID", "core", "dex", "weth", "allowanceHolder", "settlerRegistry", "ownerOf"} {
		out := "address"
		if n == "CHAIN_ID" {
			out = "uint256"
		}
		in := []map[string]string{}
		if n == "ownerOf" {
			in = append(in, map[string]string{"type": "uint256"})
		}
		items = append(items, map[string]any{"type": "function", "name": n, "inputs": in, "outputs": []map[string]string{{"type": out}}})
	}
	b, _ := json.Marshal(items)
	return mustABI(string(b))
}

var quoterABI = mustABI(`[{"type":"function","name":"quoteExactOutput","inputs":[{"type":"bytes"},{"type":"uint256"}],"outputs":[{"type":"uint256"},{"type":"uint160[]"},{"type":"uint32[]"},{"type":"uint256"}]}]`)
var coreMethod = func() abi.Method {
	a, e := protocol.ProtocolABI()
	if e != nil {
		panic(e)
	}
	for _, m := range a.Methods {
		if m.RawName == "transact" && len(m.Inputs) == 2 && m.Inputs[1].Type.T == abi.TupleTy {
			return m
		}
	}
	panic("current transact tuple unavailable")
}()
var routerABI = func() abi.ABI {
	output, _ := abi.NewType("uint256", "", nil)
	path, _ := abi.NewType("bytes", "", nil)
	target, _ := abi.NewType("address", "", nil)
	pay := append(abi.Arguments(nil), coreMethod.Inputs...)
	pay = append(pay, abi.Argument{Name: "reversePath", Type: path}, abi.Argument{Name: "swapDeadline", Type: output}, abi.Argument{Name: "refundRecipient", Type: target})
	zero := append(abi.Arguments(nil), coreMethod.Inputs...)
	zero = append(zero, abi.Argument{Name: "target", Type: target}, abi.Argument{Name: "data", Type: path}, abi.Argument{Name: "swapDeadline", Type: output}, abi.Argument{Name: "refundRecipient", Type: target})
	a := abi.ABI{Methods: map[string]abi.Method{}}
	for n, args := range map[string]abi.Arguments{"pay": pay, "payWithZeroEx": zero} {
		a.Methods[n] = abi.NewMethod(n, n, abi.Function, "payable", false, true, args, abi.Arguments{{Name: "ethSpent", Type: output}})
	}
	return a
}()
