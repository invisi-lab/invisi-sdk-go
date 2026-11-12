package transaction

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestPreparationWorkersBoundedOrderedAndCancelled(t *testing.T) {
	var active, peak atomic.Int32
	got, err := MapPreparation(context.Background(), 10, func(ctx context.Context, i int) (int, error) {
		n := active.Add(1)
		defer active.Add(-1)
		for old := peak.Load(); n > old && !peak.CompareAndSwap(old, n); old = peak.Load() {
		}
		time.Sleep(time.Duration(10-i) * time.Millisecond)
		return i * 2, nil
	})
	if err != nil || peak.Load() > 4 || peak.Load() < 2 {
		t.Fatalf("workers: %d %v", peak.Load(), err)
	}
	for i, v := range got {
		if v != i*2 {
			t.Fatal("output order changed")
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := MapPreparation(ctx, 10, func(context.Context, int) (int, error) { t.Fatal("started cancelled work"); return 0, nil }); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}
func TestPreparationReadsOwnResultsRetryAndExpire(t *testing.T) {
	reads := newPreparationReads()
	var calls atomic.Int32
	now := time.Now()
	reads.now = func() time.Time { return now }
	load := func() ([]byte, error) { calls.Add(1); return []byte{1, 2}, nil }
	one, err := reads.read(context.Background(), "x", load)
	if err != nil {
		t.Fatal(err)
	}
	one[0] = 9
	two, err := reads.read(context.Background(), "x", load)
	if err != nil || two[0] != 1 || calls.Load() != 1 {
		t.Fatal(two, calls.Load(), err)
	}
	now = now.Add(31 * time.Second)
	if _, err = reads.read(context.Background(), "x", load); err != nil || calls.Load() != 2 {
		t.Fatal(calls.Load(), err)
	}
	failure := errors.New("offline")
	_, err = reads.read(context.Background(), "retry", func() ([]byte, error) { return nil, failure })
	if !errors.Is(err, failure) {
		t.Fatal(err)
	}
	if _, err = reads.read(context.Background(), "retry", load); err != nil || calls.Load() != 3 {
		t.Fatal(calls.Load(), err)
	}
	for i := 0; i < 1100; i++ {
		_, _ = reads.read(context.Background(), time.Unix(int64(i), 0).String(), load)
	}
	if len(reads.entries) > 1024 {
		t.Fatal("unbounded read cache")
	}
	shared := []byte{7}
	overflow, err := reads.read(context.Background(), "overflow", func() ([]byte, error) { return shared, nil })
	if err != nil {
		t.Fatal(err)
	}
	overflow[0] = 9
	if shared[0] != 7 {
		t.Fatal("cache bypass returned provider-owned bytes")
	}

}
