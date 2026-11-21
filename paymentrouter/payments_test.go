package paymentrouter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
	"os"
	"testing"
)

type canonicalChain struct {
	*chain
	nonce *big.Int
}

func (c *canonicalChain) CodeAt(ctx context.Context, a common.Address, b *big.Int) ([]byte, error) {
	if a == addr(9) {
		return nil, nil
	}
	return c.chain.CodeAt(ctx, a, b)
}
func (c *canonicalChain) CallContract(ctx context.Context, m ethereum.CallMsg, b *big.Int) ([]byte, error) {
	if bytes.Equal(m.Data[:4], canonical.Methods["nonces"].ID) {
		return canonical.Methods["nonces"].Outputs.Pack(c.nonce)
	}
	for _, name := range []string{"payNative", "payToken", "payTokenFor"} {
		if bytes.Equal(m.Data[:4], canonical.Methods[name].ID) {
			c.call = m
			return canonical.Methods[name].Outputs.Pack(bi(100))
		}
	}
	return c.chain.CallContract(ctx, m, b)
}
func TestCanonicalClientModesAndNonce(t *testing.T) {
	ctx := context.Background()
	provider := &canonicalChain{newChain(), bi(7)}
	c, e := NewClient(deployment(), provider)
	if e != nil {
		t.Fatal(e)
	}
	for _, mode := range []uint8{DirectNative, DirectToken, SponsoredNative} {
		t.Run(fmt.Sprint(mode), func(t *testing.T) {
			o := PaymentOptions{Core: coreCall(t), Mode: mode, Payer: addr(9), RefundRecipient: addr(9), AmountInMaximum: bi(1000), Deadline: bi(1900)}
			if mode == DirectToken {
				o.TokenIn = addr(2)
				o.Route = Route{Channel: ChannelNone}
			} else {
				o.Route = Route{Channel: ChannelUniswapV3}
				o.Route.Path, _ = EncodeV3ExactOutputPath([]common.Address{addr(4), addr(2)}, []uint32{3000})
				if mode == SponsoredNative {
					o.Relayer = addr(10)
				}
			}
			call, e := c.BuildPayment(ctx, o)
			if e != nil {
				t.Fatal(e)
			}
			p, e := DecodeRouterCall(call.Data)
			if e != nil || p.Intent.Nonce.Cmp(bi(7)) != 0 {
				t.Fatal(e)
			}
			if e = c.ValidatePayment(ctx, call, addr(9)); e != nil {
				t.Fatal(e)
			}
			result, e := c.SimulateCanonicalPayment(ctx, call, addr(9))
			if e != nil || result.GasEstimate != 123456 || result.AmountIn.Cmp(bi(100)) != 0 {
				t.Fatal(result, e)
			}
			if e = c.ValidatePayment(ctx, call, addr(11)); e == nil {
				t.Fatal("wrong sender accepted")
			}
			provider.nonce = bi(8)
			if e = c.ValidatePayment(ctx, call, addr(9)); e == nil {
				t.Fatal("stale nonce accepted")
			}
			provider.nonce = bi(7)
			call.Value = bi(123)
			if e = c.ValidatePayment(ctx, call, addr(9)); e == nil {
				t.Fatal("wrong value accepted")
			}
		})
	}
}
func TestSigningAuthorizationAndApproval(t *testing.T) {
	key, e := crypto.GenerateKey()
	if e != nil {
		t.Fatal(e)
	}
	player := crypto.PubkeyToAddress(key.PublicKey)
	td, e := PermitTypedData(bi(31337), addr(2), "Token", "1", player, addr(3), bi(100), bi(4), bi(1900))
	if e != nil {
		t.Fatal(e)
	}
	sig, e := SignTypedData(key, td)
	if e != nil {
		t.Fatal(e)
	}
	permit, e := PermitFromSignature(bi(1900), sig)
	if e != nil || permit.V < 27 {
		t.Fatal(e)
	}
	h, e := typedDigest(td, nil)
	if e != nil || VerifySignature(h, sig, player) != nil {
		t.Fatal("permit signature failed", e)
	}
	sig[10] ^= 1
	if VerifySignature(h, sig, player) == nil {
		t.Fatal("signature substitution accepted")
	}
	auth, e := SignDelegation(key, bi(31337), addr(12), 4)
	if e != nil {
		t.Fatal(e)
	}
	a, e := auth.Authority()
	if e != nil || a != player {
		t.Fatal(e)
	}
	c, _ := NewClient(deployment(), newChain())
	approval, e := c.ApprovalTransaction(addr(2), bi(100))
	if e != nil || approval.To != addr(2) || approval.Value.Sign() != 0 {
		t.Fatal(approval, e)
	}
	if common.BytesToAddress(approval.Data[16:36]) != addr(3) {
		t.Fatal("wrong approval spender")
	}
}
func TestEnvelopeGoldenHashes(t *testing.T) {
	var rows []struct {
		ChainID  string
		Core     common.Address
		Envelope RouterEnvelope
		Hash     common.Hash
	}
	b, e := os.ReadFile("testdata/payment-router-envelope-vectors.json")
	if e != nil {
		t.Fatal(e)
	}
	if e = json.Unmarshal(b, &rows); e != nil {
		t.Fatal(e)
	}
	for _, row := range rows {
		h, e := EnvelopeHash(bi(31337), row.Core, row.Envelope)
		if e != nil || h != row.Hash {
			t.Fatal(h, e)
		}
	}
}

type accountChain struct {
	*canonicalChain
	player, impl       common.Address
	wrongCode          bool
	undelegated        bool
	actionNonce        *big.Int
	authorizationNonce uint64
}

func (c *accountChain) CodeAt(ctx context.Context, a common.Address, b *big.Int) ([]byte, error) {
	if a == c.player {
		if c.undelegated {
			return nil, nil
		}
		return types.AddressToDelegation(c.impl), nil
	}
	if a == c.impl && c.wrongCode {
		return []byte{0xff}, nil
	}
	return []byte{0x60, 0}, nil
}
func (c *accountChain) StorageAt(context.Context, common.Address, common.Hash, *big.Int) ([]byte, error) {
	return common.LeftPadBytes(c.actionNonce.Bytes(), 32), nil
}
func (c *accountChain) NonceAt(context.Context, common.Address, *big.Int) (uint64, error) {
	return c.authorizationNonce, nil
}
func (c *accountChain) CallContract(ctx context.Context, m ethereum.CallMsg, b *big.Int) ([]byte, error) {
	for name, method := range account.Methods {
		if !bytes.Equal(m.Data[:4], method.ID) {
			continue
		}
		switch name {
		case "router":
			return method.Outputs.Pack(addr(3))
		case "implementation":
			return method.Outputs.Pack(c.impl)
		case "routerCodeHash":
			return method.Outputs.Pack(crypto.Keccak256Hash([]byte{0x60, 0}))
		case "nonce":
			return method.Outputs.Pack(c.actionNonce)
		case "executeNativePayment":
			return method.Outputs.Pack(bi(100))
		}
	}
	return c.canonicalChain.CallContract(ctx, m, b)
}
func TestSponsoredPolicyAndOuterSimulation(t *testing.T) {
	ctx := context.Background()
	key, _ := crypto.ToECDSA(common.LeftPadBytes(bi(12345).Bytes(), 32))
	player := crypto.PubkeyToAddress(key.PublicKey)
	provider := &accountChain{canonicalChain: &canonicalChain{newChain(), bi(7)}, player: player, impl: addr(12), actionNonce: bi(3), authorizationNonce: 5}
	client, _ := NewClient(deployment(), provider)
	path, _ := EncodeV3ExactOutputPath([]common.Address{addr(4), addr(2)}, []uint32{3000})
	inner, e := client.BuildPayment(ctx, PaymentOptions{Core: coreCall(t), Mode: SponsoredNative, Payer: player, RefundRecipient: player, Relayer: addr(10), AmountInMaximum: bi(1000), Deadline: bi(1400), Route: Route{Channel: ChannelUniswapV3, Path: path}})
	if e != nil {
		t.Fatal(e)
	}
	action := AccountPayment{RouterCalldata: inner.Data, Value: inner.Value, Nonce: bi(3), Deadline: bi(1400), Relayer: addr(10)}
	codeHash := crypto.Keccak256Hash([]byte{0x60, 0})
	td, e := NativePaymentTypedData(bi(31337), player, addr(12), addr(3), codeHash, action)
	if e != nil {
		t.Fatal(e)
	}
	action.Signature, e = SignTypedData(key, td)
	if e != nil {
		t.Fatal(e)
	}
	auth, e := SignDelegation(key, bi(31337), addr(12), 5)
	if e != nil {
		t.Fatal(e)
	}
	policy := SponsoredPaymentOptions{Player: player, Implementation: addr(12), ImplementationCodeHash: codeHash, Payment: action, AuthorizationList: []types.SetCodeAuthorization{auth}}
	env, e := client.BuildSponsoredPayment(ctx, policy)
	if e != nil {
		t.Fatal(e)
	}
	gas, e := client.SimulateEnvelope(ctx, env, addr(10), policy)
	if e != nil || gas != 123456 {
		t.Fatal(gas, e)
	}
	if len(provider.call.AuthorizationList) != 1 || *provider.call.To != player || provider.call.Value.Sign() != 0 {
		t.Fatal("simulation omitted actual outer call or authorization list")
	}
	if _, e = client.SimulateEnvelope(ctx, env, addr(10)); e == nil {
		t.Fatal("account simulation accepted missing policy")
	}
	provider.undelegated = true
	if _, e = client.BuildSponsoredPayment(ctx, policy); e != nil {
		t.Fatal("persistent action nonce must survive delegation removal", e)
	}
	provider.authorizationNonce = 6
	if _, e = client.BuildSponsoredPayment(ctx, policy); e == nil {
		t.Fatal("stale delegation nonce accepted")
	}
	provider.authorizationNonce = 5
	provider.actionNonce = bi(4)
	if _, e = client.BuildSponsoredPayment(ctx, policy); e == nil {
		t.Fatal("stale action nonce accepted")
	}
	provider.actionNonce = bi(3)
	provider.wrongCode = true
	if _, e = client.BuildSponsoredPayment(ctx, policy); e == nil {
		t.Fatal("implementation substitution accepted")
	}
}

func TestCanonicalERC20Quotes(t *testing.T) {
	ctx := context.Background()
	provider := newChain()
	client, _ := NewClient(deployment(), provider)
	adapter, e := NewUniswapV3Adapter(provider, addr(9), []V3Route{{Tokens: []common.Address{addr(10), addr(2)}, Fees: []uint32{3000}}})
	if e != nil {
		t.Fatal(e)
	}
	params := parameters()
	params.TokenIn = addr(10)
	quote, route, e := client.QuoteRoute(ctx, adapter, params)
	if e != nil {
		t.Fatal(e)
	}
	if quote.MaxInput.Cmp(params.MaxInput) > 0 || ValidateV3Path(route.Path, addr(10), addr(2)) != nil {
		t.Fatal("ERC20 V3 endpoints not preserved")
	}
	inner, e := zeroABI.Pack("execute", slippage{Recipient: addr(3), BuyToken: addr(2), MinAmountOut: bi(10)}, [][]byte{{1}}, [32]byte{})
	if e != nil {
		t.Fatal(e)
	}
	data, e := zeroABI.Pack("exec", addr(6), addr(10), bi(100), addr(6), inner)
	if e != nil {
		t.Fatal(e)
	}
	body := map[string]any{"liquidityAvailable": true, "buyToken": addr(2).Hex(), "buyAmount": "10", "sellToken": addr(10).Hex(), "sellAmount": "60", "maxSellAmount": "100", "transaction": map[string]any{"to": addr(7).Hex(), "data": hexutil.Encode(data), "value": "0"}}
	zero, _ := NewZeroExAdapter(func(context.Context, SwapQuoteRequest) (map[string]any, error) { return body, nil })
	_, route, e = client.QuoteRoute(ctx, zero, params)
	if e != nil || route.Target != addr(7) {
		t.Fatal(route, e)
	}
	body["transaction"].(map[string]any)["value"] = "100"
	if _, _, e = client.QuoteRoute(ctx, zero, params); e == nil {
		t.Fatal("ERC20 quote accepted native principal")
	}
	body["transaction"].(map[string]any)["value"] = "0"
	body["transaction"].(map[string]any)["to"] = addr(6).Hex()
	if _, _, e = client.QuoteRoute(ctx, zero, params); e == nil {
		t.Fatal("direct Settler approval target accepted")
	}
}
