package rpc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
)

// TrustedProvider trusts the Protocol RPC service for chain facts. It performs
// no independent chain authentication and never connects to a chain endpoint.
// Each uncached response must match the explicitly configured deployment pins.
type TrustedProvider struct {
	client     *ProtocolRpcClient
	scope      DeploymentContext
	deployment ProtocolDeployment
}

func NewTrustedProvider(client *ProtocolRpcClient, scope DeploymentContext) (*TrustedProvider, error) {
	if client == nil {
		return nil, errors.New("Protocol RPC client required")
	}
	client.mu.Lock()
	pinned := client.pinned
	client.mu.Unlock()
	if !pinned {
		return nil, errors.New("trusted provider requires explicit deployment pins")
	}
	d, e := client.context(scope)
	if e != nil {
		return nil, e
	}
	return &TrustedProvider{client: client, scope: scope, deployment: d}, nil
}

type trustedRequest struct {
	Kind            string         `json:"kind"`
	Block           string         `json:"block,omitempty"`
	Address         common.Address `json:"address,omitempty"`
	From            common.Address `json:"from,omitempty"`
	Data            hexutil.Bytes  `json:"data,omitempty"`
	Gas             uint64         `json:"gas,omitempty"`
	TransactionHash common.Hash    `json:"transactionHash,omitempty"`
}
type trustedResponse struct {
	APIVersion      string             `json:"apiVersion"`
	ProtocolVersion string             `json:"protocolVersion"`
	Deployment      ProtocolDeployment `json:"deployment"`
	Header          *types.Header      `json:"header"`
	HeaderHash      common.Hash        `json:"headerHash"`
	Data            hexutil.Bytes      `json:"data"`
	Receipt         *types.Receipt     `json:"receipt"`
	Transaction     *types.Transaction `json:"transaction"`
	Pending         bool               `json:"pending"`
}

func (p *TrustedProvider) read(ctx context.Context, r trustedRequest) (trustedResponse, error) {
	var out trustedResponse
	e := p.client.call(ctx, p.scope.ChainID, "invisi_readChainState", []any{r}, &out)
	if e != nil {
		return out, e
	}
	want, _ := json.Marshal(p.deployment)
	got, _ := json.Marshal(out.Deployment)
	if out.APIVersion != "v1" || out.ProtocolVersion != "1" || !bytes.Equal(want, got) {
		return out, errors.New("trusted chain state deployment or version mismatch")
	}
	return out, nil
}
func trustedBlock(n *big.Int) (string, error) {
	if n == nil {
		return "latest", nil
	}
	if n.IsInt64() {
		switch n.Int64() {
		case -2:
			return "latest", nil
		case -3:
			return "finalized", nil
		case -4:
			return "safe", nil
		}
	}
	if n.Sign() < 0 || !n.IsUint64() || n.BitLen() > 63 {
		return "", errors.New("invalid block number")
	}
	return hexutil.EncodeBig(n), nil
}
func (p *TrustedProvider) ChainID(ctx context.Context) (*big.Int, error) {
	if _, e := p.HeaderByNumber(ctx, nil); e != nil {
		return nil, e
	}
	return new(big.Int).SetUint64(p.scope.ChainID), nil
}
func (p *TrustedProvider) header(ctx context.Context, b string) (*types.Header, error) {
	out, e := p.read(ctx, trustedRequest{Kind: "header", Block: b})
	if e != nil {
		return nil, e
	}
	h := out.Header
	if h == nil {
		return nil, ethereum.NotFound
	}
	if h.Number == nil || !h.Number.IsUint64() || h.Hash() != out.HeaderHash {
		return nil, errors.New("invalid trusted header")
	}
	return h, nil
}
func (p *TrustedProvider) HeaderByNumber(ctx context.Context, n *big.Int) (*types.Header, error) {
	b, e := trustedBlock(n)
	if e != nil {
		return nil, e
	}
	h, e := p.header(ctx, b)
	if e == nil && n != nil && n.Sign() >= 0 && h.Number.Cmp(n) != 0 {
		return nil, errors.New("trusted header number mismatch")
	}
	return h, e
}
func (p *TrustedProvider) HeaderByTag(ctx context.Context, t FinalityTag) (*types.Header, error) {
	switch t {
	case FinalityLatest, FinalitySafe, FinalityFinalized:
		return p.header(ctx, string(t))
	default:
		return nil, ErrFinalityUnsupported
	}
}
func (p *TrustedProvider) CodeAt(ctx context.Context, a common.Address, n *big.Int) ([]byte, error) {
	b, e := trustedBlock(n)
	if e != nil {
		return nil, e
	}
	out, e := p.read(ctx, trustedRequest{Kind: "code", Block: b, Address: a})
	return out.Data, e
}
func (p *TrustedProvider) CallContract(ctx context.Context, m ethereum.CallMsg, n *big.Int) ([]byte, error) {
	if m.To == nil || (m.Value != nil && m.Value.Sign() != 0) || m.GasPrice != nil || m.GasFeeCap != nil || m.GasTipCap != nil || len(m.AccessList) > 0 || len(m.BlobHashes) > 0 || m.BlobGasFeeCap != nil || len(m.AuthorizationList) > 0 {
		return nil, errors.New("unsupported trusted call fields")
	}
	b, e := trustedBlock(n)
	if e != nil {
		return nil, e
	}
	out, e := p.read(ctx, trustedRequest{Kind: "call", Block: b, Address: *m.To, From: m.From, Data: m.Data, Gas: m.Gas})
	return out.Data, e
}
func (p *TrustedProvider) TransactionReceipt(ctx context.Context, hash common.Hash) (*types.Receipt, error) {
	out, e := p.read(ctx, trustedRequest{Kind: "receipt", TransactionHash: hash})
	if e != nil {
		return nil, e
	}
	r := out.Receipt
	if r == nil {
		return nil, ethereum.NotFound
	}
	if r.TxHash != hash || r.BlockNumber == nil || !r.BlockNumber.IsUint64() || r.Status > 1 {
		return nil, errors.New("invalid trusted receipt")
	}
	h, e := p.HeaderByNumber(ctx, r.BlockNumber)
	if e != nil {
		return nil, e
	}
	if h.Hash() != r.BlockHash {
		return nil, errors.New("trusted receipt block mismatch")
	}
	return r, nil
}

func (p *TrustedProvider) TransactionByHash(ctx context.Context, hash common.Hash) (*types.Transaction, bool, error) {
	out, e := p.read(ctx, trustedRequest{Kind: "transaction", TransactionHash: hash})
	if e != nil {
		return nil, false, e
	}
	if out.Transaction == nil {
		return nil, false, ethereum.NotFound
	}
	if out.Transaction.Hash() != hash {
		return nil, false, errors.New("trusted transaction hash mismatch")
	}
	return out.Transaction, out.Pending, nil
}
