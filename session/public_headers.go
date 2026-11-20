package session

import (
	"context"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/core/types"
)

type publicHeaderProvider interface {
	HeaderByNumber(context.Context, *big.Int) (*types.Header, error)
}
type publicBatchHeaderProvider interface {
	HeadersByNumber(context.Context, []*big.Int) ([]*types.Header, error)
}

func publicPageHeaders(ctx context.Context, provider publicHeaderProvider, q BlockRange) ([]*types.Header, error) {
	numbers := make([]*big.Int, q.ToBlock-q.FromBlock+1)
	for i := range numbers {
		numbers[i] = new(big.Int).SetUint64(q.FromBlock + uint64(i))
	}
	if batch, ok := provider.(publicBatchHeaderProvider); ok {
		return batch.HeadersByNumber(ctx, numbers)
	}
	headers := make([]*types.Header, len(numbers))
	for offset := 0; offset < len(numbers); offset += 8 {
		end := min(offset+8, len(numbers))
		errs := make([]error, end-offset)
		var group sync.WaitGroup
		for i := offset; i < end; i++ {
			group.Add(1)
			go func(i int) { defer group.Done(); headers[i], errs[i-offset] = provider.HeaderByNumber(ctx, numbers[i]) }(i)
		}
		group.Wait()
		for _, err := range errs {
			if err != nil {
				return nil, err
			}
		}
	}
	return headers, nil
}
