package pool

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

const (
	// Maximum number of shards to steal from when the preferred shard is empty
	stealShardCnt = 4
	// Count of shard
	shardCount = 16
	// Capacity of each shard
	shardCap = 128
)

// Pool represents an object pool.
type Pool struct {
	shards    []poolShard
	shardMask uint64
	newFunc   func() interface{}
	tick      uint64
}

// NewPool creates a new object pool.
// fn is the function used to create a new object when the pool is empty.
func NewPool(fn func() interface{}) *Pool {
	if fn == nil {
		panic("newFunc cannot be nil")
	}
	p := &Pool{
		shards:    make([]poolShard, shardCount),
		shardMask: uint64(shardCount - 1),
		newFunc:   fn,
	}
	return p
}

// Get retrieves an object from the pool.
// 1. Try to get an object from the preferred shard.
// 2. If the preferred shard is empty, try to steal from other shards (up to 4 shards).
// 3. If all shards are empty, create a new object using the newFunc.
func (p *Pool) Get() interface{} {
	// 1. Try to get an object from the preferred shard
	shardID := p.shardID()
	shard := &p.shards[shardID]
	if obj := shard.pop(); obj != nil {
		return obj
	}

	// 2. Try to steal from other shards, up to 4 shards
	for i := 0; i < stealShardCnt; i++ {
		shardID = (shardID + 1) & p.shardMask
		shard = &p.shards[shardID]
		if obj := shard.pop(); obj != nil {
			return obj
		}
	}

	// 3. All shards are empty, create a new object
	return p.newFunc()
}

// Put returns an object to the pool.
// If the object is nil, it will be ignored.
func (p *Pool) Put(obj interface{}) {
	if obj == nil {
		return
	}
	shardID := p.shardID()
	p.shards[shardID].push(obj)
}

// shardID returns the ID of the shard to use.
func (p *Pool) shardID() uint64 {
	return p.shardIDGoID() & p.shardMask
}

// shardIDRand returns a shard ID using a random-like approach (incrementing tick).
func (p *Pool) shardIDRand() uint64 {
	return atomic.AddUint64(&p.tick, 1)
}

// shardIDGoID returns a shard ID using a fake goroutine ID approach.
// It uses the low bits of the goroutine stack address as the shard selection basis.
func (p *Pool) shardIDGoID() uint64 {
	var dummy int
	stackPtr := uintptr(unsafe.Pointer(&dummy))
	return uint64(stackPtr)
}

// Clear clears all objects from the pool.
func (p *Pool) Clear() {
	for i := range p.shards {
		shard := &p.shards[i]
		shard.mu.Lock()
		shard.objs = nil
		shard.mu.Unlock()
	}
}

// poolShard represents a single shard in the pool.
type poolShard struct {
	mu   sync.Mutex
	objs []interface{}
}

// pop removes and returns an object from the shard.
// If the shard is empty, it returns nil.
func (s *poolShard) pop() interface{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.objs) == 0 {
		return nil
	}
	obj := s.objs[len(s.objs)-1]
	s.objs = s.objs[:len(s.objs)-1]
	return obj
}

// push adds an object to the shard.
// If the shard has reached its capacity, the object will not be added.
func (s *poolShard) push(obj interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.objs) < shardCap {
		s.objs = append(s.objs, obj)
	}
}
