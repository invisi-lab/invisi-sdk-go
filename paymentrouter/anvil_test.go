package paymentrouter

import (
	"context"
	"encoding/json"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
	"os"
	"strings"
	"testing"
	"time"
)

func TestCompiledContractsAnvil(t *testing.T) {
	path := os.Getenv("INVISI_PAYMENT_ROUTER_FIXTURE")
	if path == "" {
		t.Skip("set INVISI_PAYMENT_ROUTER_FIXTURE to an isolated harness fixture")
	}
	raw, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var f struct {
		RPCURL                            string
		Player, TokenOut, RefundRecipient common.Address
		Deployment                        struct {
			ChainID                                                       string
			RouterAddress, CoreAddress, WETHAddress, UniswapRouterAddress common.Address
			RuntimeCodeHash                                               common.Hash
			ZeroEx                                                        *ZeroExDeployment
		}
		CorePayment vectorCall
		Quotes      []vectorQuote
		Expected    struct{ ETHSpent, ETHRefunded, FundingAmount string }
	}
	if e = json.Unmarshal(raw, &f); e != nil {
		t.Fatal(e)
	}
	if !strings.HasPrefix(f.RPCURL, "http://127.0.0.1:") {
		t.Fatal("integration harness must be loopback")
	}
	ctx := context.Background()
	rpc, e := ethclient.DialContext(ctx, f.RPCURL)
	if e != nil {
		t.Fatal(e)
	}
	defer rpc.Close()
	d := Deployment{ChainID: number(f.Deployment.ChainID), RouterAddress: f.Deployment.RouterAddress, CoreAddress: f.Deployment.CoreAddress, WETHAddress: f.Deployment.WETHAddress, UniswapRouterAddress: f.Deployment.UniswapRouterAddress, RuntimeCodeHash: &f.Deployment.RuntimeCodeHash, ZeroEx: f.Deployment.ZeroEx}
	c, e := NewClient(d, GethProvider{rpc})
	if e != nil {
		t.Fatal(e)
	}
	erc := mustABI(`[{"type":"function","name":"balanceOf","inputs":[{"type":"address"}],"outputs":[{"type":"uint256"}]},{"type":"function","name":"allowance","inputs":[{"type":"address"},{"type":"address"}],"outputs":[{"type":"uint256"}]}]`)
	read := func(to common.Address, name string, args ...any) *big.Int {
		t.Helper()
		data, e := erc.Pack(name, args...)
		if e != nil {
			t.Fatal(e)
		}
		out, e := rpc.CallContract(ctx, ethereum.CallMsg{To: &to, Data: data}, nil)
		if e != nil {
			t.Fatal(e)
		}
		v, e := erc.Methods[name].Outputs.Unpack(out)
		if e != nil {
			t.Fatal(e)
		}
		return v[0].(*big.Int)
	}
	send := func(to common.Address, data []byte, value *big.Int) common.Hash {
		t.Helper()
		var hash common.Hash
		e := rpc.Client().CallContext(ctx, &hash, "eth_sendTransaction", map[string]any{"from": f.Player, "to": to, "data": hexutil.Bytes(data), "value": (*hexutil.Big)(value), "gas": hexutil.Uint64(4000000)})
		if e != nil {
			t.Fatal(e)
		}
		for j := 0; j < 100; j++ {
			if _, e := rpc.TransactionReceipt(ctx, hash); e == nil {
				return hash
			}
			time.Sleep(10 * time.Millisecond)
		}
		t.Fatal("transaction was not mined")
		return hash
	}
	for i, q := range f.Quotes {
		t.Run(string(q.Channel)+string(rune('0'+i)), func(t *testing.T) {
			before := read(f.TokenOut, "balanceOf", d.CoreAddress)
			balance, e := rpc.BalanceAt(ctx, f.Player, nil)
			if e != nil {
				t.Fatal(e)
			}
			call, e := c.BuildPaymentTransaction(ctx, f.CorePayment.call(), q.quote(), f.RefundRecipient)
			if e != nil {
				t.Fatal(e)
			}
			sim, e := c.SimulatePayment(ctx, call, f.Player)
			if e != nil {
				t.Fatal(e)
			}
			if sim.ETHSpent.String() != f.Expected.ETHSpent || sim.ETHRefunded.String() != f.Expected.ETHRefunded {
				t.Fatal("simulation accounting", sim)
			}
			hash := send(call.To, call.Data, call.Value)
			receipt, e := rpc.TransactionReceipt(ctx, hash)
			if e != nil || receipt.Status != 1 {
				t.Fatal("execution failed", e)
			}
			after := read(f.TokenOut, "balanceOf", d.CoreAddress)
			if new(big.Int).Sub(after, before).Cmp(number(f.Expected.FundingAmount)) != 0 {
				t.Fatal("core custody")
			}
			afterBalance, e := rpc.BalanceAt(ctx, f.Player, nil)
			if e != nil {
				t.Fatal(e)
			}
			actual := new(big.Int).Sub(balance, afterBalance)
			gasCost := new(big.Int).Mul(new(big.Int).SetUint64(receipt.GasUsed), receipt.EffectiveGasPrice)
			actual.Sub(actual, gasCost)
			if actual.Cmp(number(f.Expected.ETHSpent)) != 0 {
				t.Fatal("sender ETH accounting", actual)
			}
			rb, e := rpc.BalanceAt(ctx, d.RouterAddress, nil)
			if e != nil || rb.Sign() != 0 {
				t.Fatal("residual Router ETH")
			}
			for _, token := range []common.Address{f.TokenOut, d.WETHAddress} {
				if read(token, "balanceOf", d.RouterAddress).Sign() != 0 {
					t.Fatal("residual router tokens")
				}
				for _, spender := range []common.Address{d.CoreAddress, d.UniswapRouterAddress} {
					if read(token, "allowance", d.RouterAddress, spender).Sign() != 0 {
						t.Fatal("residual approvals")
					}
				}
			}
			found := false
			event := ZeroExPaymentRouterABI().Events["Payment"]
			for _, log := range receipt.Logs {
				if log.Address == d.RouterAddress && len(log.Topics) > 0 && log.Topics[0] == event.ID {
					found = true
					args, e := event.Inputs.NonIndexed().Unpack(log.Data)
					if e != nil || args[0].(*big.Int).Cmp(number(f.Expected.FundingAmount)) != 0 || args[1].(*big.Int).Cmp(sim.ETHSpent) != 0 || args[2].(*big.Int).Cmp(sim.ETHRefunded) != 0 {
						t.Fatal("Payment event mismatch", e)
					}
				}
			}
			if !found {
				t.Fatal("missing Payment event")
			}
			// The host configures a reverting core; the SDK surfaces the full-call revert.
			config := mustABI(`[{"type":"function","name":"configure","inputs":[{"type":"bool"},{"type":"bool"}],"outputs":[]}]`)
			data, _ := config.Pack("configure", true, false)
			send(d.CoreAddress, data, bi(0))
			if _, e = c.SimulatePayment(ctx, call, f.Player); e == nil {
				t.Fatal("core failure did not revert")
			}
			failedHash := send(call.To, call.Data, call.Value)
			failed, e := rpc.TransactionReceipt(ctx, failedHash)
			if e != nil || failed.Status != 0 {
				t.Fatal("expected rollback", e)
			}
			if read(f.TokenOut, "balanceOf", d.CoreAddress).Cmp(after) != 0 {
				t.Fatal("failure changed core custody")
			}
			data, _ = config.Pack("configure", false, false)
			send(d.CoreAddress, data, bi(0))
		})
	}
}
