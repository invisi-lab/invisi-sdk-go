package protocol

import (
	"context"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
)

// Protocol describes one multi-token deployment. It owns no connections or keys.
type Protocol struct {
	deployment protocolrpc.ProtocolDeployment
}

func NewProtocol(deployment protocolrpc.ProtocolDeployment) (*Protocol, error) {
	if e := deployment.Validate(); e != nil {
		return nil, e
	}
	if deployment.Registry != nil {
		r := *deployment.Registry
		deployment.Registry = &r
	}
	return &Protocol{deployment}, nil
}
func ProtocolFromRPC(ctx context.Context, rpc *protocolrpc.ProtocolRpcClient, provider protocolrpc.ChainProvider) (*Protocol, error) {
	if rpc == nil || provider == nil {
		return nil, errors.New("Protocol RPC and Chain RPC provider required")
	}
	chain, e := provider.ChainID(ctx)
	if e != nil {
		return nil, e
	}
	if chain.Sign() <= 0 || !chain.IsUint64() {
		return nil, errors.New("invalid chainId")
	}
	d, e := rpc.GetDeployment(ctx, chain.Uint64())
	if e != nil {
		return nil, e
	}
	p, e := NewProtocol(d)
	if e != nil {
		return nil, e
	}
	if e = p.ValidateDeployment(ctx, provider); e != nil {
		return nil, e
	}
	return p, nil
}
func (p *Protocol) Deployment() protocolrpc.ProtocolDeployment {
	d := p.deployment
	if d.Registry != nil {
		r := *d.Registry
		d.Registry = &r
	}
	return d
}
func (p *Protocol) Context() protocolrpc.DeploymentContext {
	n, _ := p.deployment.ChainNumber()
	return protocolrpc.DeploymentContext{ChainID: n.Uint64(), ProtocolAddress: p.deployment.Protocol.Address}
}
func (p *Protocol) ValidateDeployment(ctx context.Context, provider protocolrpc.ChainProvider) error {
	if provider == nil {
		return errors.New("Chain RPC provider required")
	}
	chain, e := provider.ChainID(ctx)
	if e != nil {
		return e
	}
	expected, _ := p.deployment.ChainNumber()
	if chain.Cmp(expected) != 0 {
		return errors.New("deployment chain mismatch")
	}
	code, e := provider.CodeAt(ctx, p.deployment.Protocol.Address, nil)
	if e != nil {
		return e
	}
	if len(code) == 0 || crypto.Keccak256Hash(code) != p.deployment.Protocol.RuntimeCodeHash {
		return errors.New("protocol runtime code mismatch")
	}
	data := crypto.Keccak256([]byte("VERSION()"))[:4]
	out, e := provider.CallContract(ctx, ethereum.CallMsg{To: &p.deployment.Protocol.Address, Data: data}, nil)
	if e != nil {
		return e
	}
	if len(out) != 32 || new(big.Int).SetBytes(out).Cmp(big.NewInt(1)) != 0 {
		return errors.New("unsupported Protocol VERSION")
	}
	return nil
}
func (p *Protocol) ForToken(token common.Address) (TransactionContext, error) {
	if token == (common.Address{}) {
		return TransactionContext{}, errors.New("zero token")
	}
	chain, _ := p.deployment.ChainNumber()
	return TransactionContext{ChainID: chain, ProtocolAddress: p.deployment.Protocol.Address, TokenAddress: token}, nil
}
