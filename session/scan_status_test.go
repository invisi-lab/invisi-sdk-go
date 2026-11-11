package session

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
	"testing"
)

func TestScanStatusGapsDoNotFreezeKnownFunds(t *testing.T) {
	state := emptyState()
	state.Scan.RecoveryBoundary = 900
	state.Scan.Coverage = []CoverageRange{{100, protocolrpc.CanonicalBlock{Number: 300, Hash: common.HexToHash("1")}}, {900, protocolrpc.CanonicalBlock{Number: 1000, Hash: common.HexToHash("2")}}}
	state.Scan.ChainHead = &protocolrpc.CanonicalBlock{Number: 1000}
	state.Scan.IndexedHead = state.Scan.ChainHead
	status := BuildScanStatus(100, state, PublicHistoryState{})
	if status.MaxScannedHeight == nil || *status.MaxScannedHeight != 1000 || status.FullyScannedHeight == nil || *status.FullyScannedHeight != 300 || status.IsSynced || status.NextAction != "recover-history" {
		t.Fatalf("%+v", status)
	}
	if status.Progress.Overall.ScannedBlocks != 302 || status.Progress.Recovery.ScannedBlocks != 201 || status.Progress.Overall.Ratio == nil || *status.Progress.Overall.Ratio != 0.3351 {
		t.Fatalf("%+v", status.Progress)
	}
}
func TestScanStatusIndexerLagAndCreditPending(t *testing.T) {
	state := emptyState()
	state.Scan.Coverage = []CoverageRange{{100, protocolrpc.CanonicalBlock{Number: 1000, Hash: common.HexToHash("1")}}}
	state.Scan.ChainHead = &protocolrpc.CanonicalBlock{Number: 1100}
	state.Scan.IndexedHead = &protocolrpc.CanonicalBlock{Number: 1000, Hash: common.HexToHash("1")}
	status := BuildScanStatus(100, state, PublicHistoryState{})
	if status.ScanComplete || status.NextAction != "wait-for-indexer" {
		t.Fatalf("%+v", status)
	}
	state.Scan.ChainHead = state.Scan.IndexedHead
	state.Scan.CreditReplayTarget = 1000
	status = BuildScanStatus(100, state, PublicHistoryState{})
	if !status.ScanComplete || status.IsSynced || !status.CreditRecoveryPending {
		t.Fatalf("%+v", status)
	}
}

func TestSharedNewHeadMakesPreviouslyCheckedFundsStale(t *testing.T) {
	s, token, _ := testSession(t)
	ctx := context.Background()
	head := s.options.Provider.(sessionProvider).head
	b := protocolrpc.CanonicalBlock{Number: 10, Hash: head.Hash(), ParentHash: head.ParentHash, Timestamp: head.Time}
	if e := s.options.Store.Update(ctx, s.scope, func(v *WalletState) error {
		v.Scan.ChainHead = &b
		for key, n := range v.Notes {
			checked := b
			n.CheckedThrough = &checked
			v.Notes[key] = n
		}
		return nil
	}); e != nil {
		t.Fatal(e)
	}
	status, e := s.GetFundsStatus(ctx, token)
	if e != nil || !status.AreFundsSpendable {
		t.Fatal("fresh known funds unavailable", e)
	}
	if e = s.publicHistory.Store().Update(ctx, s.publicHistory.Key(), func(v *PublicHistoryState) error {
		v.KnownHead = &protocolrpc.CanonicalBlock{Number: 11, Hash: common.HexToHash("0x11")}
		return nil
	}); e != nil {
		t.Fatal(e)
	}
	status, e = s.GetFundsStatus(ctx, token)
	if e != nil || status.AreFundsSpendable || status.SpendableKnown.Sign() != 0 {
		t.Fatal("shared newer head left stale funds spendable", e)
	}
}
