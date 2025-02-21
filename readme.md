# Go Sync Pool Implementation

This pool is inspired by the `sync.Pool` but completely independent of the Go runtime. Through the sharding and pseudo-local cache strategy, this implementation performs excellently in high-concurrency scenarios and is suitable for scenarios where objects need to be frequently allocated and recycled.

---

## Features

- **Sharding Design**: The object pool is divided into multiple shards to reduce lock contention.
- **Stealing Mechanism**: When there are no objects in the preferred shard, try to steal objects from other shards.
- **Shard Size Limit**: Each shard has a fixed capacity to prevent unlimited memory growth.
- **High Performance**: Achieve high-concurrency performance by optimizing the lock contention and shard selection strategy.

---

## Installation

```bash
go get github.com/ongniud/pool
```

---

## Usage Example

```go
package main

import (
	"fmt"
	
	"github.com/ongniud/pool"
)

func main() {
	// Create an object pool
	pl := pool.NewPool(func() interface{} {
		return make([]byte, 1024)
	})

	// Get an object from the pool
	obj := pl.Get().([]byte)
	fmt.Println("Got object from pool")

	// Use the object
	obj[0] = 1

	// Put the object back into the pool
	pl.Put(obj)
	fmt.Println("Put object back to pool")
}
```

---

## Performance Optimization

### Shard Selection Strategy

- **Pseudo-Local Cache**: Select the shard by using the low bits of the goroutine stack address to simulate the local cache effect.
- **Random Sharding**: Use random shard selection in the stealing mechanism to avoid hot spot issues.

### Stealing Mechanism

- When there are no objects in the preferred shard, try to steal objects from other shards, and try at most `stealShardCnt` shards.

### Shard Size Limit

- The maximum capacity of each shard is `shardCap` to prevent unlimited memory growth.

---

## Notes

1. **Selection of the Number of Shards**:
    - It is recommended that the number of shards be twice the number of CPU cores (such as 16, 32), and it should be a power of 2 to simplify the hash calculation.

2. **Object Lifecycle**:
    - The object pool will not automatically clean up objects that have not been used for a long time. You need to call the `Clear` method regularly.

3. **Concurrency Performance**:
    - In high-concurrency scenarios, the shard lock may become a performance bottleneck. It is recommended to adjust the number of shards and the shard size according to the actual load.

---

## Benchmark Tests

The following is a simple performance comparison with the `sync.Pool` in the Go standard library:

| Test Scenario         | This Pool (ns/op) | sync.Pool (ns/op) |
|------------------|-------------------|-------------------|
| Single-threaded Get/Put  | 27                | 12                |
| High-concurrency Get/Put  | 567               | 375                |

---

## Contribution

Welcome to submit issues and pull requests! Please ensure that the code style is consistent and all tests pass.

---

## License

This project is licensed under the MIT License. 
