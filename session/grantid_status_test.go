package session

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-labs/invisi-sdk-go/credit"
	"github.com/invisi-labs/invisi-sdk-go/primitives"
	"github.com/invisi-labs/invisi-sdk-go/protocol"
	protocolrpc "github.com/invisi-labs/invisi-sdk-go/rpc"
	"github.com/invisi-labs/invisi-sdk-go/sdk"
	creditsession "github.com/invisi-labs/invisi-sdk-go/session/credit"
)

type scopedGrantStatusProvider struct {
	sdk.Provider
	mu       sync.Mutex
	statuses map[string]uint8
	seen     map[string]bool
	block    uint64
}

func grantStatusScope(owner, token common.Address, id *big.Int) string {
	return fmt.Sprintf("%s:%s:%s", owner, token, id)
}
func (p *scopedGrantStatusProvider) CallContract(ctx context.Context, msg ethereum.CallMsg, block *big.Int) ([]byte, error) {
	abi, _ := protocol.ProtocolABI()
	if len(msg.Data) >= 4 {
		if method, err := abi.MethodById(msg.Data[:4]); err == nil && method.RawName == "creditStatus" {
			if block == nil || block.Uint64() != p.block {
				return nil, fmt.Errorf("status read was not pinned to canonical block")
			}
			args, err := method.Inputs.Unpack(msg.Data[4:])
			if err != nil {
				return nil, err
			}
			key := grantStatusScope(args[0].(common.Address), args[1].(common.Address), args[2].(*big.Int))
			value, ok := p.statuses[key]
			if !ok {
				return nil, fmt.Errorf("unexpected status namespace %s", key)
			}
			p.mu.Lock()
			p.seen[key] = true
			p.mu.Unlock()
			return method.Outputs.Pack(new(big.Int).SetUint64(uint64(value)))
		}
	}
	return p.Provider.CallContract(ctx, msg, block)
}

func TestGrantStatusesScopeOwnerTokenAndFullUint128AtCanonicalBlock(t *testing.T) {
	s, _, _ := testSession(t)
	ctx := context.Background()
	snapshot, err := s.options.Store.ReadSnapshot(ctx, s.scope)
	if err != nil {
		t.Fatal(err)
	}
	head := s.options.Provider.(sessionProvider).head
	block := protocolrpc.CanonicalBlock{Number: head.Number.Uint64(), Hash: head.Hash(), ParentHash: head.ParentHash, Timestamp: head.Time}
	provider := &scopedGrantStatusProvider{Provider: s.options.Provider, statuses: map[string]uint8{}, seen: map[string]bool{}, block: block.Number}
	id := new(big.Int).Add(new(big.Int).Lsh(big.NewInt(1), 100), big.NewInt(77))
	for i, status := range []uint8{1, 0, 2, 1} {
		state := syncGrantState()
		owner := common.BigToAddress(big.NewInt(int64(1 + i%2)))
		token := common.BigToAddress(big.NewInt(int64(34 + i/2)))
		state.OwnerWalletIdentity = primitives.BigIntToElement(owner.Big())
		state.TokenAddress = primitives.BigIntToElement(token.Big())
		state.GrantID = primitives.BigIntToElement(id)
		nf, err := state.Nullifier()
		if err != nil {
			t.Fatal(err)
		}
		key := fmt.Sprint(i)
		snapshot.PostpaidCredit.Grants[key] = creditsession.StoredState{Role: credit.Owner, Delivery: credit.DeliveryState{State: state}, Nullifier: common.Hash(nf.Bytes())}
		provider.statuses[grantStatusScope(owner, token, id)] = status
	}
	s.options.Provider = provider
	if err = s.checkKnownState(withScanStateCache(ctx), &snapshot, block); err != nil {
		t.Fatal(err)
	}
	for i, status := range []uint8{1, 0, 2, 1} {
		g := snapshot.PostpaidCredit.Grants[fmt.Sprint(i)]
		if g.GrantStatus != status || g.Revoked != (status == 2) || g.GrantStatusBlock == nil || *g.GrantStatusBlock != block {
			t.Fatalf("status %d incorrectly merged across owner/token namespaces", status)
		}
	}
	if len(provider.seen) != 4 {
		t.Fatalf("merged distinct grant identities: %d", len(provider.seen))
	}
}
