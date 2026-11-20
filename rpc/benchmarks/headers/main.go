package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/ethereum/go-ethereum/core/types"
	r "github.com/invisi-labs/invisi-sdk-go/rpc"
	"math/big"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type counted struct {
	http.RoundTripper
	n atomic.Int64
}

func (c *counted) RoundTrip(q *http.Request) (*http.Response, error) {
	c.n.Add(1)
	return c.RoundTripper.RoundTrip(q)
}
func main() {
	pins := flag.String("deployment", "", "Pinned deployment JSON file")
	endpoint := flag.String("rpc", "https://rpc.invisipay.fi/v1/chains/97", "Protocol RPC URL")
	first := flag.Uint64("from", 0, "First historical block")
	length := flag.Int("count", 256, "Block count (1..256)")
	runs := flag.Int("runs", 3, "Alternating comparison runs")
	flag.Parse()
	if *pins == "" || *first == 0 || *length < 1 || *length > 256 || *runs < 1 || *first > 1<<53-uint64(*length) {
		panic("deployment/from required; count 1..256 and runs > 0")
	}
	raw, e := os.ReadFile(*pins)
	if e != nil {
		panic(e)
	}
	var d r.ProtocolDeployment
	if e = json.Unmarshal(raw, &d); e != nil {
		panic(e)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer cancel()
	t := &counted{RoundTripper: http.DefaultTransport}
	client, e := r.NewProtocolRpcClient(*endpoint, r.ClientOptions{Deployment: &d, HTTPClient: &http.Client{Transport: t, Timeout: 40 * time.Second}})
	if e != nil {
		panic(e)
	}
	chain, e := d.ChainNumber()
	if e != nil {
		panic(e)
	}
	p, e := r.NewTrustedProvider(client, r.DeploymentContext{ChainID: chain.Uint64(), ProtocolAddress: d.Protocol.Address})
	if e != nil {
		panic(e)
	}
	start, count := *first, *length
	nums := make([]*big.Int, count)
	for i := range nums {
		nums[i] = new(big.Int).SetUint64(start + uint64(i))
	}
	anchor, e := p.HeaderByNumber(ctx, nums[count-1])
	if e != nil {
		panic(e)
	}
	for run := 0; run < *runs; run++ {
		for _, batch := range []bool{false, true} {
			before := t.n.Load()
			beg := time.Now()
			var heads []*types.Header
			if batch {
				heads, e = p.HeadersByNumber(ctx, nums)
			} else {
				heads = make([]*types.Header, count)
				for offset := 0; offset < count; offset += 8 {
					var group sync.WaitGroup
					size := min(8, count-offset)
					errs := make([]error, size)
					for j := 0; j < size; j++ {
						group.Add(1)
						go func(j int) { defer group.Done(); heads[offset+j], errs[j] = p.HeaderByNumber(ctx, nums[offset+j]) }(j)
					}
					group.Wait()
					for _, err := range errs {
						if err != nil {
							e = err
							break
						}
					}
					if e != nil {
						break
					}
				}
			}
			if e != nil {
				panic(e)
			}
			for i, h := range heads {
				if h.Number.Cmp(nums[i]) != 0 {
					panic("number mismatch")
				}
				if i > 0 && h.ParentHash != heads[i-1].Hash() {
					panic("parent mismatch")
				}
			}
			final, err := p.HeaderByNumber(ctx, nums[count-1])
			if err != nil {
				panic(err)
			}
			if final.Hash() != anchor.Hash() || heads[count-1].Hash() != anchor.Hash() {
				panic("canonical conflict")
			}
			result := map[string]any{"run": run + 1, "batch": batch, "blocks": count, "fromBlock": start, "toBlock": start + uint64(count) - 1, "anchor": anchor.Hash().Hex(), "elapsedMs": float64(time.Since(beg).Microseconds()) / 1000, "httpRequests": t.n.Load() - before, "validated": true}
			b, _ := json.Marshal(result)
			fmt.Println(string(b))
		}
	}
}
