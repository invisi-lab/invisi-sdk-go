package session

import (
	"errors"
	"github.com/ethereum/go-ethereum/common"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
	"sort"
)

type BlockRange struct{ FromBlock, ToBlock uint64 }
type CoverageRange struct {
	FromBlock uint64
	Block     protocolrpc.CanonicalBlock
}

func MergeCoverage(ranges []CoverageRange, input CoverageRange) ([]CoverageRange, error) {
	sorted := append(append([]CoverageRange(nil), ranges...), input)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].FromBlock < sorted[j].FromBlock })
	out := []CoverageRange{}
	anchors := map[uint64]common.Hash{}
	for _, r := range sorted {
		if r.Block.Number < r.FromBlock || r.Block.Hash == (common.Hash{}) {
			return nil, errors.New("invalid history coverage")
		}
		if h, ok := anchors[r.Block.Number]; ok && h != r.Block.Hash {
			return nil, errors.New("conflicting history coverage anchor")
		}
		anchors[r.Block.Number] = r.Block.Hash
		if len(out) == 0 {
			out = append(out, r)
			continue
		}
		p := &out[len(out)-1]
		if p.Block.Number != ^uint64(0) && r.FromBlock > p.Block.Number+1 {
			out = append(out, r)
			continue
		}
		if r.Block.Number > p.Block.Number {
			p.Block = r.Block
		}
	}
	if len(out) > 4096 {
		return nil, errors.New("history coverage too fragmented")
	}
	return out, nil
}
func MissingRanges(ranges []CoverageRange, from, to uint64) []BlockRange {
	out := []BlockRange{}
	if from > to {
		return out
	}
	cursor := from
	sorted := append([]CoverageRange(nil), ranges...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].FromBlock < sorted[j].FromBlock })
	for _, r := range sorted {
		if r.Block.Number < cursor {
			continue
		}
		if r.FromBlock > to {
			break
		}
		if r.FromBlock > cursor {
			out = append(out, BlockRange{cursor, r.FromBlock - 1})
		}
		if r.Block.Number >= to {
			return out
		}
		cursor = r.Block.Number + 1
	}
	out = append(out, BlockRange{cursor, to})
	return out
}
func FullyScannedHeight(ranges []CoverageRange, start uint64) *uint64 {
	for _, r := range ranges {
		if r.FromBlock <= start && r.Block.Number >= start {
			v := r.Block.Number
			return &v
		}
	}
	return nil
}
func PriorityRanges(ranges []CoverageRange, from, to, size uint64) []BlockRange {
	if size == 0 {
		size = 100
	}
	out := []BlockRange{}
	gaps := MissingRanges(ranges, from, to)
	var max uint64
	have := false
	for _, r := range ranges {
		if !have || r.Block.Number > max {
			max = r.Block.Number
			have = true
		}
	}
	appendForward := func(a, b uint64) {
		for a <= b {
			end := a + size - 1
			if end < a || end > b {
				end = b
			}
			out = append(out, BlockRange{a, end})
			if end == b {
				break
			}
			a = end + 1
		}
	}
	if have && max < to {
		start := max + 1
		if start < from {
			start = from
		}
		appendForward(start, to)
	}
	for i := len(gaps) - 1; i >= 0; i-- {
		g := gaps[i]
		if have && g.FromBlock > max {
			continue
		}
		for g.FromBlock <= g.ToBlock {
			start := g.FromBlock
			if g.ToBlock-g.FromBlock+1 > size {
				start = g.ToBlock - size + 1
			}
			out = append(out, BlockRange{start, g.ToBlock})
			if start == g.FromBlock {
				break
			}
			g.ToBlock = start - 1
		}
	}
	return out
}

// AccountPriorityRanges starts a new account at the newest retained public window, then catches up before backfilling.
func AccountPriorityRanges(ranges []CoverageRange, pages []PublicPageReference, from, to, size uint64) []BlockRange {
	if len(ranges) > 0 {
		return PriorityRanges(ranges, from, to, size)
	}
	if size == 0 {
		size = 100
	}
	start := from
	if to >= size-1 && to-size+1 > start {
		start = to - size + 1
	}
	var newest *PublicPageReference
	for i := range pages {
		p := &pages[i]
		if p.Range.FromBlock > to || p.Range.ToBlock < from {
			continue
		}
		if newest == nil || p.Range.ToBlock > newest.Range.ToBlock {
			newest = p
		}
	}
	if newest != nil {
		start = newest.Range.FromBlock
		if start < from {
			start = from
		}
	}
	out := []BlockRange{}
	for n := start; n <= to; {
		end := n + size - 1
		if end < n || end > to {
			end = to
		}
		out = append(out, BlockRange{n, end})
		if end == to {
			break
		}
		n = end + 1
	}
	if start > from {
		out = append(out, PriorityRanges(nil, from, start-1, size)...)
	}
	return out
}
