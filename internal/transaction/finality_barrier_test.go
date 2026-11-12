package transaction

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
	"math/big"
	"testing"
)

type disclosureReorgProvider struct {
	*noteProvider
	canonicalReads int
}

func (p *disclosureReorgProvider) TransactionReceipt(context.Context, common.Hash) (*types.Receipt, error) {
	return nil, nil
}
func (p *disclosureReorgProvider) HeaderByNumber(ctx context.Context, n *big.Int) (*types.Header, error) {
	h, err := p.noteProvider.HeaderByNumber(ctx, n)
	if n != nil && n.Sign() >= 0 && n.Uint64() == 10 {
		p.canonicalReads++
		if p.canonicalReads == 4 {
			changed := types.CopyHeader(p.headers[10])
			changed.Extra = []byte("reorg after finality observation")
			p.headers[10] = changed
		}
	}
	return h, err
}
func TestDisclosureRechecksCanonicalAnchorAfterNativeFinality(t *testing.T) {
	f := newNoteFixture(t)
	p := &disclosureReorgProvider{noteProvider: f.provider}
	v, err := NewDisclosureVerifier(DisclosureVerifierOptions{Protocol: f.pool, Provider: p, TokenAddress: f.note.TokenAddress()})
	if err != nil {
		t.Fatal(err)
	}
	h := f.provider.headers[10]
	block := protocolrpc.CanonicalBlock{Number: 10, Hash: h.Hash(), ParentHash: h.ParentHash, Timestamp: h.Time}
	if _, err = v.barrier(context.Background(), block); err == nil {
		t.Fatal("orphaned disclosure anchor survived native finality observation")
	}
}
