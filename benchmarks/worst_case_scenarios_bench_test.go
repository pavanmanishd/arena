package arena_test

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/pavanmanishd/arena"
)

// BenchmarkWorstCaseScenarios tests scenarios where arena might perform poorly
// These benchmarks help identify when NOT to use arena allocation
func BenchmarkWorstCaseScenarios(b *testing.B) {

	// Scenario 1: Many tiny allocations (high alignment overhead)
	// Arena has to align every allocation to pointer size, wasting space for tiny allocations
	b.Run("TinyAllocations", func(b *testing.B) {
		b.Run("Arena_1B", func(b *testing.B) {
			a := arena.NewArena(64 * 1024)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				a.AllocBytes(1)
				if i%10000 == 9999 {
					a.Reset()
				}
			}
		})

		b.Run("Builtin_1B", func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_ = make([]byte, 1)
			}
		})

		b.Run("Arena_2B", func(b *testing.B) {
			a := arena.NewArena(64 * 1024)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				a.AllocBytes(2)
				if i%10000 == 9999 {
					a.Reset()
				}
			}
		})

		b.Run("Builtin_2B", func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_ = make([]byte, 2)
			}
		})
	})

	// Scenario 2: Alternating large and small allocations (poor chunk utilization)
	// This creates fragmentation where large allocations force new chunks but leave small gaps
	b.Run("AlternatingLargeSmall", func(b *testing.B) {
		b.Run("Arena", func(b *testing.B) {
			a := arena.NewArena(8192)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				if i%2 == 0 {
					a.AllocBytes(7000) // Large allocation (forces new chunk)
				} else {
					a.AllocBytes(100) // Small allocation (new chunk needed due to fragmentation)
				}
				if i%100 == 99 {
					a.Reset()
				}
			}
		})

		b.Run("Builtin", func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				if i%2 == 0 {
					_ = make([]byte, 7000)
				} else {
					_ = make([]byte, 100)
				}
			}
		})
	})

	// Scenario 3: Very frequent resets (overhead of reset operation)
	// Reset has to iterate through all chunks, so frequent resets add overhead
	b.Run("FrequentReset", func(b *testing.B) {
		a := arena.NewArena(64 * 1024)
		defer a.Release()

		// Create multiple chunks first
		for i := 0; i < 10; i++ {
			a.AllocBytes(8192)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			a.AllocBytes(64)
			a.Reset() // Reset after every allocation
		}
	})

	// Scenario 4: Single large allocations (arena overhead without benefit)
	// For single large allocations, arena adds overhead without providing benefits
	b.Run("SingleLargeAllocations", func(b *testing.B) {
		sizes := []int{64 * 1024, 256 * 1024, 1024 * 1024} // 64KB, 256KB, 1MB

		for _, size := range sizes {
			b.Run(fmt.Sprintf("Arena_%dKB", size/1024), func(b *testing.B) {
				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					a := arena.NewArena(size * 2) // Chunk size larger than allocation
					a.AllocBytes(size)
					a.Release()
				}
			})

			b.Run(fmt.Sprintf("Builtin_%dKB", size/1024), func(b *testing.B) {
				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					_ = make([]byte, size)
				}
			})
		}
	})

	// Scenario 5: Sparse allocation patterns (poor memory utilization)
	// Allocating much less than chunk size wastes memory
	b.Run("SparseAllocations", func(b *testing.B) {
		b.Run("Arena_LowUtilization", func(b *testing.B) {
			a := arena.NewArena(64 * 1024) // 64KB chunks
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				// Only use 1KB of each 64KB chunk
				a.AllocBytes(1024)
				// Force new chunk by exceeding remaining space conceptually
				// (this simulates poor allocation patterns)
				if i%50 == 49 {
					a.Reset()
				}
			}
		})

		b.Run("Builtin", func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_ = make([]byte, 1024)
			}
		})
	})

	// Scenario 6: Long-lived allocations (arena keeps entire chunks alive)
	// Arena is designed for short-lived allocations; long-lived ones waste memory
	b.Run("LongLivedAllocations", func(b *testing.B) {
		b.Run("Arena", func(b *testing.B) {
			// Simulate keeping allocations alive for a long time
			var arenas []*arena.Arena
			var ptrs []*int64

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				a := arena.NewArena(4096)
				ptr := arena.Alloc[int64](a)
				*ptr = int64(i)

				// Keep references alive (simulating long-lived data)
				arenas = append(arenas, a)
				ptrs = append(ptrs, ptr)

				// Clean up periodically to prevent memory explosion
				if len(arenas) > 100 {
					for _, arena := range arenas[:50] {
						arena.Release()
					}
					arenas = arenas[50:]
					ptrs = ptrs[50:]
				}
			}

			// Clean up remaining
			for _, arena := range arenas {
				arena.Release()
			}
		})

		b.Run("Builtin", func(b *testing.B) {
			var ptrs []*int64

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				ptr := new(int64)
				*ptr = int64(i)

				// Keep references alive
				ptrs = append(ptrs, ptr)

				// Clean up periodically
				if len(ptrs) > 100 {
					ptrs = ptrs[50:]
				}
			}
		})
	})

	// Scenario 7: High memory pressure (frequent GC with arena overhead)
	// When memory is constrained, arena's chunk allocation can trigger more GC
	b.Run("HighMemoryPressure", func(b *testing.B) {
		// Force GC to run more frequently
		oldGCPercent := runtime.GOMAXPROCS(0)
		runtime.GC()
		defer func() {
			runtime.GOMAXPROCS(oldGCPercent)
		}()

		b.Run("Arena", func(b *testing.B) {
			a := arena.NewArena(1024 * 1024)
			defer a.Release()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Allocate large amounts of memory
				for j := 0; j < 100; j++ {
					a.AllocBytes(10240) // 10KB each
				}
				a.Reset()

				// Force GC occasionally
				if i%10 == 9 {
					runtime.GC()
				}
			}
		})

		b.Run("Builtin", func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				// Allocate large amounts of memory
				buffers := make([][]byte, 100)
				for j := 0; j < 100; j++ {
					buffers[j] = make([]byte, 10240)
				}

				// Force GC occasionally
				if i%10 == 9 {
					runtime.GC()
				}
			}
		})
	})

	// Scenario 8: Concurrent access overhead (SafeArena mutex contention)
	// SafeArena uses mutex, which can become a bottleneck under high contention
	b.Run("HighConcurrentContention", func(b *testing.B) {
		s := arena.NewSafeArena(1024 * 1024)
		defer s.Release()

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				// High contention on single SafeArena
				s.AllocBytes(64)
			}
		})
	})

	// Scenario 9: Allocation sizes close to chunk size (poor utilization)
	// Allocating close to chunk size wastes the remaining space
	b.Run("NearChunkSizeAllocations", func(b *testing.B) {
		chunkSize := 8192

		b.Run("Arena", func(b *testing.B) {
			a := arena.NewArena(chunkSize)
			defer a.Release()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Allocate 90% of chunk size, wasting 10%
				a.AllocBytes(int(float64(chunkSize) * 0.9))
				if i%100 == 99 {
					a.Reset()
				}
			}
		})

		b.Run("Builtin", func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_ = make([]byte, int(float64(chunkSize)*0.9))
			}
		})
	})
}
