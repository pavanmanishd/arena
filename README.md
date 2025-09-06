# Arena - High-Performance Memory Allocator for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/pavanmanishd/arena.svg)](https://pkg.go.dev/github.com/pavanmanishd/arena)
[![Go Report Card](https://goreportcard.com/badge/github.com/pavanmanishd/arena)](https://goreportcard.com/report/github.com/pavanmanishd/arena)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A production-ready, high-performance memory arena allocator for Go. Perfect for request-scoped allocations, reducing GC pressure, and applications requiring predictable memory allocation patterns.

## Features

- **Fast Allocations**: O(1) amortized allocation time
- **Memory Efficient**: Minimal overhead, configurable chunk sizes
- **Thread-Safe Option**: `SafeArena` for concurrent access
- **Zero-Copy Reset**: O(1) arena reset for reuse
- **Comprehensive Metrics**: Built-in monitoring and diagnostics
- **Type-Safe Generics**: Strongly-typed allocation functions
- **Production Ready**: Extensive test coverage and benchmarks

## Performance

Arena provides significant performance improvements over standard Go allocation, especially for scenarios with many temporary allocations:

### Core Allocation Benchmarks
*Tested on Apple M4 Pro (darwin/arm64)*

```
BenchmarkArenaAllocBytes/64B-14     743M ops    1.62 ns/op    0 B/op    0 allocs/op
BenchmarkArenaVsBuiltin/arena-14    747M ops    1.63 ns/op    0 B/op    0 allocs/op  
BenchmarkArenaVsBuiltin/builtin-14  1000M ops   0.23 ns/op    0 B/op    0 allocs/op
BenchmarkSafeArena/AllocBytes-14    256M ops    4.68 ns/op    0 B/op    0 allocs/op
BenchmarkSafeArena/SafeAlloc-14     233M ops    5.12 ns/op    0 B/op    0 allocs/op
```

### Real-World Scenario Comparisons

**Many Small Allocations (100 × 64B):**
```
Arena:   19.2M ops    110 ns/op     0 B/op      0 allocs/op
Builtin: 197K ops     12,386 ns/op  6400 B/op   100 allocs/op
```
**🚀 Arena is 112x faster with zero GC pressure!**

**Struct Allocations (50 structs):**
```
Arena:   16.9M ops    149 ns/op     0 B/op      0 allocs/op  
Builtin: 203K ops     11,633 ns/op  3200 B/op   50 allocs/op
```
**🚀 Arena is 78x faster with zero GC pressure!**

**Buffer Reuse (10 × mixed buffers):**
```
Arena:   48.5M ops    50.5 ns/op    0 B/op      0 allocs/op
Builtin: 92K ops      26,292 ns/op  35,841 B/op 30 allocs/op
```
**🚀 Arena is 521x faster with zero GC pressure!**

**Single Allocation Comparison:**
```
Arena:   856M ops     1.32 ns/op    0 B/op      0 allocs/op
Builtin: 1000M ops    0.26 ns/op    0 B/op      0 allocs/op
```
*Builtin is faster for individual allocations*

### Key Performance Benefits

- **Zero GC Pressure**: Arena allocations don't trigger garbage collection
- **Batch Cleanup**: O(1) cleanup of thousands of objects with `Reset()`
- **Memory Locality**: Sequential allocation improves cache performance  
- **Reduced Fragmentation**: Large chunks minimize heap fragmentation
- **Predictable Performance**: No GC pauses during allocation-heavy workloads

## Installation

```bash
go get github.com/pavanmanishd/arena
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/pavanmanishd/arena"
)

func main() {
    // Create a new arena
    a := arena.NewArena(0) // 0 = use default chunk size (64KB)
    defer a.Release()      // Always clean up

    // Allocate raw bytes
    buf := a.AllocBytes(1024)
    fmt.Printf("Allocated %d bytes\n", len(buf))

    // Allocate typed values
    ptr := arena.Alloc[int](a)
    *ptr = 42

    // Allocate slices
    slice := arena.AllocSlice[int](a, 100)
    for i := range slice {
        slice[i] = i
    }

    // Check memory usage
    fmt.Printf("Memory usage: %d/%d bytes (%.1f%%)\n",
        a.SizeInUse(), a.Capacity(), a.Utilization()*100)

    // Reset for reuse (O(1) operation)
    a.Reset()
    fmt.Printf("After reset: %d bytes in use\n", a.SizeInUse())
}
```

## Thread-Safe Usage

```go
// For concurrent access, use SafeArena
safeArena := arena.NewSafeArena(0)
defer safeArena.Release()

// All operations are thread-safe
go func() {
    buf := safeArena.AllocBytes(1024)
    ptr := arena.SafeAlloc[MyStruct](safeArena)
    // ... use allocated memory
}()
```

## API Overview

### Core Types

- `Arena`: Fast, single-threaded arena allocator
- `SafeArena`: Thread-safe wrapper around Arena
- `ArenaMetrics`: Detailed allocation statistics

### Allocation Functions

- `AllocBytes(n int) []byte`: Allocate raw bytes
- `Alloc[T any](a *Arena) *T`: Allocate typed value (zeroed)
- `AllocUninitialized[T any](a *Arena) *T`: Allocate without zeroing
- `AllocSlice[T any](a *Arena, n int) []T`: Allocate slice
- `AllocSliceZeroed[T any](a *Arena, n int) []T`: Allocate zeroed slice

### Management Functions

- `Reset()`: Reset allocation pointers (O(1))
- `Release()`: Free all memory and invalidate arena
- `EnsureCapacity(n int)`: Pre-allocate space for n bytes

### Metrics Functions

- `SizeInUse() int`: Currently allocated bytes
- `Capacity() int`: Total capacity across all chunks
- `NumChunks() int`: Number of allocated chunks
- `Utilization() float64`: Usage ratio (0.0-1.0)
- `Metrics() ArenaMetrics`: Complete metrics snapshot

## Use Cases

### Web Servers

Perfect for request-scoped allocations:

```go
func handleRequest(w http.ResponseWriter, r *http.Request) {
    arena := arena.NewArena(0)
    defer arena.Release()

    // Use arena for temporary allocations during request processing
    tempBuffer := arena.AllocBytes(4096)
    responseData := arena.AllocSlice[ResponseItem](arena, 100)
    
    // Process request using arena-allocated memory
    // Memory is automatically cleaned up when function returns
}
```

### High-Performance Computing

Reduce GC pressure in compute-intensive applications:

```go
func processData(data []float64) []Result {
    arena := arena.NewArena(1024 * 1024) // 1MB chunks
    defer arena.Release()

    results := arena.AllocSlice[Result](arena, len(data))
    workspace := arena.AllocSlice[float64](arena, len(data)*2)

    // Perform computations using arena-allocated workspace
    // No GC pressure from temporary allocations
    
    return results // Copy results before arena is released
}
```

### Object Pooling

Implement efficient object pools with arena backing:

```go
type ObjectPool struct {
    arena *arena.SafeArena
    mu    sync.Mutex
}

func (p *ObjectPool) Get() *MyObject {
    return arena.SafeAlloc[MyObject](p.arena)
}

func (p *ObjectPool) Reset() {
    p.mu.Lock()
    defer p.mu.Unlock()
    p.arena.Reset() // O(1) reset of all objects
}
```

## When to Use Arena

### ✅ **Perfect For:**
- **Web Servers**: Request-scoped allocations with automatic cleanup
- **Batch Processing**: Processing many items with temporary objects
- **High-Frequency Operations**: Reducing GC pressure in hot paths
- **Temporary Buffers**: String building, parsing, data transformation
- **Memory-Intensive Apps**: Applications with predictable allocation patterns

### ❌ **Not Ideal For:**
- Long-lived objects that outlive the arena
- Small programs with minimal allocation
- Cases where individual object deallocation is needed
- Memory-constrained environments where chunk overhead matters

## Memory Safety

- All allocations are properly aligned for the target architecture
- Panic on use-after-release for fail-fast debugging
- No buffer overruns - each allocation is bounds-checked
- Thread-safe operations when using `SafeArena`

## Configuration

### Chunk Sizes

Choose chunk size based on your allocation patterns:

- **Small chunks (4-16KB)**: Lower memory overhead, more chunks
- **Medium chunks (64KB, default)**: Good balance for most applications  
- **Large chunks (1MB+)**: Fewer chunks, better for large allocations

### Thread Safety

- Use `Arena` for single-threaded code (fastest)
- Use `SafeArena` for multi-threaded access (mutex overhead)
- Consider per-goroutine arenas to avoid contention

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
