package store

import (
	"sync"

	"github.com/spaolacci/murmur3"
)

// Store represents a in-memory Key/Value implementation
type Store interface {

	// Set takes a key and value and stores with in the underlying store.
	// Returns true if it's over writting an existing value.
	Set(key string, value []byte) bool

	// Get returns the value associated for the key with in the underlying store.
	// Returns true if the value is found along with the value.
	Get(key string) ([]byte, bool)

	// Delete removes a value associated with the key.
	// Returns true if the value is found when deleting.
	Delete(key string) bool
}

type memory struct {
	size    uint
	buckets []Store
}

// NewBucket creates a new in-memory Store according to the size required by
// the value requested.
func NewBucket(size uint) Store {
	buckets := make([]Store, size)
	for k := range buckets {
		buckets[k] = New()
	}

	return &memory{
		size:    size,
		buckets: buckets,
	}
}

func (m *memory) Set(key string, value []byte) bool {
	index := uint(murmur3.Sum32([]byte(key))) % m.size
	return m.buckets[index].Set(key, value)
}

func (m *memory) Get(key string) ([]byte, bool) {
	index := uint(murmur3.Sum32([]byte(key))) % m.size
	return m.buckets[index].Get(key)
}

func (m *memory) Delete(key string) bool {
	index := uint(murmur3.Sum32([]byte(key))) % m.size
	return m.buckets[index].Delete(key)
}

// bucket conforms to the Key/Val store interface and provides locking mechanism
// for each bucket.
// values are stored in a simple map, it is entirely possible to replace this
// map with the newer https://golang.org/pkg/sync/#Map, but we will loose some
// portability.
type bucket struct {
	mutex  sync.RWMutex
	values map[string][]byte
}

// New creates a store from a singular bucket
func New() Store {
	return &bucket{
		values: make(map[string][]byte),
	}
}

func (b *bucket) Set(key string, value []byte) bool {
	b.mutex.Lock()
	_, ok := b.values[key]
	b.values[key] = value
	b.mutex.Unlock()
	return ok
}

func (b *bucket) Get(key string) ([]byte, bool) {
	b.mutex.RLock()
	value, ok := b.values[key]
	b.mutex.RUnlock()
	return value, ok
}

func (b *bucket) Delete(key string) bool {
	b.mutex.Lock()
	_, ok := b.values[key]
	delete(b.values, key)
	b.mutex.Unlock()
	return ok
}
