package paymentrouter

import (
	"context"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/invisi-labs/invisi-sdk-go/sdk"
	"math/big"
)

type ZeroExDeployment struct{ AllowanceHolderAddress, SettlerRegistryAddress common.Address }
type Deployment struct {
	ChainID                                                       *big.Int
	RouterAddress, CoreAddress, WETHAddress, UniswapRouterAddress common.Address
	RuntimeCodeHash                                               *common.Hash
	ZeroEx                                                        *ZeroExDeployment
}

// Provider deliberately exposes no transaction sending or signing capabilities.
// Every code read, call and gas estimate must use the supplied block number.
type Provider interface {
	ChainID(context.Context) (*big.Int, error)
	HeaderByNumber(context.Context, *big.Int) (*types.Header, error)
	CodeAt(context.Context, common.Address, *big.Int) ([]byte, error)
	CallContract(context.Context, ethereum.CallMsg, *big.Int) ([]byte, error)
	EstimateGasAt(context.Context, ethereum.CallMsg, *big.Int) (uint64, error)
}
type Channel string

const (
	UniswapV3 Channel = "uniswap-v3"
	ZeroEx    Channel = "zeroex"
)

// SlippageBps defaults to 50 when nil; a pointer to zero disables slippage.
type QuoteParameters struct {
	// TokenIn is zero for native input. Canonical quote flows also support ERC20.
	TokenIn                       common.Address
	TokenOut                      common.Address
	AmountOut, MaxInput, Deadline *big.Int
	SlippageBps                   *int
}
type SwapQuoteRequest struct {
	QuoteParameters
	Deployment       Deployment
	Now, BlockNumber *big.Int
	SettlerAddress   common.Address
}
type SwapQuote struct {
	Channel                                       Channel
	ChainID                                       *big.Int
	RouterAddress, TokenOut                       common.Address
	AmountOut, EstimatedInput, MaxInput, Deadline *big.Int
	ReversePath                                   []byte
	Target                                        common.Address
	Data                                          []byte
	SettlerAddress                                common.Address
}
type QuoteAdapter interface {
	Channel() Channel
	Quote(context.Context, SwapQuoteRequest) (SwapQuote, error)
}
type RouterSimulation struct {
	GasEstimate           uint64
	ETHSpent, ETHRefunded *big.Int
}

// PreparePaymentOptions contains no private-input selection or wallet signing options.
type PreparePaymentOptions struct {
	Invisi       *sdk.Invisi
	Wallet       *sdk.Wallet
	Protocol     *sdk.Protocol
	TokenAddress common.Address
	Outputs      []sdk.NoteOutput
	Fee          *sdk.Fee
	Deadline     uint64
}

func cloneDeployment(d Deployment) Deployment {
	d.ChainID = copyInt(d.ChainID)
	if d.ZeroEx != nil {
		x := *d.ZeroEx
		d.ZeroEx = &x
	}
	if d.RuntimeCodeHash != nil {
		x := *d.RuntimeCodeHash
		d.RuntimeCodeHash = &x
	}
	return d
}
func slippageBps(p QuoteParameters) int {
	if p.SlippageBps == nil {
		return 50
	}
	return *p.SlippageBps
}
func cloneParameters(p QuoteParameters) QuoteParameters {
	if p.SlippageBps != nil {
		v := *p.SlippageBps
		p.SlippageBps = &v
	}
	p.AmountOut = copyInt(p.AmountOut)
	p.MaxInput = copyInt(p.MaxInput)
	p.Deadline = copyInt(p.Deadline)
	return p
}
func cloneRequest(r SwapQuoteRequest) SwapQuoteRequest {
	r.QuoteParameters = cloneParameters(r.QuoteParameters)
	r.Deployment = cloneDeployment(r.Deployment)
	r.Now = copyInt(r.Now)
	r.BlockNumber = copyInt(r.BlockNumber)
	return r
}
func cloneQuote(q SwapQuote) SwapQuote {
	q.ChainID = copyInt(q.ChainID)
	q.AmountOut = copyInt(q.AmountOut)
	q.EstimatedInput = copyInt(q.EstimatedInput)
	q.MaxInput = copyInt(q.MaxInput)
	q.Deadline = copyInt(q.Deadline)
	q.Data = append([]byte(nil), q.Data...)
	q.ReversePath = append([]byte(nil), q.ReversePath...)
	return q
}
