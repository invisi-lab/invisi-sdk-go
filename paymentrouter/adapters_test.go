package paymentrouter

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"io"
	"net/http"
	"strings"
	"testing"
)

func request() SwapQuoteRequest {
	return SwapQuoteRequest{QuoteParameters: parameters(), Deployment: deployment(), Now: bi(1000), BlockNumber: bi(10), SettlerAddress: addr(6)}
}
func apiQuote(t *testing.T) map[string]any {
	return map[string]any{"liquidityAvailable": true, "buyToken": addr(2).Hex(), "buyAmount": "10", "sellToken": nativeToken.Hex(), "sellAmount": "60", "maxSellAmount": "100", "transaction": map[string]any{"to": addr(6).Hex(), "data": hexutil.Encode(nativeData(t, addr(3), addr(2), 10, 100, 1100, 1)), "value": "100"}}
}
func TestZeroExMetadataAndInputPinning(t *testing.T) {
	for _, change := range []string{"buyAmount", "sellToken", "liquidity", "value", "taker", "chainId", "sellAmount", "decimal", "budget"} {
		t.Run(change, func(t *testing.T) {
			body := apiQuote(t)
			switch change {
			case "buyAmount":
				body[change] = "11"
			case "sellToken":
				body[change] = addr(4).Hex()
			case "liquidity":
				body["liquidityAvailable"] = false
			case "value":
				body["transaction"].(map[string]any)[change] = "99"
			case "taker":
				body[change] = addr(4).Hex()
			case "chainId":
				body[change] = "1"
			case "sellAmount":
				body[change] = "101"
			case "decimal":
				body["buyAmount"] = "010"
			case "budget":
				body["maxSellAmount"] = "1000"
			}
			a, _ := NewZeroExAdapter(func(context.Context, SwapQuoteRequest) (map[string]any, error) { return body, nil })
			if _, e := a.Quote(context.Background(), request()); e == nil {
				t.Fatal("accepted invalid metadata")
			}
		})
	}
	a, _ := NewZeroExAdapter(func(_ context.Context, r SwapQuoteRequest) (map[string]any, error) {
		r.MaxInput.SetInt64(1000)
		r.Deployment.ChainID.SetInt64(1)
		b := apiQuote(t)
		b["maxSellAmount"] = "1000"
		return b, nil
	})
	r := request()
	if _, e := a.Quote(context.Background(), r); e == nil {
		t.Fatal("transport rewrote budget")
	}
	if r.MaxInput.Int64() != 101 || r.Deployment.ChainID.Int64() != 31337 {
		t.Fatal("input aliases")
	}
	called := false
	a, _ = NewZeroExAdapter(func(context.Context, SwapQuoteRequest) (map[string]any, error) {
		called = true
		return apiQuote(t), nil
	})
	for _, invalid := range []SwapQuoteRequest{func() SwapQuoteRequest { r := request(); r.Deadline = bi(-1); return r }(), func() SwapQuoteRequest { r := request(); r.Deadline = number("18446744073709551616"); return r }(), func() SwapQuoteRequest { r := request(); r.SlippageBps = intPtr(10001); return r }()} {
		if _, e := a.Quote(context.Background(), invalid); e == nil {
			t.Fatal("accepted invalid standalone request")
		}
	}
	if called {
		t.Fatal("transport used before validation")
	}
}

type transportFunc func(*http.Request) (*http.Response, error)

func (f transportFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestHTTPFetcher(t *testing.T) {
	for _, endpoint := range []string{"http://example.com", "https://x:y@example.com", "https://example.com?a=b", "https://example.com#x"} {
		if _, e := NewZeroExQuoteFetcher(ZeroExHTTPOptions{APIKey: "secret", Endpoint: endpoint}); e == nil {
			t.Fatal("accepted endpoint", endpoint)
		}
	}
	key := "secret"
	client := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		if r.Header.Get("0x-api-key") != key || r.Header.Get("0x-version") != "v2" || r.URL.Query().Get("buyAmount") != "10" || r.URL.Query().Has("sellAmount") || r.URL.Query().Get("taker") != addr(3).Hex() {
			t.Fatal("wrong HTTP request")
		}
		return &http.Response{StatusCode: 401, Body: io.NopCloser(strings.NewReader(key)), Header: make(http.Header)}, nil
	})}
	f, e := NewZeroExQuoteFetcher(ZeroExHTTPOptions{APIKey: key, Client: client})
	if e != nil {
		t.Fatal(e)
	}
	if _, e = f(context.Background(), request()); e == nil || !strings.Contains(e.Error(), "401") || strings.Contains(e.Error(), key) {
		t.Fatal("unsanitized HTTP error", e)
	}
	client.Transport = transportFunc(func(*http.Request) (*http.Response, error) { return nil, fmt.Errorf("leaked %s", key) })
	f, _ = NewZeroExQuoteFetcher(ZeroExHTTPOptions{APIKey: key, Client: client})
	if _, e = f(context.Background(), request()); e == nil || strings.Contains(e.Error(), key) {
		t.Fatal("transport credential leak")
	}
	client.Transport = transportFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader("[]")), Header: make(http.Header)}, nil
	})
	f, _ = NewZeroExQuoteFetcher(ZeroExHTTPOptions{APIKey: key, Client: client})
	if _, e = f(context.Background(), request()); e == nil {
		t.Fatal("accepted invalid JSON")
	}
	client.Transport = transportFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 302, Header: http.Header{"Location": []string{"https://other.example/"}}, Body: io.NopCloser(strings.NewReader(""))}, nil
	})
	f, _ = NewZeroExQuoteFetcher(ZeroExHTTPOptions{APIKey: key, Client: client})
	if _, e = f(context.Background(), request()); e == nil {
		t.Fatal("followed redirect")
	}
}
func TestQuoteFreshnessCancellationAndMutation(t *testing.T) {
	for _, change := range []string{"time", "registry", "cancel", "mutation"} {
		t.Run(change, func(t *testing.T) {
			rpc := newChain()
			c, _ := NewClient(deployment(), rpc)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			a, _ := NewZeroExAdapter(func(_ context.Context, r SwapQuoteRequest) (map[string]any, error) {
				switch change {
				case "time":
					rpc.now = 1101
				case "registry":
					rpc.settler = addr(5)
				case "cancel":
					cancel()
				case "mutation":
					r.MaxInput.SetInt64(1000)
					b := apiQuote(t)
					b["maxSellAmount"] = "1000"
					return b, nil
				}
				return apiQuote(t), nil
			})
			if _, e := c.Quote(ctx, a, parameters()); e == nil {
				t.Fatal("accepted stale or changed quote")
			}
		})
	}
	rpc := newChain()
	a, _ := NewUniswapV3Adapter(rpc, addr(9), []V3Route{{Tokens: []common.Address{addr(4), addr(2)}, Fees: []uint32{3000}}})
	r := request()
	r.MaxInput = bi(60)
	rpc.hook = func() { r.MaxInput.SetInt64(1000) }
	if _, e := a.Quote(context.Background(), r); e == nil {
		t.Fatal("V3 caller mutated authorized budget")
	}
}

func TestSuccessfulNativeAdapterAndDefaultSlippage(t *testing.T) {
	ctx := context.Background()
	rpc := newChain()
	c, _ := NewClient(deployment(), rpc)
	zero, _ := NewZeroExAdapter(func(context.Context, SwapQuoteRequest) (map[string]any, error) { return apiQuote(t), nil })
	q, e := c.Quote(ctx, zero, parameters())
	if e != nil || q.MaxInput.Cmp(bi(100)) != 0 || q.EstimatedInput.Cmp(bi(60)) != 0 || q.Deadline.Cmp(bi(1100)) != 0 {
		t.Fatal("0x quote", q, e)
	}
	if _, e = c.BuildPaymentTransaction(ctx, coreCall(t), q, addr(10)); e != nil {
		t.Fatal(e)
	}
	v3, _ := NewUniswapV3Adapter(rpc, addr(9), []V3Route{{Tokens: []common.Address{addr(4), addr(2)}, Fees: []uint32{3000}}})
	p := parameters()
	p.SlippageBps = nil
	q, e = c.Quote(ctx, v3, p)
	if e != nil || q.MaxInput.Cmp(bi(61)) != 0 {
		t.Fatal("default slippage", q, e)
	}
	p.SlippageBps = intPtr(0)
	q, e = c.Quote(ctx, v3, p)
	if e != nil || q.MaxInput.Cmp(bi(60)) != 0 {
		t.Fatal("explicit zero slippage", q, e)
	}
}
