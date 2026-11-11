package session

import (
	"github.com/ethereum/go-ethereum/common"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
	"testing"
)

func TestCoverageKeepsGapsAndRejectsConflictingAnchor(t *testing.T) {
	r, err := MergeCoverage(nil, CoverageRange{100, protocolrpc.CanonicalBlock{Number: 300, Hash: common.HexToHash("0x01")}})
	if err != nil {
		t.Fatal(err)
	}
	r, err = MergeCoverage(r, CoverageRange{900, protocolrpc.CanonicalBlock{Number: 1000, Hash: common.HexToHash("0x02")}})
	if err != nil {
		t.Fatal(err)
	}
	gaps := MissingRanges(r, 100, 1000)
	if len(gaps) != 1 || gaps[0] != (BlockRange{301, 899}) {
		t.Fatalf("gaps: %+v", gaps)
	}
	if FullyScannedHeight(r, 100) == nil || *FullyScannedHeight(r, 100) != 300 {
		t.Fatal("max must not replace prefix")
	}
	if _, err = MergeCoverage(r, CoverageRange{950, protocolrpc.CanonicalBlock{Number: 1000, Hash: common.HexToHash("0x03")}}); err == nil {
		t.Fatal("conflicting anchors accepted")
	}
}
func TestPriorityRangesCatchUpThenDescending(t *testing.T) {
	r := []CoverageRange{{300, protocolrpc.CanonicalBlock{Number: 500, Hash: common.HexToHash("0x01")}}}
	pages := PriorityRanges(r, 100, 700, 100)
	want := []BlockRange{{501, 600}, {601, 700}, {200, 299}, {100, 199}}
	if len(pages) != len(want) {
		t.Fatalf("%+v", pages)
	}
	for i := range want {
		if pages[i] != want[i] {
			t.Fatalf("%+v", pages)
		}
	}
}

func TestNewAccountStartsAtRetainedWindowThenCatchesUp(t *testing.T) {
	pages := []PublicPageReference{{Range: BlockRange{1000, 1100}}}
	got := AccountPriorityRanges(nil, pages, 100, 1300, 100)
	want := []BlockRange{{1000, 1099}, {1100, 1199}, {1200, 1299}, {1300, 1300}, {900, 999}}
	if len(got) < len(want) {
		t.Fatal(got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("priority %+v", got)
		}
	}
}
