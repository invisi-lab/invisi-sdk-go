package transaction

import (
	"context"
	"math/big"
	"sync/atomic"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

type readTestProvider struct {
	Provider
	calls atomic.Int32
	codes atomic.Int32
	fork  atomic.Uint64
}

func (p *readTestProvider) CallContract(context.Context, ethereum.CallMsg, *big.Int) ([]byte, error) {
	return []byte{byte(p.calls.Add(1))}, nil
}
func (p *readTestProvider) CodeAt(context.Context, common.Address, *big.Int) ([]byte, error) {
	return []byte{byte(p.codes.Add(1))}, nil
}
func (p *readTestProvider) HeaderByNumber(context.Context, *big.Int) (*types.Header, error) {
	return &types.Header{Number: big.NewInt(5), Time: p.fork.Load()}, nil
}
func TestPreparationProviderKeepsLiveCallsAndInvalidatesOnReorg(t *testing.T) {
	base := &readTestProvider{}
	address := common.HexToAddress("0x01")
	p := &preparationProvider{Provider: base, reads: newPreparationReads(), protocol: address, blocks: map[string]common.Hash{}}
	ctx := context.Background()
	block := big.NewInt(5)
	_, _ = p.HeaderByNumber(ctx, block)
	call := ethereum.CallMsg{To: &address, Data: crypto.Keccak256([]byte("nonces(address)"))[:4]}
	one, _ := p.CallContract(ctx, call, block)
	one[0] = 99
	two, _ := p.CallContract(ctx, call, block)
	if two[0] != 1 || base.calls.Load() != 1 {
		t.Fatal("pinned read not reused or aliased")
	}
	_, _ = p.CallContract(ctx, call, nil)
	_, _ = p.CallContract(ctx, call, nil)
	if base.calls.Load() != 3 {
		t.Fatal("latest nonce was cached")
	}
	base.fork.Store(1)
	_, _ = p.HeaderByNumber(ctx, block)
	_, _ = p.CallContract(ctx, call, block)
	if base.calls.Load() != 4 {
		t.Fatal("reorg did not invalidate reads")
	}
	for range 2 {
		_, _ = p.CodeAt(ctx, address, block)
	}
	if base.codes.Load() != 1 {
		t.Fatal("code read not reused")
	}
}
