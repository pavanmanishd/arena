package arena

import (
	"fmt"
	"testing"
	"unsafe"
)

func TestNewArena(t *testing.T) {
	tests := []struct {
		name      string
		chunkSize int
		expected  int
	}{
		{"default chunk size", 0, DefaultChunkSize},
		{"negative chunk size", -1, DefaultChunkSize},
		{"custom chunk size", 8192, 8192},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := NewArena(tt.chunkSize)
			if a.chunkSize != tt.expected {
				t.Errorf("NewArena(%d) chunk size = %d, want %d", tt.chunkSize, a.chunkSize, tt.expected)
			}
			if len(a.chunks) != 1 {
				t.Errorf("NewArena(%d) chunks = %d, want 1", tt.chunkSize, len(a.chunks))
			}
		})
	}
}

func TestArenaAllocBytes(t *testing.T) {
	a := NewArena(1024)

	// Test normal allocation
	b1 := a.AllocBytes(100)
	if len(b1) != 100 {
		t.Errorf("AllocBytes(100) length = %d, want 100", len(b1))
	}

	// Test zero allocation
	b2 := a.AllocBytes(0)
	if b2 != nil {
		t.Errorf("AllocBytes(0) = %v, want nil", b2)
	}

	// Test negative allocation
	b3 := a.AllocBytes(-1)
	if b3 != nil {
		t.Errorf("AllocBytes(-1) = %v, want nil", b3)
	}

	// Test allocation that forces chunk growth
	b4 := a.AllocBytes(2000) // Larger than initial chunk
	if len(b4) != 2000 {
		t.Errorf("AllocBytes(2000) length = %d, want 2000", len(b4))
	}
	if a.NumChunks() != 2 {
		t.Errorf("NumChunks after large allocation = %d, want 2", a.NumChunks())
	}
}

func TestArenaEnsureCapacity(t *testing.T) {
	a := NewArena(1024)
	initialChunks := a.NumChunks()

	// Ensure capacity within current chunk
	a.EnsureCapacity(100)
	if a.NumChunks() != initialChunks {
		t.Errorf("EnsureCapacity(100) changed chunk count")
	}

	// Ensure capacity that requires new chunk
	a.EnsureCapacity(2000)
	if a.NumChunks() != initialChunks+1 {
		t.Errorf("EnsureCapacity(2000) chunks = %d, want %d", a.NumChunks(), initialChunks+1)
	}
}

func TestArenaReset(t *testing.T) {
	a := NewArena(1024)

	// Allocate some data
	a.AllocBytes(100)
	a.AllocBytes(200)

	initialSizeInUse := a.SizeInUse()
	if initialSizeInUse == 0 {
		t.Error("Expected non-zero size in use after allocations")
	}

	// Reset and check
	a.Reset()
	if a.SizeInUse() != 0 {
		t.Errorf("SizeInUse after Reset() = %d, want 0", a.SizeInUse())
	}

	// Verify chunks are still there
	if a.NumChunks() == 0 {
		t.Error("Expected chunks to remain after Reset()")
	}
}

func TestArenaRelease(t *testing.T) {
	a := NewArena(1024)
	a.AllocBytes(100)

	a.Release()

	if a.chunks != nil {
		t.Error("Expected chunks to be nil after Release()")
	}

	// Test panic on use after release
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic on use after Release()")
		}
	}()
	a.AllocBytes(100)
}

func TestAlignPtr(t *testing.T) {
	ptrSize := unsafe.Sizeof(uintptr(0))

	tests := []struct {
		input    uintptr
		expected uintptr
	}{
		{0, 0},
		{1, ptrSize},
		{ptrSize, ptrSize},
		{ptrSize + 1, ptrSize * 2},
	}

	for _, tt := range tests {
		result := alignPtr(tt.input)
		if result != tt.expected {
			t.Errorf("alignPtr(%d) = %d, want %d", tt.input, result, tt.expected)
		}
	}
}

func BenchmarkArenaAllocBytes(b *testing.B) {
	a := NewArena(1024 * 1024) // 1MB chunks
	sizes := []int{8, 64, 256, 1024}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size-%d", size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				a.AllocBytes(size)
				if i%1000 == 999 { // Reset periodically to avoid growing too much
					a.Reset()
				}
			}
		})
	}
}

func BenchmarkArenaVsBuiltin(b *testing.B) {
	b.Run("arena", func(b *testing.B) {
		a := NewArena(1024 * 1024)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			a.AllocBytes(64)
			if i%1000 == 999 {
				a.Reset()
			}
		}
	})

	b.Run("builtin", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = make([]byte, 64)
		}
	})
}
