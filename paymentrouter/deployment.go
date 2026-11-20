package paymentrouter

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
)

type deploymentFrame struct {
	Now, BlockNumber *big.Int
	Hash             common.Hash
	SettlerAddress   common.Address
}

func normalizeDeployment(d Deployment) (Deployment, error) {
	d = cloneDeployment(d)
	if e := uintCheck(d.ChainID, 64, true, "chain ID"); e != nil {
		return d, e
	}
	addresses := []common.Address{d.RouterAddress, d.CoreAddress, d.WETHAddress, d.UniswapRouterAddress}
	if d.ZeroEx != nil {
		addresses = append(addresses, d.ZeroEx.AllowanceHolderAddress, d.ZeroEx.SettlerRegistryAddress)
	}
	seen := map[common.Address]bool{}
	for _, a := range addresses {
		if !nonzero(a) || seen[a] {
			return d, fmt.Errorf("deployment addresses must be nonzero and distinct")
		}
		seen[a] = true
	}
	return d, nil
}
func assertFrame(ctx context.Context, p Provider, f deploymentFrame) error {
	h, e := p.HeaderByNumber(ctx, copyInt(f.BlockNumber))
	if e != nil {
		return e
	}
	if h == nil || h.Hash() != f.Hash {
		return fmt.Errorf("canonical Router block changed")
	}
	return ctx.Err()
}
func frame(ctx context.Context, p Provider, d Deployment, zero bool) (deploymentFrame, error) {
	var f deploymentFrame
	id, e := p.ChainID(ctx)
	if e != nil {
		return f, e
	}
	if id == nil || id.Cmp(d.ChainID) != 0 {
		return f, fmt.Errorf("payment Router chain mismatch")
	}
	h, e := p.HeaderByNumber(ctx, nil)
	if e != nil {
		return f, e
	}
	if h == nil || uintCheck(h.Number, 53, false, "canonical block") != nil || h.Time > 9007199254740991 {
		return f, fmt.Errorf("missing canonical chain block")
	}
	f = deploymentFrame{Now: new(big.Int).SetUint64(h.Time), BlockNumber: copyInt(h.Number), Hash: h.Hash()}
	code := func(a common.Address) ([]byte, error) {
		b, e := p.CodeAt(ctx, a, copyInt(f.BlockNumber))
		if e == nil && len(b) == 0 {
			e = fmt.Errorf("missing contract code: %s", a)
		}
		return b, e
	}
	runtime, e := code(d.RouterAddress)
	if e != nil {
		return f, e
	}
	if d.RuntimeCodeHash != nil && crypto.Keccak256Hash(runtime) != *d.RuntimeCodeHash {
		return f, fmt.Errorf("Router runtime code hash mismatch")
	}
	call := func(to common.Address, name string, args ...any) (any, error) {
		data, e := gettersABI.Pack(name, args...)
		if e != nil {
			return nil, e
		}
		out, e := p.CallContract(ctx, ethereum.CallMsg{To: &to, Data: data}, copyInt(f.BlockNumber))
		if e != nil {
			return nil, e
		}
		v, e := gettersABI.Methods[name].Outputs.Unpack(out)
		if e != nil {
			return nil, e
		}
		return v[0], nil
	}
	pinned := map[string]common.Address{"core": d.CoreAddress, "dex": d.UniswapRouterAddress, "weth": d.WETHAddress}
	if d.ZeroEx != nil {
		pinned["allowanceHolder"] = d.ZeroEx.AllowanceHolderAddress
		pinned["settlerRegistry"] = d.ZeroEx.SettlerRegistryAddress
	}
	for name, expected := range pinned {
		v, e := call(d.RouterAddress, name)
		if e != nil {
			return f, e
		}
		if v.(common.Address) != expected {
			return f, fmt.Errorf("Router %s deployment mismatch", name)
		}
		if _, e = code(expected); e != nil {
			return f, e
		}
	}
	v, e := call(d.RouterAddress, "CHAIN_ID")
	if e != nil {
		return f, e
	}
	if v.(*big.Int).Cmp(d.ChainID) != 0 {
		return f, fmt.Errorf("Router deployment chain mismatch")
	}
	if zero {
		if d.ZeroEx == nil {
			return f, fmt.Errorf("0x is not enabled for this deployment")
		}
		v, e = call(d.ZeroEx.SettlerRegistryAddress, "ownerOf", big.NewInt(2))
		if e != nil {
			return f, e
		}
		f.SettlerAddress = v.(common.Address)
		if !nonzero(f.SettlerAddress) || f.SettlerAddress == d.RouterAddress || f.SettlerAddress == d.ZeroEx.AllowanceHolderAddress {
			return f, fmt.Errorf("invalid current Settler")
		}
		if _, e = code(f.SettlerAddress); e != nil {
			return f, e
		}
	}
	return f, assertFrame(ctx, p, f)
}
