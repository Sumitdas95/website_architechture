package memorycache

import (
	"fmt"
	"time"

	lru "github.com/hashicorp/golang-lru"
)

// Backend returns a new in memory storage backend that uses a least recently
// used cache eviction strategy. Once the cache is full, the least recently used
// item is evicted from the cache.
func LRUBackend(size int) *LRUStorageBackend {
	l, _ := lru.New(size)
	return &LRUStorageBackend{lru: l}
}

// LRUStorageBackend is a storage backend that persists in memory.
type LRUStorageBackend struct {
	lru *lru.Cache
}

type item struct {
	Value  []byte
	Expiry time.Time
}

// Get returns the value related to a key, or nil if it doesn't exist or if the
// value is expired.
func (b *LRUStorageBackend) Get(key string) ([]byte, error) {
	rawCachedItem, ok := b.lru.Get(key)
	if !ok {
		return nil, nil
	}
	var cachedItem item
	if cachedItem, ok = rawCachedItem.(item); !ok {
		return nil, fmt.Errorf("could not deserialize item %+v", rawCachedItem)
	}
	if time.Now().After(cachedItem.Expiry) {
		b.lru.Remove(key)
		return nil, nil
	}
	return cachedItem.Value, nil
}

// Set stores the given value in memory.
func (b *LRUStorageBackend) Set(key string, val []byte, ttl time.Duration) error {
	b.lru.Add(key, item{Value: val, Expiry: time.Now().Add(ttl)})
	return nil
}

// Flush empties the in-memory cache used to store the cached values.
func (b *LRUStorageBackend) Flush() error {
	b.lru.Purge()
	return nil
}
