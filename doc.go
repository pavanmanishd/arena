// Package arena implements a chunked bump allocator (memory arena) for Go.
//
// # Overview
//
// An arena allocator is a fast memory allocation strategy that allocates
// memory in large chunks and then hands out portions of those chunks
// on demand. This is particularly useful for:
//
//   - Request-scoped allocations in web servers
//   - Temporary object allocation with batch cleanup
//   - Reducing garbage collection pressure
//   - High-performance applications requiring predictable allocation patterns
//
// # Basic Usage
//
//	arena := arena.NewArena(0) // Use default chunk size
//	defer arena.Release()      // Clean up when done
//
//	// Allocate raw bytes
//	buf := arena.AllocBytes(1024)
//
//	// Allocate typed values
//	ptr := arena.Alloc[MyStruct](arena)
//	slice := arena.AllocSlice[int](arena, 100)
//
//	// Reset for reuse (O(1) operation)
//	arena.Reset()
//
// # Thread Safety
//
// The basic Arena type is not thread-safe. For concurrent access, use SafeArena:
//
//	safeArena := arena.NewSafeArena(0)
//	defer safeArena.Release()
//
//	// All operations are thread-safe
//	buf := safeArena.AllocBytes(1024)
//	ptr := arena.SafeAlloc[MyStruct](safeArena)
//
// # Memory Layout
//
// The arena allocates memory in chunks (default 64KB). When a chunk fills up,
// a new chunk is allocated. Memory within chunks is allocated sequentially
// with proper alignment for the target architecture.
//
// # Performance Characteristics
//
//   - Allocation: O(1) amortized
//   - Reset: O(number of chunks) - typically very fast
//   - Release: O(1)
//   - Memory overhead: Minimal (just chunk metadata)
//
// # Important Notes
//
//   - Allocated memory is only valid while the arena exists
//   - No individual deallocation - use Reset() or Release() for bulk cleanup
//   - Memory is not automatically zeroed unless using Alloc() or AllocZeroed()
//   - Proper alignment is maintained for all allocations
//
// # Metrics and Monitoring
//
// The arena provides detailed metrics for monitoring memory usage:
//
//	metrics := arena.Metrics()
//	fmt.Printf("Utilization: %.2f%%\n", metrics.Utilization * 100)
//	fmt.Printf("Memory in use: %d bytes\n", metrics.SizeInUse)
//	fmt.Printf("Total capacity: %d bytes\n", metrics.Capacity)
package arena
