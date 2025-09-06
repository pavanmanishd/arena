package arena

import (
	"runtime"
	"testing"
)

// BenchmarkRealisticUsage tests scenarios where arena should excel
func BenchmarkRealisticUsage(b *testing.B) {

	// Test 1: Many small allocations with periodic cleanup
	b.Run("ManySmallAllocs/Arena", func(b *testing.B) {
		a := NewArena(64 * 1024) // 64KB chunks
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

	b.Run("ManySmallAllocs/Builtin", func(b *testing.B) {
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

	// Test 2: Struct allocation patterns
	type TestStruct struct {
		ID   int64
		Data [56]byte // Total 64 bytes
	}

	b.Run("StructAllocs/Arena", func(b *testing.B) {
		a := NewArena(64 * 1024)
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			// Allocate 50 structs
			for j := 0; j < 50; j++ {
				s := Alloc[TestStruct](a)
				s.ID = int64(j)
			}
			a.Reset()
		}
	})

	b.Run("StructAllocs/Builtin", func(b *testing.B) {
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

	// Test 3: Buffer reuse pattern
	b.Run("BufferReuse/Arena", func(b *testing.B) {
		a := NewArena(1024 * 1024) // 1MB
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

	b.Run("BufferReuse/Builtin", func(b *testing.B) {
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

	// Test 4: No GC pressure test
	b.Run("NoGCPressure/Arena", func(b *testing.B) {
		a := NewArena(1024 * 1024)

		// Force GC before test
		runtime.GC()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			a.AllocBytes(128)
			if i%1000 == 999 {
				a.Reset()
			}
		}
	})

	b.Run("NoGCPressure/Builtin", func(b *testing.B) {
		// Force GC before test
		runtime.GC()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = make([]byte, 128)
		}
	})
}
