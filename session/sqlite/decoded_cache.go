package sqlite

import (
	"container/list"
	"crypto/sha256"
	"sync"

	"github.com/invisi-lab/invisi-sdk-go/session"
)

type decodedRecord struct {
	key    string
	digest [32]byte
	plain  []byte
	weight int
}
type decodedRecordCache struct {
	mu      sync.Mutex
	entries map[string]*list.Element
	lru     list.List
	bytes   int
}

func (c *decodedRecordCache) open(key string, ciphertext []byte, decode func() ([]byte, error)) ([]byte, error) {
	digest := sha256.Sum256(ciphertext)
	c.mu.Lock()
	if c.entries == nil {
		c.entries = make(map[string]*list.Element)
	}
	if e := c.entries[key]; e != nil {
		item := e.Value.(decodedRecord)
		if item.digest == digest {
			c.lru.MoveToBack(e)
			result := append([]byte(nil), item.plain...)
			c.mu.Unlock()
			return result, nil
		}
		c.remove(e)
	}
	c.mu.Unlock()
	plain, err := decode()
	if err != nil {
		return nil, err
	}
	weight := max(len(ciphertext), len(plain))
	if weight > 8<<20 {
		return plain, nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if old := c.entries[key]; old != nil {
		c.remove(old)
	}
	for len(c.entries) >= 512 || c.bytes+weight > 8<<20 {
		c.remove(c.lru.Front())
	}
	item := decodedRecord{key: key, digest: digest, plain: append([]byte(nil), plain...), weight: weight}
	c.entries[key] = c.lru.PushBack(item)
	c.bytes += weight
	return plain, nil
}
func (c *decodedRecordCache) remove(e *list.Element) {
	item := e.Value.(decodedRecord)
	clear(item.plain)
	delete(c.entries, item.key)
	c.bytes -= item.weight
	c.lru.Remove(e)
}
func (s *WalletStore) openRecord(scope session.WalletScope, locator string, data []byte) ([]byte, error) {
	return s.decoded.open(scope.Key()+":"+locator, data, func() ([]byte, error) { return s.codec.OpenRecord(scope, locator, data) })
}
