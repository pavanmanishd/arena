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

## Performance

Benchmarks on typical hardware show significant performance improvements:

```
BenchmarkArenaAlloc-8           50000000    25.2 ns/op    0 B/op    0 allocs/op
BenchmarkBuiltinAlloc-8         20000000    65.4 ns/op   64 B/op    1 allocs/op
BenchmarkArenaAllocSlice-8      10000000   156.3 ns/op    0 B/op    0 allocs/op
BenchmarkBuiltinAllocSlice-8     5000000   312.7 ns/op  800 B/op    1 allocs/op
```

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

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Run tests (`go test -v ./...`)
4. Run benchmarks (`go test -bench=. -benchmem`)
5. Commit changes (`git commit -am 'Add amazing feature'`)
6. Push to branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Inspired by arena allocators in systems programming languages
- Built with Go's type safety and performance in mind
- Designed for production use in high-performance applications
