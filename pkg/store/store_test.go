package store_test

import (
	"reflect"
	"testing"
	"testing/quick"

	"github.com/SimonRichardson/keyval/pkg/store"
)

func testStore(t *testing.T, store func() store.Store) {
	t.Run("setting store value returns false", func(t *testing.T) {
		fn := func(key string, value []byte) bool {
			s := store()
			return !s.Set(key, value)
		}
		if err := quick.Check(fn, nil); err != nil {
			t.Error(err)
		}
	})

	t.Run("setting store value twice returns true", func(t *testing.T) {
		fn := func(key string, value []byte) bool {
			s := store()
			s.Set(key, value)
			return s.Set(key, value)
		}
		if err := quick.Check(fn, nil); err != nil {
			t.Error(err)
		}
	})

	t.Run("getting empty store value returns false", func(t *testing.T) {
		fn := func(key string) bool {
			s := store()
			_, ok := s.Get(key)
			return !ok
		}
		if err := quick.Check(fn, nil); err != nil {
			t.Error(err)
		}
	})

	t.Run("setting then getting empty store value returns true", func(t *testing.T) {
		fn := func(key string, value []byte) bool {
			s := store()
			s.Set(key, value)
			_, ok := s.Get(key)
			return ok
		}
		if err := quick.Check(fn, nil); err != nil {
			t.Error(err)
		}
	})

	t.Run("setting then getting empty store value returns value", func(t *testing.T) {
		fn := func(key string, value []byte) bool {
			s := store()
			s.Set(key, value)
			result, _ := s.Get(key)
			return reflect.DeepEqual(value, result)
		}
		if err := quick.Check(fn, nil); err != nil {
			t.Error(err)
		}
	})

	t.Run("deleting empty store value returns false", func(t *testing.T) {
		fn := func(key string) bool {
			s := store()
			return !s.Delete(key)
		}
		if err := quick.Check(fn, nil); err != nil {
			t.Error(err)
		}
	})

	t.Run("deleting store value returns true", func(t *testing.T) {
		fn := func(key string, value []byte) bool {
			s := store()
			s.Set(key, value)
			return s.Delete(key)
		}
		if err := quick.Check(fn, nil); err != nil {
			t.Error(err)
		}
	})
}

func TestStore(t *testing.T) {
	t.Parallel()

	testStore(t, func() store.Store {
		return store.New()
	})
}

func TestBucketStore(t *testing.T) {
	t.Parallel()

	testStore(t, func() store.Store {
		return store.NewBucket(10)
	})
}

// value used to make sure that we don't get compiled away
var benchResult []byte

func benchmarkStore(b *testing.B, store store.Store) {
	var (
		result []byte
		value  = []byte{62, 28, 16, 68, 46, 60, 84, 213, 9, 129}
	)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Force the integer to a string, so we're testing the store and not
		// how to convert ints to strings.
		key := string(i)

		store.Set(key, value)

		// Make `get` asymmetric
		for k := 0; k < 10; k++ {
			result, _ = store.Get(key)
		}
	}

	benchResult = result
}

func BenchmarkStore(b *testing.B)   { benchmarkStore(b, store.New()) }
func BenchmarkStore1(b *testing.B)  { benchmarkStore(b, store.NewBucket(1)) }
func BenchmarkStore2(b *testing.B)  { benchmarkStore(b, store.NewBucket(2)) }
func BenchmarkStore4(b *testing.B)  { benchmarkStore(b, store.NewBucket(4)) }
func BenchmarkStore8(b *testing.B)  { benchmarkStore(b, store.NewBucket(8)) }
func BenchmarkStore16(b *testing.B) { benchmarkStore(b, store.NewBucket(16)) }
func BenchmarkStore32(b *testing.B) { benchmarkStore(b, store.NewBucket(32)) }
func BenchmarkStore64(b *testing.B) { benchmarkStore(b, store.NewBucket(64)) }
