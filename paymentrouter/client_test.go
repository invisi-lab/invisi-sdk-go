package paymentrouter

import (
	"bytes"
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/invisi-labs/invisi-sdk-go/protocol"
	"math/big"
	"testing"
)

func deployment() Deployment {
	return Deployment{ChainID: bi(31337), RouterAddress: addr(3), CoreAddress: addr(1), WETHAddress: addr(4), UniswapRouterAddress: addr(5), ZeroEx: &ZeroExDeployment{addr(7), addr(8)}}
}

type chain struct {
	now           uint64
	settler, core common.Address
	chainID       *big.Int
	reorg, fail   bool
	call          ethereum.CallMsg
	block         *big.Int
	hook          func()
}

func newChain() *chain { return &chain{now: 1000, settler: addr(6), core: addr(1), chainID: bi(31337)} }
func (c *chain) ChainID(context.Context) (*big.Int, error) {
	if c.hook != nil {
		h := c.hook
		c.hook = nil
		h()
	}
	return copyInt(c.chainID), nil
}
func (c *chain) HeaderByNumber(_ context.Context, n *big.Int) (*types.Header, error) {
	h := &types.Header{Number: bi(10), Time: c.now}
	if n != nil && c.reorg {
		h.Extra = []byte{1}
	}
	return h, nil
}
func (c *chain) CodeAt(context.Context, common.Address, *big.Int) ([]byte, error) {
	return []byte{0x60, 0}, nil
}
func (c *chain) EstimateGasAt(_ context.Context, m ethereum.CallMsg, b *big.Int) (uint64, error) {
	c.call = m
	c.block = b
	return 123456, nil
}
func (c *chain) CallContract(_ context.Context, m ethereum.CallMsg, b *big.Int) ([]byte, error) {
	if b == nil || b.Cmp(bi(10)) != 0 {
		return nil, fmt.Errorf("unpinned block")
	}
	for name, method := range gettersABI.Methods {
		if bytes.Equal(method.ID, m.Data[:4]) {
			var v any
			switch name {
			case "CHAIN_ID":
				v = c.chainID
			case "core":
				v = c.core
			case "dex":
				v = addr(5)
			case "weth":
				v = addr(4)
			case "allowanceHolder":
				v = addr(7)
			case "settlerRegistry":
				v = addr(8)
			case "ownerOf":
				v = c.settler
			}
			return method.Outputs.Pack(v)
		}
	}
	if *m.To == addr(9) {
		return quoterABI.Methods["quoteExactOutput"].Outputs.Pack(bi(60), []*big.Int{}, []uint32{}, bi(80000))
	}
	if c.fail {
		return nil, fmt.Errorf("payment reverted")
	}
	c.call = m
	c.block = b
	return routerABI.Methods["pay"].Outputs.Pack(bi(60))
}
func coreCall(t *testing.T) protocol.TransactionCall {
	t.Helper()
	p := protocol.CrossTreeTransactParams{Token: addr(2), WithdrawalAmount: bi(0), Deadline: bi(1500), FeeAmount: bi(0), NoteCommitments: [][32]byte{{31: 1}, {31: 2}}}
	f := protocol.FundingParams{Payer: addr(3), Amount: bi(10), SigNonce: bi(0), Deadline: bi(0), Permit: protocol.TokenPermit{Deadline: bi(0)}}
	d, e := protocol.BuildTransactCalldata(protocol.BuildTransactCalldataParams{Context: protocol.TransactionContext{ChainID: bi(31337), ProtocolAddress: addr(1), TokenAddress: addr(2), Now: bi(1000)}, Proof: []byte{0x12, 0x34}, Params: p, Deliveries: []protocol.Delivery{{}, {}}, Funding: f, Caller: &f.Payer})
	if e != nil {
		t.Fatal(e)
	}
	return protocol.TransactionCall{To: addr(1), Data: d.Data, Value: bi(0)}
}
func parameters() QuoteParameters {
	return QuoteParameters{TokenOut: addr(2), AmountOut: bi(10), MaxInput: bi(101), Deadline: bi(1100), SlippageBps: intPtr(100)}
}
func TestClientFullCall(t *testing.T) {
	ctx := context.Background()
	rpc := newChain()
	c, e := NewClient(deployment(), rpc)
	if e != nil {
		t.Fatal(e)
	}
	a, e := NewUniswapV3Adapter(rpc, addr(9), []V3Route{{[]common.Address{addr(4), addr(2)}, []uint32{3000}}})
	if e != nil {
		t.Fatal(e)
	}
	q, e := c.Quote(ctx, a, parameters())
	if e != nil {
		t.Fatal(e)
	}
	if q.MaxInput.Cmp(bi(61)) != 0 {
		t.Fatal("slippage rounding")
	}
	call, e := c.BuildPaymentTransaction(ctx, coreCall(t), q, addr(10))
	if e != nil {
		t.Fatal(e)
	}
	r, e := c.SimulatePayment(ctx, call, addr(10))
	if e != nil {
		t.Fatal(e)
	}
	if r.ETHSpent.Cmp(bi(60)) != 0 || r.ETHRefunded.Cmp(bi(1)) != 0 || r.GasEstimate != 123456 || rpc.call.From != addr(10) || rpc.call.Value.Cmp(bi(61)) != 0 {
		t.Fatal("wrong simulation", r)
	}
	rpc.reorg = true
	if _, e = c.BuildPaymentTransaction(ctx, coreCall(t), q, addr(10)); e == nil {
		t.Fatal("accepted reorg")
	}
	rpc.reorg = false
	rpc.core = addr(5)
	if _, e = c.Quote(ctx, a, parameters()); e == nil {
		t.Fatal("accepted deployment mismatch")
	}
}
func TestClientBindings(t *testing.T) {
	ctx := context.Background()
	rpc := newChain()
	c, _ := NewClient(deployment(), rpc)
	path, _ := EncodeV3ExactOutputPath([]common.Address{addr(4), addr(2)}, []uint32{3000})
	q := SwapQuote{Channel: UniswapV3, ChainID: bi(31337), RouterAddress: addr(3), TokenOut: addr(2), AmountOut: bi(10), EstimatedInput: bi(60), MaxInput: bi(100), Deadline: bi(1100), ReversePath: path}
	for _, change := range []string{"core", "value", "proof", "amount", "token", "deadline", "recipient", "budget", "chain"} {
		p := coreCall(t)
		v := cloneQuote(q)
		recipient := addr(10)
		switch change {
		case "core":
			p.To = addr(5)
		case "value":
			p.Value = bi(1)
		case "proof":
			p.Data = []byte{1}
		case "amount":
			v.AmountOut = bi(11)
		case "token":
			v.TokenOut = addr(5)
		case "deadline":
			v.Deadline = bi(999)
		case "recipient":
			recipient = addr(3)
		case "budget":
			v.MaxInput = bi(59)
		case "chain":
			v.ChainID = bi(1)
		}
		if _, e := c.BuildPaymentTransaction(ctx, p, v, recipient); e == nil {
			t.Fatal("accepted", change)
		}
	}
}

func intPtr(n int) *int { return &n }
