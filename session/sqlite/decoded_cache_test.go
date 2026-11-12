package sqlite

import (
	"bytes"
	"errors"
	"fmt"
	"testing"
)

func TestDecodedCacheAuthenticatesChangedBytesAndOwnsValues(t *testing.T) {
	var cache decodedRecordCache
	calls := 0
	load := func() ([]byte, error) { calls++; return []byte("private"), nil }
	data := []byte("ciphertext")
	one, err := cache.open("scope/row", data, load)
	if err != nil {
		t.Fatal(err)
	}
	one[0] = 'X'
	two, err := cache.open("scope/row", data, load)
	if err != nil || !bytes.Equal(two, []byte("private")) || calls != 1 {
		t.Fatal(string(two), calls, err)
	}
	bad := errors.New("authentication failed")
	if _, err = cache.open("scope/row", []byte("changed"), func() ([]byte, error) { return nil, bad }); !errors.Is(err, bad) {
		t.Fatal(err)
	}
	if _, err = cache.open("another-scope/row", data, load); err != nil || calls != 2 {
		t.Fatal(calls, err)
	}
	for i := 0; i < 600; i++ {
		_, _ = cache.open(fmt.Sprint(i), data, load)
	}
	if len(cache.entries) > 512 || cache.bytes > 8<<20 {
		t.Fatal("unbounded decoded cache")
	}
}
