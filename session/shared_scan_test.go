package session

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
	"testing"
)

func TestPrivateCoverageCodecRoundtrip(t *testing.T) {
	c, e := NewEncryptedWalletCodec(make([]byte, 32))
	if e != nil {
		t.Fatal(e)
	}
	s := emptyState()
	s.Scan.Coverage = []CoverageRange{{100, protocolrpc.CanonicalBlock{Number: 300, Hash: common.HexToHash("0x1")}}}
	s.Scan.CreditReplayTarget = 1000
	s.Scan.CreditReplayRevision = 9
	s.Scan.RecoveryBoundary = 900
	data, e := c.Seal(WalletScope{}, s)
	if e != nil {
		t.Fatal(e)
	}
	got, e := c.Open(WalletScope{}, data)
	if e != nil {
		t.Fatal(e)
	}
	if got.Scan.CreditReplayTarget != 1000 || got.Scan.CreditReplayRevision != 9 || got.Scan.RecoveryBoundary != 900 || len(got.Scan.Coverage) != 1 {
		t.Fatal("scan metadata lost")
	}
	_ = context.Background()
}

func TestOlderSnapshotCannotClearCreditReplayObligation(t *testing.T) {
	s, _ := sharedEmptyFixture(t)
	ctx := context.Background()
	if e := s.options.Store.Update(ctx, s.scope, func(v *WalletState) error {
		v.Scan.CreditReplayTarget = 16
		v.Scan.CreditReplayCursor = 10
		v.Scan.CreditReplayRevision = 3
		return nil
	}); e != nil {
		t.Fatal(e)
	}
	if e := s.replayCredits(ctx, protocolrpc.CanonicalBlock{Number: 14}, 100, ScanOptions{}, &RefreshResult{}); e != nil {
		t.Fatal(e)
	}
	v, e := s.options.Store.ReadSnapshot(ctx, s.scope)
	if e != nil {
		t.Fatal(e)
	}
	if v.Scan.CreditReplayTarget != 16 || v.Scan.CreditReplayCursor != 10 || v.Scan.CreditReplayRevision != 3 {
		t.Fatal("older snapshot discarded required replay")
	}
}
