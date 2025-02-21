package pool

import (
	"sync"
	"testing"
)

// TestBasic tests the basic functionality of Get and Put methods.
func TestBasic(t *testing.T) {
	p := NewPool(func() interface{} {
		return new(int)
	})

	obj := p.Get()
	if obj == nil {
		t.Error("Expected non-nil object from Get")
	}

	p.Put(obj)
}

// TestConcurrency tests the concurrency safety of the Pool.
func TestConcurrency(t *testing.T) {
	p := NewPool(func() interface{} {
		return new(int)
	})

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			obj := p.Get()
			if obj == nil {
				t.Error("Expected non-nil object from Get")
			}
			p.Put(obj)
		}()
	}
	wg.Wait()
}

// TestClear tests the Clear method of the Pool.
func TestClear(t *testing.T) {
	p := NewPool(func() interface{} {
		return new(int)
	})

	obj := p.Get()
	p.Put(obj)
	p.Clear()

	// After clearing, the Pool should be empty
	if obj := p.Get(); obj == nil {
		t.Error("Expected non-nil object from Get after Clear")
	}
}

// TestCapacity tests the capacity limit of the Pool.
func TestCapacity(t *testing.T) {
	p := NewPool(func() interface{} {
		return new(int)
	})

	// Fill up a shard
	for i := 0; i < shardCap; i++ {
		p.Put(new(int))
	}

	// Try to put one more object, it should not be added
	obj := new(int)
	p.Put(obj)

	// Check if the object was not actually added
	for i := 0; i < shardCap; i++ {
		p.Get()
	}
	if p.Get() != nil {
		t.Error("Expected nil object from Get after exceeding capacity")
	}
}

// TestShardDistribution tests the shard distribution mechanism of the Pool.
func TestShardDistribution(t *testing.T) {
	p := NewPool(func() interface{} {
		return new(int)
	})

	// Get multiple objects and check if they come from different shards
	shardIDs := make(map[uint64]bool)
	for i := 0; i < shardCount; i++ {
		obj := p.Get()
		shardID := p.shardID()
		shardIDs[shardID] = true
		p.Put(obj)
	}

	// Check if multiple shards were used
	if len(shardIDs) < 2 {
		t.Error("Expected objects to be distributed across multiple shards")
	}
}

// BenchmarkCustomPool tests the performance of the custom Pool.
func BenchmarkCustomPool(b *testing.B) {
	p := NewPool(func() interface{} {
		return new(int)
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		obj := p.Get()
		p.Put(obj)
	}
}

// BenchmarkSyncPool tests the performance of the standard library sync.Pool.
func BenchmarkSyncPool(b *testing.B) {
	p := &sync.Pool{
		New: func() interface{} {
			return new(int)
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		obj := p.Get()
		p.Put(obj)
	}
}

// BenchmarkCustomPoolParallel tests the performance of the custom Pool in multiple goroutines.
func BenchmarkCustomPoolParallel(b *testing.B) {
	p := NewPool(func() interface{} {
		return new(int)
	})

	var wg sync.WaitGroup
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			obj := p.Get()
			p.Put(obj)
		}()
	}
	wg.Wait()
}

// BenchmarkSyncPoolParallel tests the performance of the standard library sync.Pool in multiple goroutines.
func BenchmarkSyncPoolParallel(b *testing.B) {
	p := &sync.Pool{
		New: func() interface{} {
			return new(int)
		},
	}

	var wg sync.WaitGroup
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			obj := p.Get()
			p.Put(obj)
		}()
	}
	wg.Wait()
}
