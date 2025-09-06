package arena_test

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/pavanmanishd/arena"
)

// BenchmarkSmallAllocations tests small allocation patterns (8-64 bytes)
// These are common for small objects, pointers, and basic data structures
func BenchmarkSmallAllocations(b *testing.B) {
	sizes := []int{8, 16, 32, 64}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Arena_%dB", size), func(b *testing.B) {
			a := arena.NewArena(64 * 1024)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				a.AllocBytes(size)
				if i%1000 == 999 {
					a.Reset()
				}
			}
		})

		b.Run(fmt.Sprintf("Builtin_%dB", size), func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_ = make([]byte, size)
			}
		})
	}
}

// BenchmarkMediumAllocations tests medium allocation patterns (128-1024 bytes)
// These are common for structs, small buffers, and data processing
func BenchmarkMediumAllocations(b *testing.B) {
	sizes := []int{128, 256, 512, 1024}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Arena_%dB", size), func(b *testing.B) {
			a := arena.NewArena(64 * 1024)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				a.AllocBytes(size)
				if i%500 == 499 {
					a.Reset()
				}
			}
		})

		b.Run(fmt.Sprintf("Builtin_%dB", size), func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_ = make([]byte, size)
			}
		})
	}
}

// BenchmarkLargeAllocations tests large allocation patterns (2KB-64KB)
// These are less common but important for buffers and large data structures
func BenchmarkLargeAllocations(b *testing.B) {
	sizes := []int{2048, 8192, 32768, 65536}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Arena_%dB", size), func(b *testing.B) {
			a := arena.NewArena(128 * 1024)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				a.AllocBytes(size)
				if i%100 == 99 {
					a.Reset()
				}
			}
		})

		b.Run(fmt.Sprintf("Builtin_%dB", size), func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_ = make([]byte, size)
			}
		})
	}
}

// BenchmarkTypedAllocations tests allocation of various Go types
func BenchmarkTypedAllocations(b *testing.B) {

	// Basic types
	b.Run("BasicTypes", func(b *testing.B) {
		b.Run("Arena_int", func(b *testing.B) {
			a := arena.NewArena(64 * 1024)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				arena.Alloc[int](a)
				if i%1000 == 999 {
					a.Reset()
				}
			}
		})

		b.Run("Builtin_int", func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_ = new(int)
			}
		})

		b.Run("Arena_int64", func(b *testing.B) {
			a := arena.NewArena(64 * 1024)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				arena.Alloc[int64](a)
				if i%1000 == 999 {
					a.Reset()
				}
			}
		})

		b.Run("Builtin_int64", func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_ = new(int64)
			}
		})
	})

	// Struct allocations
	type SmallStruct struct {
		A int32
		B int32
	}

	type MediumStruct struct {
		A int64
		B int64
		C int64
		D int64
		E [32]byte
	}

	type LargeStruct struct {
		A [256]byte
		B int64
		C string
		D []int
	}

	b.Run("Structs", func(b *testing.B) {
		b.Run("Arena_SmallStruct", func(b *testing.B) {
			a := arena.NewArena(64 * 1024)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				arena.Alloc[SmallStruct](a)
				if i%1000 == 999 {
					a.Reset()
				}
			}
		})

		b.Run("Builtin_SmallStruct", func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_ = new(SmallStruct)
			}
		})

		b.Run("Arena_MediumStruct", func(b *testing.B) {
			a := arena.NewArena(64 * 1024)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				arena.Alloc[MediumStruct](a)
				if i%500 == 499 {
					a.Reset()
				}
			}
		})

		b.Run("Builtin_MediumStruct", func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_ = new(MediumStruct)
			}
		})

		b.Run("Arena_LargeStruct", func(b *testing.B) {
			a := arena.NewArena(128 * 1024)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				arena.Alloc[LargeStruct](a)
				if i%200 == 199 {
					a.Reset()
				}
			}
		})

		b.Run("Builtin_LargeStruct", func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_ = new(LargeStruct)
			}
		})
	})
}

// BenchmarkSliceAllocations tests slice allocation patterns
func BenchmarkSliceAllocations(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Arena_Slice_%d", size), func(b *testing.B) {
			a := arena.NewArena(1024 * 1024)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				arena.AllocSlice[int](a, size)
				if i%100 == 99 {
					a.Reset()
				}
			}
		})

		b.Run(fmt.Sprintf("Arena_SliceZeroed_%d", size), func(b *testing.B) {
			a := arena.NewArena(1024 * 1024)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				arena.AllocSliceZeroed[int](a, size)
				if i%100 == 99 {
					a.Reset()
				}
			}
		})

		b.Run(fmt.Sprintf("Builtin_Slice_%d", size), func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_ = make([]int, size)
			}
		})
	}
}

// BenchmarkBatchAllocations tests scenarios with many allocations followed by reset
// This simulates request processing, batch operations, etc.
func BenchmarkBatchAllocations(b *testing.B) {

	// Many small allocations with periodic cleanup
	b.Run("ManySmallAllocs", func(b *testing.B) {
		b.Run("Arena", func(b *testing.B) {
			a := arena.NewArena(64 * 1024) // 64KB chunks
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				// Allocate 100 small objects
				for j := 0; j < 100; j++ {
					a.AllocBytes(64)
				}
				// Reset every 100 allocations (simulates request cleanup)
				a.Reset()
			}
		})

		b.Run("Builtin", func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				// Allocate 100 small objects
				objects := make([][]byte, 100)
				for j := 0; j < 100; j++ {
					objects[j] = make([]byte, 64)
				}
				// Force GC to clean up (simulates request cleanup)
				if i%10 == 0 {
					runtime.GC()
				}
			}
		})
	})

	// Struct allocation patterns
	type TestStruct struct {
		ID   int64
		Data [56]byte // Total 64 bytes
	}

	b.Run("StructAllocs", func(b *testing.B) {
		b.Run("Arena", func(b *testing.B) {
			a := arena.NewArena(64 * 1024)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				// Allocate 50 structs
				for j := 0; j < 50; j++ {
					s := arena.Alloc[TestStruct](a)
					s.ID = int64(j)
				}
				a.Reset()
			}
		})

		b.Run("Builtin", func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				// Allocate 50 structs
				structs := make([]*TestStruct, 50)
				for j := 0; j < 50; j++ {
					structs[j] = &TestStruct{ID: int64(j)}
				}
				if i%10 == 0 {
					runtime.GC()
				}
			}
		})
	})

	// Buffer reuse pattern
	b.Run("BufferReuse", func(b *testing.B) {
		b.Run("Arena", func(b *testing.B) {
			a := arena.NewArena(1024 * 1024) // 1MB
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				// Simulate processing 10 items with temporary buffers
				for j := 0; j < 10; j++ {
					buf1 := a.AllocBytes(1024)
					buf2 := a.AllocBytes(2048)
					buf3 := a.AllocBytes(512)

					// Simulate work
					buf1[0] = byte(j)
					buf2[0] = byte(j)
					buf3[0] = byte(j)
				}
				// O(1) cleanup
				a.Reset()
			}
		})

		b.Run("Builtin", func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				// Simulate processing 10 items with temporary buffers
				buffers := make([][]byte, 30) // 3 buffers per item
				for j := 0; j < 10; j++ {
					buffers[j*3] = make([]byte, 1024)
					buffers[j*3+1] = make([]byte, 2048)
					buffers[j*3+2] = make([]byte, 512)

					// Simulate work
					buffers[j*3][0] = byte(j)
					buffers[j*3+1][0] = byte(j)
					buffers[j*3+2][0] = byte(j)
				}
				if i%5 == 0 {
					runtime.GC()
				}
			}
		})
	})
}

// BenchmarkGCPressure measures GC impact
func BenchmarkGCPressure(b *testing.B) {

	b.Run("HighGCPressure", func(b *testing.B) {
		b.Run("Arena", func(b *testing.B) {
			a := arena.NewArena(1024 * 1024)

			// Force GC before test
			runtime.GC()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Allocate many objects
				for j := 0; j < 1000; j++ {
					a.AllocBytes(128)
				}
				a.Reset() // O(1) cleanup
			}
		})

		b.Run("Builtin", func(b *testing.B) {
			// Force GC before test
			runtime.GC()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Allocate many objects
				objects := make([][]byte, 1000)
				for j := 0; j < 1000; j++ {
					objects[j] = make([]byte, 128)
				}
				// Let GC clean up
			}
		})
	})

	b.Run("LowGCPressure", func(b *testing.B) {
		b.Run("Arena", func(b *testing.B) {
			a := arena.NewArena(64 * 1024)

			runtime.GC()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				a.AllocBytes(64)
				if i%10000 == 9999 {
					a.Reset()
				}
			}
		})

		b.Run("Builtin", func(b *testing.B) {
			runtime.GC()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = make([]byte, 64)
			}
		})
	})
}
