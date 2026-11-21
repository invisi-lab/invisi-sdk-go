package paymentrouter

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

type V3Route struct {
	Tokens []common.Address
	Fees   []uint32
}
type UniswapV3Adapter struct {
	provider Provider
	quoter   common.Address
	routes   []V3Route
}

func NewUniswapV3Adapter(p Provider, quoter common.Address, routes []V3Route) (*UniswapV3Adapter, error) {
	if p == nil || !nonzero(quoter) || len(routes) == 0 {
		return nil, fmt.Errorf("provider, quoter and at least one V3 route required")
	}
	saved := make([]V3Route, len(routes))
	for i, r := range routes {
		if _, e := EncodeV3ExactOutputPath(r.Tokens, r.Fees); e != nil {
			return nil, e
		}
		saved[i] = V3Route{append([]common.Address(nil), r.Tokens...), append([]uint32(nil), r.Fees...)}
	}
	return &UniswapV3Adapter{p, quoter, saved}, nil
}
func (*UniswapV3Adapter) Channel() Channel { return UniswapV3 }
func (a *UniswapV3Adapter) Quote(ctx context.Context, input SwapQuoteRequest) (SwapQuote, error) {
	r := cloneRequest(input)
	var best SwapQuote
	if e := ctx.Err(); e != nil {
		return best, e
	}
	if _, e := normalizeDeployment(r.Deployment); e != nil {
		return best, e
	}
	if e := uintCheck(r.Now, 64, false, "chain timestamp"); e != nil {
		return best, e
	}
	id, e := a.provider.ChainID(ctx)
	if e != nil {
		return best, e
	}
	if id == nil || r.Deployment.ChainID == nil || id.Cmp(r.Deployment.ChainID) != 0 {
		return best, fmt.Errorf("V3 quoter chain mismatch")
	}
	if e = validateParameters(r.QuoteParameters, r.Now); e != nil {
		return best, e
	}
	if e = uintCheck(r.BlockNumber, 53, false, "block number"); e != nil {
		return best, e
	}
	code, e := a.provider.CodeAt(ctx, a.quoter, copyInt(r.BlockNumber))
	if e != nil {
		return best, e
	}
	if len(code) == 0 {
		return best, fmt.Errorf("missing QuoterV2 code")
	}
	for _, route := range a.routes {
		if route.Tokens[0] != quoteInput(r) || route.Tokens[len(route.Tokens)-1] != r.TokenOut {
			continue
		}
		path, _ := EncodeV3ExactOutputPath(route.Tokens, route.Fees)
		data, e := quoterABI.Pack("quoteExactOutput", path, r.AmountOut)
		if e != nil {
			continue
		}
		raw, e := a.provider.CallContract(ctx, ethereum.CallMsg{To: &a.quoter, Data: data}, copyInt(r.BlockNumber))
		if e != nil {
			continue
		}
		v, e := quoterABI.Methods["quoteExactOutput"].Outputs.Unpack(raw)
		if e != nil {
			continue
		}
		amount := v[0].(*big.Int)
		if uintCheck(amount, 256, true, "quoted input") != nil {
			continue
		}
		max := new(big.Int).Mul(amount, big.NewInt(int64(10000+slippageBps(r.QuoteParameters))))
		max.Add(max, big.NewInt(9999)).Div(max, big.NewInt(10000))
		if max.Cmp(r.MaxInput) > 0 {
			continue
		}
		if best.EstimatedInput == nil || amount.Cmp(best.EstimatedInput) < 0 {
			best = SwapQuote{Channel: UniswapV3, ChainID: copyInt(r.Deployment.ChainID), RouterAddress: r.Deployment.RouterAddress, TokenOut: r.TokenOut, AmountOut: copyInt(r.AmountOut), EstimatedInput: copyInt(amount), MaxInput: max, Deadline: copyInt(r.Deadline), ReversePath: path}
		}
	}
	if e = ctx.Err(); e != nil {
		return SwapQuote{}, e
	}
	if best.EstimatedInput == nil {
		return best, fmt.Errorf("No executable V3 route within the ETH budget")
	}
	return best, nil
}

type ZeroExQuoteTransport func(context.Context, SwapQuoteRequest) (map[string]any, error)
type ZeroExAdapter struct{ transport ZeroExQuoteTransport }

func NewZeroExAdapter(t ZeroExQuoteTransport) (*ZeroExAdapter, error) {
	if t == nil {
		return nil, fmt.Errorf("0x transport required")
	}
	return &ZeroExAdapter{t}, nil
}
func (*ZeroExAdapter) Channel() Channel { return ZeroEx }

var decimalPattern = regexp.MustCompile(`^(0|[1-9][0-9]*)$`)

func decimal(v any, label string) (*big.Int, error) {
	s, ok := v.(string)
	if !ok || !decimalPattern.MatchString(s) {
		return nil, fmt.Errorf("invalid %s", label)
	}
	n, _ := new(big.Int).SetString(s, 10)
	return n, uintCheck(n, 256, false, label)
}
func parseAddress(v any) (common.Address, error) {
	s, ok := v.(string)
	if !ok || !common.IsHexAddress(s) {
		return common.Address{}, fmt.Errorf("invalid address")
	}
	a := common.HexToAddress(s)
	unprefixed := strings.TrimPrefix(s, "0x")
	if unprefixed != strings.ToLower(unprefixed) && unprefixed != strings.ToUpper(unprefixed) && s != a.Hex() {
		return common.Address{}, fmt.Errorf("invalid address checksum")
	}
	if !nonzero(a) {
		return a, fmt.Errorf("address must be nonzero")
	}
	return a, nil
}
func addressEqual(v any, expected common.Address) bool {
	a, e := parseAddress(v)
	return e == nil && a == expected
}
func (a *ZeroExAdapter) Quote(ctx context.Context, input SwapQuoteRequest) (SwapQuote, error) {
	r := cloneRequest(input)
	var q SwapQuote
	fail := func(s string) (SwapQuote, error) { return q, fmt.Errorf("%s", s) }
	if e := ctx.Err(); e != nil {
		return q, e
	}
	if _, e := normalizeDeployment(r.Deployment); e != nil {
		return q, e
	}
	if e := uintCheck(r.Now, 64, false, "chain timestamp"); e != nil {
		return q, e
	}
	if r.Deployment.ZeroEx == nil || !nonzero(r.SettlerAddress) {
		return fail("0x deployment and current Settler required")
	}
	if e := validateParameters(r.QuoteParameters, r.Now); e != nil {
		return q, e
	}
	body, e := a.transport(ctx, cloneRequest(r))
	if e != nil {
		return q, e
	}
	if e = ctx.Err(); e != nil {
		return q, e
	}
	if body["liquidityAvailable"] != true {
		return fail("0x liquidity is unavailable")
	}
	if !addressEqual(body["buyToken"], r.TokenOut) {
		return fail("0x buy token mismatch")
	}
	sellToken := r.TokenIn
	if !nonzero(sellToken) {
		sellToken = nativeToken
	}
	if !addressEqual(body["sellToken"], sellToken) {
		return fail("0x sell token must be native ETH")
	}
	amount, e := decimal(body["buyAmount"], "buyAmount")
	if e != nil {
		return q, e
	}
	if amount.Cmp(r.AmountOut) != 0 {
		return fail("0x buy amount mismatch")
	}
	max, e := decimal(body["maxSellAmount"], "maxSellAmount")
	if e != nil {
		return q, e
	}
	if max.Sign() == 0 || max.Cmp(r.MaxInput) > 0 {
		return fail("0x quote exceeds ETH budget")
	}
	tx, ok := body["transaction"].(map[string]any)
	if !ok {
		return fail("invalid 0x transaction")
	}
	value, e := decimal(tx["value"], "transaction value")
	if e != nil {
		return q, e
	}
	if (!nonzero(r.TokenIn) && value.Cmp(max) != 0) || (nonzero(r.TokenIn) && value.Sign() != 0) {
		return fail("0x native transaction value mismatch")
	}
	for _, field := range []string{"taker", "recipient"} {
		if v, ok := body[field]; ok && !addressEqual(v, r.Deployment.RouterAddress) {
			return fail("0x " + field + " mismatch")
		}
	}
	for _, m := range []map[string]any{body, tx} {
		if v, ok := m["chainId"]; ok {
			var id *big.Int
			switch n := v.(type) {
			case float64:
				if n < 0 || n > 9007199254740991 || n != float64(uint64(n)) {
					return fail("0x quote chain mismatch")
				}
				id = new(big.Int).SetUint64(uint64(n))
			case json.Number:
				f, err := strconv.ParseFloat(string(n), 64)
				if err != nil || f < 0 || f > 9007199254740991 || f != float64(uint64(f)) {
					return fail("0x quote chain mismatch")
				}
				id = new(big.Int).SetUint64(uint64(f))
			default:
				id, e = decimal(v, "chainId")
				if e != nil {
					return q, e
				}
			}
			if id.Cmp(r.Deployment.ChainID) != 0 {
				return fail("0x quote chain mismatch")
			}
		}
	}
	target, e := parseAddress(tx["to"])
	if e != nil {
		return q, e
	}
	s, ok := tx["data"].(string)
	if !ok {
		return fail("invalid transaction data")
	}
	data, e := hexutil.Decode(s)
	if e != nil {
		return fail("invalid transaction data")
	}
	options := ZeroExPayloadOptions{Target: target, Data: data, RouterAddress: r.Deployment.RouterAddress, TokenOut: r.TokenOut, AmountOut: r.AmountOut, MaxInput: max, SettlerAddress: r.SettlerAddress, AllowanceHolderAddress: r.Deployment.ZeroEx.AllowanceHolderAddress, Now: r.Now}
	var checked CheckedZeroExPayload
	if nonzero(r.TokenIn) {
		e = ValidateZeroExTokenPayload(r.TokenIn, options)
		checked = CheckedZeroExPayload{Target: target, Data: data, SettlerAddress: r.SettlerAddress, NativeDeadline: r.Deadline}
	} else {
		checked, e = ValidateZeroExPayload(options)
	}
	if e != nil {
		return q, e
	}
	estimate := copyInt(max)
	if v, ok := body["sellAmount"]; ok {
		estimate, e = decimal(v, "sellAmount")
		if e != nil {
			return q, e
		}
	}
	if estimate.Cmp(max) > 0 {
		return fail("0x estimated input exceeds budget")
	}
	deadline := copyInt(r.Deadline)
	if checked.NativeDeadline.Cmp(deadline) < 0 {
		deadline = copyInt(checked.NativeDeadline)
	}
	return SwapQuote{Channel: ZeroEx, ChainID: copyInt(r.Deployment.ChainID), RouterAddress: r.Deployment.RouterAddress, TokenOut: r.TokenOut, AmountOut: copyInt(r.AmountOut), EstimatedInput: estimate, MaxInput: max, Deadline: deadline, Target: checked.Target, Data: checked.Data, SettlerAddress: checked.SettlerAddress}, nil
}

// ZeroExHTTPOptions requires an explicit credential; no environment variables are read.
// Client is copied and redirect handling is always replaced with a strict rejection.
type ZeroExHTTPOptions struct {
	APIKey, Endpoint string
	Client           *http.Client
}

func NewZeroExQuoteFetcher(o ZeroExHTTPOptions) (ZeroExQuoteTransport, error) {
	if strings.TrimSpace(o.APIKey) == "" {
		return nil, fmt.Errorf("0x API key required")
	}
	endpoint := o.Endpoint
	if endpoint == "" {
		endpoint = "https://api.0x.org/swap/allowance-holder/quote"
	}
	u, e := url.Parse(endpoint)
	if e != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" {
		return nil, fmt.Errorf("0x endpoint must be an HTTPS URL without credentials or query parameters")
	}
	client := *http.DefaultClient
	if o.Client != nil {
		client = *o.Client
	}
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return fmt.Errorf("redirect rejected") }
	key := o.APIKey
	return func(ctx context.Context, input SwapQuoteRequest) (map[string]any, error) {
		r := cloneRequest(input)
		if e := ctx.Err(); e != nil {
			return nil, e
		}
		if e := validateParameters(r.QuoteParameters, r.Now); e != nil {
			return nil, e
		}
		v := *u
		sellToken := r.TokenIn
		if !nonzero(sellToken) {
			sellToken = nativeToken
		}
		values := url.Values{"chainId": {r.Deployment.ChainID.String()}, "sellToken": {sellToken.Hex()}, "buyToken": {r.TokenOut.Hex()}, "buyAmount": {r.AmountOut.String()}, "taker": {r.Deployment.RouterAddress.Hex()}, "recipient": {r.Deployment.RouterAddress.Hex()}, "slippageBps": {strconv.Itoa(slippageBps(r.QuoteParameters))}}
		v.RawQuery = values.Encode()
		req, e := http.NewRequestWithContext(ctx, http.MethodGet, v.String(), nil)
		if e != nil {
			return nil, fmt.Errorf("0x quote request failed")
		}
		req.Header.Set("0x-api-key", key)
		req.Header.Set("0x-version", "v2")
		resp, e := client.Do(req)
		if e != nil {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			return nil, fmt.Errorf("0x quote request failed")
		}
		defer resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil, fmt.Errorf("0x quote HTTP %d", resp.StatusCode)
		}
		decoder := json.NewDecoder(resp.Body)
		decoder.UseNumber()
		var body map[string]any
		if decoder.Decode(&body) != nil || body == nil {
			return nil, fmt.Errorf("Invalid 0x quote JSON response")
		}
		var extra any
		if decoder.Decode(&extra) != io.EOF {
			return nil, fmt.Errorf("Invalid 0x quote JSON response")
		}
		return body, nil
	}, nil
}

func quoteInput(r SwapQuoteRequest) common.Address {
	if nonzero(r.TokenIn) {
		return r.TokenIn
	}
	return r.Deployment.WETHAddress
}
