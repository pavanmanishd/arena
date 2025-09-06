# Arena - High-Performance Memory Allocator for Go

A fast, chunked bump allocator (memory arena) implementation for Go that provides significant performance improvements for specific allocation patterns.

## Performance Overview

Arena delivers substantial performance improvements for batch allocation patterns:

- **4-12x faster** for small allocations (8-64 bytes)
- **240x faster** for batch allocations (100 objects)
- **1,100x faster** for buffer reuse patterns
- **54x faster** under high GC pressure
- **Zero GC pressure** for temporary allocations

## Benchmark Results

### Small Allocations (8-64 bytes)
```
Arena:   1.3 ns/op,     0 B/op,   0 allocs/op
Builtin: 5.7-16.2 ns/op, 8-64 B/op, 1 allocs/op
Result:  4-12x faster with zero GC impact
```

### Batch Processing (100 small objects)
```
Arena:   114 ns/op,     0 B/op,     0 allocs/op
Builtin: 27,307 ns/op,  6,402 B/op, 100 allocs/op
Result:  240x faster
```

### High GC Pressure (1000 objects)
```
Arena:   1,047 ns/op,   0 B/op,       0 allocs/op
Builtin: 56,389 ns/op,  128,000 B/op, 1,000 allocs/op
Result:  54x faster with zero GC impact
```

## When to Use Arena

### Excellent Performance Scenarios

**High-frequency small allocations (8-1024 bytes)**
- Web server request processing
- JSON parsing and serialization
- Protocol buffer handling
- Small struct allocations

**Request-scoped lifecycles**
- HTTP request handlers
- RPC call processing
- Database query processing
- API endpoint handling

**Batch processing with clear boundaries**
- Stream processing (process N events, reset)
- Database result processing
- File parsing operations
- Data transformation pipelines

**GC-sensitive applications**
- Low-latency trading systems
- Real-time game engines
- High-frequency data processing
- Systems requiring predictable latency

**Temporary object creation patterns**
- Graph algorithm processing
- Tree traversal operations
- Temporary buffer management
- Cache entry processing

### Poor Performance Scenarios

**Very tiny allocations (1-2 bytes)**
```
Arena:   3.2 ns/op (alignment overhead)
Builtin: 0.23 ns/op (optimized away)
Result:  Builtin is 14x faster
```

**Single large allocations (> chunk size)**
- Arena adds chunk management overhead
- No benefit from batch allocation patterns

**Long-lived objects (hours/days)**
- Arena keeps entire chunks alive
- Memory waste due to poor utilization

**Sparse allocation patterns**
- Poor chunk utilization
- Memory fragmentation issues

**Memory-constrained environments**
- Arena pre-allocates chunks
- Higher memory overhead

## Real-World Performance Analysis

### Web Server Request Handling
```go
func handleRequest(w http.ResponseWriter, r *http.Request) {
    arena := arena.NewArena(8192) // 8KB per request
    defer arena.Release()
    
    // All temporary allocations use arena
    headers := arena.AllocSlice[string](arena, 20)
    buffer := arena.AllocBytes(2048)
    tempData := arena.AllocSlice[int64](arena, 100)
    
    // Process request...
    // Arena automatically cleaned up
}
```
**Performance**: Arena 1,035 ns/op vs Builtin 5,472 ns/op

### Database Query Processing
```go
func processQueryResults(rows []DatabaseRow) {
    arena := arena.NewArena(512 * 1024) // 512KB
    defer arena.Release()
    
    // Allocate processing structures
    results := arena.AllocSlice[ProcessedRow](arena, len(rows))
    tempBuffers := arena.AllocSlice[[]byte](arena, 10)
    
    // Process data...
    arena.Reset() // O(1) cleanup for next batch
}
```
**Performance**: 188x faster than standard allocation

### JSON Document Parsing
```go
func parseJSONDocument(data []byte) *JSONObject {
    arena := arena.NewArena(256 * 1024) // 256KB
    defer arena.Release()
    
    // All parsed objects allocated in arena
    root := arena.Alloc[JSONObject](arena)
    root.Children = arena.AllocSlice[*JSONObject](arena, 100)
    
    // Parse document...
    return root // Valid until arena.Release()
}
```
**Performance**: Excellent for temporary parsing operations

## Project Structure

```
arena/
├── arena.go              # Core arena implementation
├── alloc.go              # Typed allocation functions
├── safe.go               # Thread-safe SafeArena
├── metrics.go            # Memory usage metrics
├── example_test.go       # Usage examples
├── tests/                # Comprehensive test suite
│   ├── go.mod           # Test module configuration
│   └── edge_cases_test.go # Edge cases and stress tests
└── benchmarks/           # Performance benchmarks
    ├── go.mod           # Benchmark module configuration
    ├── allocation_patterns_bench_test.go    # Basic allocation patterns
    ├── real_world_scenarios_bench_test.go   # Real-world usage scenarios
    ├── worst_case_scenarios_bench_test.go   # Poor performance scenarios
    └── concurrency_bench_test.go            # Concurrent usage patterns
```

## Comprehensive Testing

### Edge Cases Testing (`tests/edge_cases_test.go`)
- Zero/negative chunk sizes
- Integer overflow scenarios
- Memory alignment edge cases
- Use-after-release behavior
- Concurrent access patterns
- Memory corruption detection
- Boundary conditions

### Benchmark Scenarios

#### Allocation Patterns (`allocation_patterns_bench_test.go`)
Tests fundamental allocation performance across different sizes:

**Small Allocations (8-64 bytes)**
- Simulates pointer allocations, small structs, basic data types
- Tests alignment overhead impact
- Measures GC pressure differences

**Medium Allocations (128-1024 bytes)**
- Simulates typical struct allocations, small buffers
- Tests chunk utilization efficiency
- Measures reset operation benefits

**Large Allocations (2KB-64KB)**
- Tests chunk growth behavior
- Measures overhead of large object handling
- Identifies crossover points where arena becomes inefficient

**Typed Allocations**
- Tests generic allocation functions (Alloc[T], AllocSlice[T])
- Measures type-specific performance characteristics
- Compares zeroed vs uninitialized allocation performance

**Batch Allocations**
- Simulates request processing patterns (allocate many, reset)
- Measures cumulative allocation performance
- Tests GC pressure under batch scenarios

**GC Pressure Scenarios**
- **High GC Pressure**: Allocates 1000 objects repeatedly, triggering frequent GC
- **Low GC Pressure**: Single allocations with minimal GC interaction
- Measures GC pause impact on allocation performance

#### Real-World Scenarios (`real_world_scenarios_bench_test.go`)
Tests realistic usage patterns:

**Web Server Scenarios**
- **HTTP Request Handler**: Simulates typical request processing with headers, body buffers, temporary objects
- **Connection Pool**: Tests per-connection arena usage vs shared SafeArena
- Measures real-world web server allocation patterns

**Database Scenarios**
- **Query Result Processing**: Simulates processing 1000 database rows with temporary structures
- **Transaction Processing**: Tests batch transaction processing with metadata
- Measures database driver allocation patterns

**JSON Processing Scenarios**
- **Document Parsing**: Simulates parsing complex JSON with nested objects
- Tests temporary object creation during parsing
- Measures parser allocation efficiency

**Graph Algorithm Scenarios**
- **Graph Traversal**: Simulates BFS traversal with 1000 nodes
- Tests algorithm-heavy allocation patterns
- Measures temporary data structure performance

**Concurrent Workload Scenarios**
- **Worker Pool Pattern**: Tests arena-per-worker vs shared SafeArena
- Measures concurrent allocation efficiency
- Tests scalability under parallel workloads

#### Worst-Case Scenarios (`worst_case_scenarios_bench_test.go`)
Identifies when arena performs poorly:

**Tiny Allocations (1-2 bytes)**
- Tests alignment overhead impact
- Demonstrates when builtin allocation is superior
- Measures memory waste due to alignment

**Alternating Large/Small Allocations**
- Tests poor chunk utilization patterns
- Simulates fragmentation scenarios
- Measures memory waste in suboptimal usage

**Frequent Reset Operations**
- Tests reset overhead with many chunks
- Measures O(n) reset cost impact
- Identifies reset frequency limits

**Single Large Allocations**
- Tests arena overhead for single allocations
- Compares against direct allocation
- Measures chunk management overhead

**Sparse Allocation Patterns**
- Tests low chunk utilization scenarios
- Measures memory waste in sparse usage
- Identifies utilization thresholds

**Long-Lived Allocations**
- Tests memory retention with long-lived objects
- Measures chunk lifetime impact
- Demonstrates arena design limitations

**High Memory Pressure**
- Tests arena behavior under memory constraints
- Measures GC interaction under pressure
- Tests allocation failure scenarios

**Concurrent Contention**
- Tests SafeArena mutex contention
- Measures scalability limits
- Identifies contention bottlenecks

#### Concurrency Patterns (`concurrency_bench_test.go`)
Tests thread-safety and scalability:

**SafeArena Performance**
- Sequential vs parallel SafeArena usage
- Mutex overhead measurement
- Thread-safe operation costs

**Scalability Testing**
- Performance scaling with goroutine count (1, 2, 4, 8, 16)
- Arena-per-goroutine vs shared SafeArena comparison
- Contention measurement under load

**Concurrent Reset Operations**
- Tests reset safety under concurrent access
- Measures reset performance impact
- Tests allocation/reset race conditions

## Usage Examples

### Basic Usage
```go
package main

import "github.com/pavanmanishd/arena"

func main() {
    // Create arena with default chunk size (64KB)
    a := arena.NewArena(0)
    defer a.Release()
    
    // Allocate raw bytes
    buf := a.AllocBytes(1024)
    
    // Allocate typed values
    ptr := arena.Alloc[int](a)
    *ptr = 42
    
    // Allocate slices
    slice := arena.AllocSlice[int](a, 100)
    
    // Reset for reuse (O(1) operation)
    a.Reset()
}
```

### Thread-Safe Usage
```go
// For concurrent access, use SafeArena
safeArena := arena.NewSafeArena(0)
defer safeArena.Release()

// All operations are thread-safe
buf := safeArena.AllocBytes(1024)
ptr := arena.SafeAlloc[MyStruct](safeArena)
```

### Web Server Integration
```go
func requestHandler(w http.ResponseWriter, r *http.Request) {
    // Create arena for this request
    a := arena.NewArena(8192)
    defer a.Release()
    
    // Use arena for all temporary allocations
    requestData := a.AllocBytes(1024)
    responseBuffer := a.AllocBytes(2048)
    tempObjects := arena.AllocSlice[TempData](a, 50)
    
    // Process request using arena-allocated memory
    // Automatic cleanup when function returns
}
```

## Performance Configuration

### Optimal Settings

**Chunk Size Selection**
- **General use**: 64KB (default)
- **Small objects**: 4-16KB
- **Large objects**: 256KB-1MB
- **Memory constrained**: 4-8KB

**Reset Frequency**
- **Web requests**: After each request
- **Batch processing**: Every 100-1000 operations
- **Stream processing**: After each batch

**Thread Safety**
- **Single goroutine**: Use `Arena` (faster)
- **Multiple goroutines**: Use `SafeArena` (10-20% overhead)
- **High contention**: Use per-goroutine arenas

**Memory Utilization**
- **Target**: >70% chunk utilization
- **Monitor**: Use `arena.Metrics()` for tracking
- **Optimize**: Adjust chunk size based on allocation patterns

### Monitoring
```go
metrics := arena.Metrics()
fmt.Printf("Utilization: %.2f%%\n", metrics.Utilization * 100)
fmt.Printf("Memory in use: %d bytes\n", metrics.SizeInUse)
fmt.Printf("Total capacity: %d bytes\n", metrics.Capacity)
fmt.Printf("Number of chunks: %d\n", metrics.NumChunks)
```

## Testing and Benchmarking

### Running Tests

Tests are organized in the `tests/` directory:

```bash
# Run all tests (from root directory)
go test -v ./...

# Run edge case tests specifically
cd tests && go test -v -run=TestEdgeCases

# Run all tests in tests directory
cd tests && go test -v

# Run tests with coverage
cd tests && go test -v -cover
```

### Running Benchmarks

Benchmarks are organized in the `benchmarks/` directory as a separate module:

```bash
# Navigate to benchmarks directory first
cd benchmarks

# Run all benchmarks
go test -bench=. -benchmem

# Run specific benchmark categories
go test -bench=BenchmarkSmallAllocations -benchmem
go test -bench=BenchmarkWebServerScenarios -benchmem
go test -bench=BenchmarkWorstCaseScenarios -benchmem
go test -bench=BenchmarkConcurrencyPatterns -benchmem

# Compare arena vs builtin performance
go test -bench=BenchmarkBatchAllocations -benchmem

# Run with multiple iterations for more accurate results
go test -bench=BenchmarkSmallAllocations -benchmem -count=3
```

### Benchmark Categories

- **`allocation_patterns_bench_test.go`**: Small, medium, large allocations, typed allocations, batch patterns
- **`real_world_scenarios_bench_test.go`**: Web servers, databases, JSON processing, graph algorithms
- **`worst_case_scenarios_bench_test.go`**: Scenarios where arena performs poorly
- **`concurrency_bench_test.go`**: Thread safety, scalability, concurrent patterns

## Architecture

The arena uses a chunked bump allocator design:

1. **Chunks**: Large contiguous memory blocks (default 64KB)
2. **Bump allocation**: Sequential allocation within chunks
3. **Alignment**: All allocations aligned to pointer size
4. **Growth**: New chunks allocated when current chunk fills
5. **Reset**: O(1) operation that resets all chunk offsets
6. **Release**: Frees all chunks and makes arena unusable

## Memory Layout

```
Arena
├── Chunk 1 (64KB)
│   ├── [Allocation 1] [Padding] [Allocation 2] [Padding] ...
│   └── [Free Space]
├── Chunk 2 (64KB)
│   ├── [Large Allocation] 
│   └── [Free Space]
└── Chunk 3 (Custom Size)
    └── [Very Large Allocation]
```

## Performance Summary

| Scenario | Arena Performance | Standard Go | Speedup | Best Use Case |
|----------|------------------|-------------|---------|---------------|
| Small allocs (8-64B) | 1.3 ns/op | 5.7-16.2 ns/op | **4-12x** | Frequent small objects |
| Batch allocs (100x) | 114 ns/op | 27,307 ns/op | **240x** | Request processing |
| Struct allocs (50x) | 143 ns/op | 27,000 ns/op | **188x** | Object creation |
| Buffer reuse | 50.5 ns/op | 56,000 ns/op | **1,100x** | Temporary buffers |
| High GC pressure | 1,047 ns/op | 56,389 ns/op | **54x** | GC-sensitive apps |
| Tiny allocs (1-2B) | 3.2 ns/op | 0.23 ns/op | **0.07x** | Avoid |

## Important Considerations

### Memory Safety
- Allocated memory is only valid while arena exists
- No individual deallocation - use `Reset()` or `Release()` for cleanup
- Avoid storing arena-allocated pointers beyond arena lifetime

### Performance Trade-offs
- **Excellent**: Batch allocations, temporary objects, request-scoped data
- **Poor**: Tiny allocations, single large allocations, long-lived objects
- **Memory overhead**: Pre-allocated chunks, alignment padding

### Thread Safety
- `Arena`: Not thread-safe (faster, single goroutine)
- `SafeArena`: Thread-safe (mutex overhead, multiple goroutines)

## Conclusion

Arena is a specialized, high-performance memory allocator that excels in specific scenarios:

- **Use for**: High-frequency small allocations, request-scoped lifecycles, batch processing
- **Avoid for**: Tiny allocations, long-lived objects, single large allocations
- **Best fit**: Web servers, real-time systems, data processing pipelines

The performance benefits are substantial (4-1100x faster) when used correctly, making it an excellent tool for performance-critical Go applications.

## License

[License details here]

## Contributing

[Contributing guidelines here]