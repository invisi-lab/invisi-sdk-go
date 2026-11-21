package broadcaster

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
	"time"
)

type RouterEnvelope struct {
	To                common.Address               `json:"to"`
	Data              hexutil.Bytes                `json:"data"`
	Value             string                       `json:"value"`
	AuthorizationList []types.SetCodeAuthorization `json:"authorizationList,omitempty"`
}
type envelopeAuthorization struct {
	ChainId  *big.Int
	Delegate common.Address
	Nonce    uint64
	YParity  uint8
	R, S     [32]byte
}

var authTuple = func() abi.Type {
	t, e := abi.NewType("tuple[]", "", []abi.ArgumentMarshaling{{Name: "chainId", Type: "uint256"}, {Name: "delegate", Type: "address"}, {Name: "nonce", Type: "uint64"}, {Name: "yParity", Type: "uint8"}, {Name: "r", Type: "bytes32"}, {Name: "s", Type: "bytes32"}})
	if e != nil {
		panic(e)
	}
	return t
}()

func EnvelopeHash(chainID *big.Int, core common.Address, env RouterEnvelope) (common.Hash, error) {
	if e := routerUintCheck(chainID); e != nil {
		return common.Hash{}, e
	}
	if core == (common.Address{}) || env.To == (common.Address{}) || env.Value != "0" || len(env.Data) < 4 {
		return common.Hash{}, fmt.Errorf("invalid relay envelope")
	}
	auth := make([]envelopeAuthorization, len(env.AuthorizationList))
	for i, a := range env.AuthorizationList {
		auth[i] = envelopeAuthorization{a.ChainID.ToBig(), a.Address, a.Nonce, a.V, a.R.Bytes32(), a.S.Bytes32()}
	}
	authHash, e := routerHashArgs(abi.Arguments{{Type: authTuple}}, auth)
	if e != nil {
		return common.Hash{}, e
	}
	typeHash := crypto.Keccak256Hash([]byte("InvisiPaymentRouterEnvelope(uint256 chainId,address core,address to,uint256 value,bytes32 dataHash,bytes32 authorizationListHash)"))
	return routerHashArgs(abi.Arguments{routerKind("bytes32"), routerKind("uint256"), routerKind("address"), routerKind("address"), routerKind("uint256"), routerKind("bytes32"), routerKind("bytes32")}, typeHash, chainID, core, env.To, new(big.Int), crypto.Keccak256Hash(env.Data), authHash)
}

func routerUintCheck(n *big.Int) error {
	if n == nil || n.Sign() <= 0 || n.BitLen() > 256 {
		return fmt.Errorf("invalid chain ID")
	}
	return nil
}
func routerKind(name string) abi.Argument {
	t, e := abi.NewType(name, "", nil)
	if e != nil {
		panic(e)
	}
	return abi.Argument{Type: t}
}
func routerHashArgs(args abi.Arguments, v ...any) (common.Hash, error) {
	b, e := args.Pack(v...)
	if e != nil {
		return common.Hash{}, e
	}
	return crypto.Keccak256Hash(b), nil
}

const LegacyRouterRequestHashFormat = "payment-router-envelope-v1"
const RouterRequestHashFormat = "payment-router-intent-v1"

func VerifyRouterQuote(s SignedQuote, t Trust, data []byte, env RouterEnvelope, now uint64) error {
	if s.RequestHashFormat != RouterRequestHashFormat || s.Quote.Execution != "direct" || s.Quote.Schedule != nil {
		return fmt.Errorf("unsupported Router quote policy")
	}
	limit, e := relayDeadline(data, env)
	if e != nil {
		return e
	}
	deadline, e := decimal(s.Quote.Deadline)
	if e != nil {
		return e
	}
	if deadline.Cmp(limit) > 0 {
		return fmt.Errorf("Router quote outlives payment authorization")
	}
	h, e := RouterIntentHash(t.ChainID, t.Protocol, env)
	if e != nil {
		return e
	}
	return verifyQuoteWithHash(s, t, data, now, h)
}
func (c *Client) EstimateRouterQuote(ctx context.Context, data []byte, env RouterEnvelope) (Fee, error) {
	var fee Fee
	if _, e := relayDeadline(data, env); e != nil {
		return fee, e
	}
	if _, e := DecodeIntent(data); e != nil {
		return fee, e
	}
	if env.Value != "0" {
		return fee, fmt.Errorf("relay principal value must be zero")
	}
	e := c.call(ctx, "broadcaster_estimateQuote", map[string]any{"data": hexutil.Encode(data), "router": env}, &fee)
	if e != nil {
		return fee, e
	}
	amount, e := decimal(fee.FeeAmount)
	if e != nil {
		return fee, e
	}
	gas, e := decimal(fee.MaxGas)
	if e != nil {
		return fee, e
	}
	if fee.FeeRecipient != c.configuration.FeeRecipient || amount.Sign() <= 0 || amount.BitLen() > 128 || gas.Sign() <= 0 {
		return fee, fmt.Errorf("invalid Router fee")
	}
	return fee, nil
}
func (c *Client) GetRouterQuote(ctx context.Context, data []byte, env RouterEnvelope, trust Trust) (SignedQuote, error) {
	var s SignedQuote
	if _, e := RouterIntentHash(trust.ChainID, trust.Protocol, env); e != nil {
		return s, e
	}
	if e := c.call(ctx, "broadcaster_getQuote", map[string]any{"data": hexutil.Encode(data), "router": env}, &s); e != nil {
		return s, e
	}
	trust.Signer = c.configuration.QuoteSigner
	if e := VerifyRouterQuote(s, trust, data, env, uint64(time.Now().Unix())); e != nil {
		return s, e
	}
	if s.Quote.FeeRecipient != c.configuration.FeeRecipient {
		return s, fmt.Errorf("unregistered fee recipient")
	}
	trust.ChainID = new(big.Int).Set(trust.ChainID)
	s.Context = &trust
	return s, nil
}
func (c *Client) SubmitRouterTransaction(ctx context.Context, data []byte, env RouterEnvelope, s SignedQuote) (Status, error) {
	var out Status
	if s.Context == nil {
		return out, fmt.Errorf("missing verified quote context")
	}
	if e := ValidateRouterSubmission(data, env); e != nil {
		return out, e
	}
	trust := *s.Context
	trust.Signer = c.configuration.QuoteSigner
	if e := VerifyRouterQuote(s, trust, data, env, uint64(time.Now().Unix())); e != nil {
		return out, e
	}
	if s.Quote.FeeRecipient != c.configuration.FeeRecipient {
		return out, fmt.Errorf("unregistered fee recipient")
	}
	e := c.call(ctx, "broadcaster_submitTransaction", map[string]any{"data": hexutil.Encode(data), "router": env, "quote": s.Quote, "signature": s.Signature}, &out)
	if e != nil {
		return out, e
	}
	return out, checkStatus(out, "")
}
