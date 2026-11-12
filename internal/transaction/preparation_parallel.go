package transaction

import (
	"context"
	"sync"
)

// MapPreparation preserves statement order while limiting independent work to four workers.
func MapPreparation[T any](ctx context.Context, count int, work func(context.Context, int) (T, error)) ([]T, error) {
	values := make([]T, count)
	failures := make([]error, count)
	var group sync.WaitGroup
	var mu sync.Mutex
	next := 0
	for worker := 0; worker < min(4, count); worker++ {
		group.Add(1)
		go func() {
			defer group.Done()
			for {
				mu.Lock()
				i := next
				next++
				mu.Unlock()
				if i >= count {
					return
				}
				if err := ctx.Err(); err != nil {
					failures[i] = err
					continue
				}
				values[i], failures[i] = work(ctx, i)
			}
		}()
	}
	group.Wait()
	for _, err := range failures {
		if err != nil {
			return nil, err
		}
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return values, nil
}
