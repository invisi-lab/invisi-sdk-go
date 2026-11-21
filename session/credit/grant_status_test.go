package session

import (
	"context"
	"github.com/invisi-labs/invisi-sdk-go/primitives"
	protocolrpc "github.com/invisi-labs/invisi-sdk-go/rpc"
	"testing"
)

func TestUnknownStatusIsUnavailableWithoutReportingRevocation(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	session, err := New(Options{Store: store})
	if err != nil {
		t.Fatal(err)
	}
	prepared, grant := testPlan(t)
	if err = grant.SetGrantStatus(0, nil); err != nil {
		t.Fatal(err)
	}
	if err = session.Track(ctx, grant); err != nil {
		t.Fatal(err)
	}
	if _, err = session.Prepare(ctx, "unknown", prepared); err == nil {
		t.Fatal("unknown grant reserved for draw")
	}
	history, err := session.GetHistory(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if history.Grants[0].Revoked || history.Grants[0].GrantStatus != 0 || history.Grants[0].ReservedBy != "" {
		t.Fatal("unknown grant reported revoked or available")
	}
	_ = grant.SetGrantStatus(1, nil)
	if err = session.Track(ctx, grant); err != nil {
		t.Fatal(err)
	}
	if _, err = session.Prepare(ctx, "active", prepared); err != nil {
		t.Fatal("active grant unavailable", err)
	}
	_ = grant.SetGrantStatus(2, nil)
	if err = session.Track(ctx, grant); err != nil {
		t.Fatal(err)
	}
	snapshot, _ := store.ReadSnapshot(ctx)
	stored := snapshot.Grants[grant.key()]
	if stored.GrantStatus != 2 || !stored.Revoked {
		t.Fatal("terminal status not reported revoked")
	}
}

func TestHistoricalGrantStatusMergePreservesCanonicalObservation(t *testing.T) {
	_, grant := testPlan(t)
	older := protocolrpc.CanonicalBlock{Number: 10}
	later := protocolrpc.CanonicalBlock{Number: 20}
	for _, status := range []uint8{0, 1, 2} {
		previous := grant
		_ = previous.SetGrantStatus(status, &later)
		incoming := grant
		_ = incoming.SetGrantStatus(1, &older)
		merged := incoming.MergeGrantStatus(previous)
		if merged.GrantStatus != status || merged.Revoked != (status == 2) || merged.GrantStatusBlock.Number != 20 {
			t.Fatalf("older receipt overwrote status %d", status)
		}
	}
	terminal := grant
	_ = terminal.SetGrantStatus(2, &older)
	incoming := grant
	_ = incoming.SetGrantStatus(1, &later)
	if incoming.MergeGrantStatus(terminal).GrantStatus != 2 {
		t.Fatal("reused terminal grant")
	}
	incoming.Delivery.State.GrantID = primitives.ElementFromUint64(2)
	if incoming.MergeGrantStatus(terminal).GrantStatus != 1 {
		t.Fatal("status leaked into different grant ID")
	}
}

func TestFirstDiscoveredChildInheritsLaterScopedStatus(t *testing.T) {
	_, parent := testPlan(t)
	later := protocolrpc.CanonicalBlock{Number: 20}
	older := protocolrpc.CanonicalBlock{Number: 10}
	for _, status := range []uint8{0, 1, 2} {
		previous := parent
		_ = previous.SetGrantStatus(status, &later)
		child := parent
		child.Commitment[0] ^= 1
		child.Delivery.State.Nonce = primitives.ElementFromUint64(444)
		_ = child.SetGrantStatus(1, &older)
		merged := child.MergeGrantStatusFrom(map[string]StoredState{parent.key(): previous})
		if merged.GrantStatus != status || merged.Revoked != (status == 2) {
			t.Fatalf("new child lost later scoped status %d", status)
		}
	}
}

func TestTerminalGrantStatusMergeIsOrderIndependent(t *testing.T) {
	_, grant := testPlan(t)
	older := protocolrpc.CanonicalBlock{Number: 10}
	later := protocolrpc.CanonicalBlock{Number: 20}
	disabled := grant
	_ = disabled.SetGrantStatus(2, &older)
	active := grant
	_ = active.SetGrantStatus(1, &later)
	for _, first := range []StoredState{disabled, active} {
		second := disabled
		if first.GrantStatus == 2 {
			second = active
		}
		incoming := grant
		_ = incoming.SetGrantStatus(1, &older)
		result := incoming.MergeGrantStatus(first).MergeGrantStatus(second)
		if result.GrantStatus != 2 || !result.Revoked {
			t.Fatal("terminal grant revived by merge order")
		}
	}
	for range 32 {
		result := active.MergeGrantStatusFrom(map[string]StoredState{"disabled": disabled, "active": active})
		if result.GrantStatus != 2 || !result.Revoked {
			t.Fatal("terminal grant revived by map iteration")
		}
	}
}

func TestUnknownCanonicalObservationWinsEqualBlockEvent(t *testing.T) {
	_, grant := testPlan(t)
	block := protocolrpc.CanonicalBlock{Number: 20}
	unknown := grant
	_ = unknown.SetGrantStatus(0, &block)
	active := grant
	_ = active.SetGrantStatus(1, &block)
	for range 32 {
		result := active.MergeGrantStatusFrom(map[string]StoredState{"event": active, "observation": unknown})
		if result.GrantStatus != 0 || result.Revoked {
			t.Fatal("same-block event promoted unknown observation to active or revoked")
		}
	}
}
