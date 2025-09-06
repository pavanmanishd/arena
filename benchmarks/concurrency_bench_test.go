package arena_test

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/pavanmanishd/arena"
)

// BenchmarkConcurrencyPatterns tests various concurrent usage patterns
func BenchmarkConcurrencyPatterns(b *testing.B) {

	// Sequential vs Parallel SafeArena usage
	b.Run("SafeArena_Sequential", func(b *testing.B) {
		s := arena.NewSafeArena(1024 * 1024)
		defer s.Release()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			s.AllocBytes(64)
			if i%1000 == 999 {
				s.Reset()
			}
		}
	})

	b.Run("SafeArena_Parallel", func(b *testing.B) {
		s := arena.NewSafeArena(1024 * 1024)
		defer s.Release()

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				s.AllocBytes(64)
				i++
				if i%1000 == 999 {
					s.Reset()
				}
			}
		})
	})

	// Arena per goroutine vs shared SafeArena
	b.Run("Arena_PerGoroutine", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			a := arena.NewArena(1024 * 1024)
			defer a.Release()

			i := 0
			for pb.Next() {
				a.AllocBytes(64)
				i++
				if i%1000 == 999 {
					a.Reset()
				}
			}
		})
	})

	// Standard allocation parallel baseline
	b.Run("Builtin_Parallel", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = make([]byte, 64)
			}
		})
	})

	// Different allocation sizes under contention
	sizes := []int{32, 128, 512}
	for _, size := range sizes {
		b.Run(fmt.Sprintf("SafeArena_Contention_%dB", size), func(b *testing.B) {
			s := arena.NewSafeArena(2 * 1024 * 1024)
			defer s.Release()

			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					s.AllocBytes(size)
				}
			})
		})

		b.Run(fmt.Sprintf("Arena_PerGoroutine_%dB", size), func(b *testing.B) {
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				a := arena.NewArena(2 * 1024 * 1024)
				defer a.Release()

				for pb.Next() {
					a.AllocBytes(size)
				}
			})
		})
	}
}

// BenchmarkSafeArenaOperations tests thread-safe operations performance
func BenchmarkSafeArenaOperations(b *testing.B) {
	s := arena.NewSafeArena(1024 * 1024)
	defer s.Release()

	// Pre-allocate some data for metrics tests
	for i := 0; i < 100; i++ {
		s.AllocBytes(1000)
	}

	b.Run("AllocBytes", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				s.AllocBytes(64)
			}
		})
	})

	b.Run("SafeAlloc", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				arena.SafeAlloc[int64](s)
			}
		})
	})

	b.Run("SafeAllocSlice", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				arena.SafeAllocSlice[int](s, 10)
			}
		})
	})

	b.Run("Metrics", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = s.Metrics()
			}
		})
	})

	b.Run("SizeInUse", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = s.SizeInUse()
			}
		})
	})
}

// BenchmarkConcurrentReset tests reset performance under concurrent access
func BenchmarkConcurrentReset(b *testing.B) {

	b.Run("SafeArena_ConcurrentAllocAndReset", func(b *testing.B) {
		s := arena.NewSafeArena(2 * 1024 * 1024)
		defer s.Release()

		b.ResetTimer()

		// Run allocations and resets concurrently
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				if i%1000 == 0 {
					s.Reset() // Occasional reset
				} else {
					s.AllocBytes(128)
				}
				i++
			}
		})
	})

	b.Run("Arena_PerGoroutine_Reset", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			a := arena.NewArena(2 * 1024 * 1024)
			defer a.Release()

			i := 0
			for pb.Next() {
				if i%1000 == 0 {
					a.Reset()
				} else {
					a.AllocBytes(128)
				}
				i++
			}
		})
	})
}

// BenchmarkScalability tests how performance scales with number of goroutines
func BenchmarkScalability(b *testing.B) {
	goroutineCounts := []int{1, 2, 4, 8, 16}

	for _, numGoroutines := range goroutineCounts {
		b.Run(fmt.Sprintf("SafeArena_%dGoroutines", numGoroutines), func(b *testing.B) {
			s := arena.NewSafeArena(4 * 1024 * 1024)
			defer s.Release()

			// Limit parallelism to test specific goroutine counts
			oldProcs := runtime.GOMAXPROCS(numGoroutines)
			defer runtime.GOMAXPROCS(oldProcs)

			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					s.AllocBytes(128)
				}
			})
		})

		b.Run(fmt.Sprintf("Arena_PerGoroutine_%dGoroutines", numGoroutines), func(b *testing.B) {
			oldProcs := runtime.GOMAXPROCS(numGoroutines)
			defer runtime.GOMAXPROCS(oldProcs)

			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				a := arena.NewArena(4 * 1024 * 1024)
				defer a.Release()

				for pb.Next() {
					a.AllocBytes(128)
				}
			})
		})

		b.Run(fmt.Sprintf("Builtin_%dGoroutines", numGoroutines), func(b *testing.B) {
			oldProcs := runtime.GOMAXPROCS(numGoroutines)
			defer runtime.GOMAXPROCS(oldProcs)

			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					_ = make([]byte, 128)
				}
			})
		})
	}
}
