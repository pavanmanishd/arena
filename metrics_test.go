package arena

import (
	"testing"
)

func TestArenaMetrics(t *testing.T) {
	a := NewArena(1024)

	// Test initial state
	if a.SizeInUse() != 0 {
		t.Errorf("Initial SizeInUse = %d, want 0", a.SizeInUse())
	}
	if a.NumChunks() != 1 {
		t.Errorf("Initial NumChunks = %d, want 1", a.NumChunks())
	}
	if a.Capacity() == 0 {
		t.Error("Initial Capacity should be > 0")
	}
	if a.ChunkSize() != 1024 {
		t.Errorf("ChunkSize = %d, want 1024", a.ChunkSize())
	}
	if a.Utilization() != 0 {
		t.Errorf("Initial Utilization = %f, want 0", a.Utilization())
	}

	// Allocate some data
	a.AllocBytes(100)
	a.AllocBytes(200)

	sizeInUse := a.SizeInUse()
	if sizeInUse == 0 {
		t.Error("SizeInUse should be > 0 after allocations")
	}

	utilization := a.Utilization()
	if utilization <= 0 || utilization > 1 {
		t.Errorf("Utilization = %f, want 0 < x <= 1", utilization)
	}

	// Force chunk growth
	a.AllocBytes(2000) // Larger than chunk size
	if a.NumChunks() != 2 {
		t.Errorf("NumChunks after growth = %d, want 2", a.NumChunks())
	}

	capacity := a.Capacity()
	if capacity <= 1024 {
		t.Errorf("Capacity after growth = %d, want > 1024", capacity)
	}

	// Test metrics snapshot
	metrics := a.Metrics()
	if metrics.SizeInUse != a.SizeInUse() {
		t.Errorf("Metrics.SizeInUse = %d, want %d", metrics.SizeInUse, a.SizeInUse())
	}
	if metrics.Capacity != a.Capacity() {
		t.Errorf("Metrics.Capacity = %d, want %d", metrics.Capacity, a.Capacity())
	}
	if metrics.NumChunks != a.NumChunks() {
		t.Errorf("Metrics.NumChunks = %d, want %d", metrics.NumChunks, a.NumChunks())
	}
	if metrics.ChunkSize != a.ChunkSize() {
		t.Errorf("Metrics.ChunkSize = %d, want %d", metrics.ChunkSize, a.ChunkSize())
	}
	if metrics.Utilization != a.Utilization() {
		t.Errorf("Metrics.Utilization = %f, want %f", metrics.Utilization, a.Utilization())
	}
}

func TestArenaMetricsAfterReset(t *testing.T) {
	a := NewArena(1024)

	// Allocate and verify
	a.AllocBytes(500)
	if a.SizeInUse() == 0 {
		t.Error("Expected non-zero SizeInUse before reset")
	}
	if a.Utilization() == 0 {
		t.Error("Expected non-zero Utilization before reset")
	}

	// Reset and verify
	a.Reset()
	if a.SizeInUse() != 0 {
		t.Errorf("SizeInUse after Reset = %d, want 0", a.SizeInUse())
	}
	if a.Utilization() != 0 {
		t.Errorf("Utilization after Reset = %f, want 0", a.Utilization())
	}
	// Chunks should remain
	if a.NumChunks() == 0 {
		t.Error("NumChunks should not be 0 after Reset")
	}
	if a.Capacity() == 0 {
		t.Error("Capacity should not be 0 after Reset")
	}
}

func TestArenaMetricsAfterRelease(t *testing.T) {
	a := NewArena(1024)
	a.AllocBytes(100)

	a.Release()

	if a.SizeInUse() != 0 {
		t.Errorf("SizeInUse after Release = %d, want 0", a.SizeInUse())
	}
	if a.NumChunks() != 0 {
		t.Errorf("NumChunks after Release = %d, want 0", a.NumChunks())
	}
	if a.Capacity() != 0 {
		t.Errorf("Capacity after Release = %d, want 0", a.Capacity())
	}
	if a.Utilization() != 0 {
		t.Errorf("Utilization after Release = %f, want 0", a.Utilization())
	}
}

func TestSafeArenaMetrics(t *testing.T) {
	s := NewSafeArena(2048)

	// Test that SafeArena metrics match underlying Arena
	s.AllocBytes(300)

	if s.SizeInUse() == 0 {
		t.Error("SafeArena SizeInUse should be > 0")
	}
	if s.NumChunks() == 0 {
		t.Error("SafeArena NumChunks should be > 0")
	}
	if s.Capacity() == 0 {
		t.Error("SafeArena Capacity should be > 0")
	}
	if s.ChunkSize() != 2048 {
		t.Errorf("SafeArena ChunkSize = %d, want 2048", s.ChunkSize())
	}

	utilization := s.Utilization()
	if utilization <= 0 || utilization > 1 {
		t.Errorf("SafeArena Utilization = %f, want 0 < x <= 1", utilization)
	}

	// Test metrics snapshot for SafeArena
	metrics := s.Metrics()
	if metrics.ChunkSize != 2048 {
		t.Errorf("SafeArena Metrics.ChunkSize = %d, want 2048", metrics.ChunkSize)
	}
	if metrics.SizeInUse == 0 {
		t.Error("SafeArena Metrics.SizeInUse should be > 0")
	}
}

func TestUtilizationEdgeCases(t *testing.T) {
	// Test with released arena
	a := NewArena(1024)
	a.Release()
	if a.Utilization() != 0 {
		t.Errorf("Released arena Utilization = %f, want 0", a.Utilization())
	}

	// Test with arena that has capacity but no allocations
	a2 := NewArena(1024)
	if a2.Utilization() != 0 {
		t.Errorf("Empty arena Utilization = %f, want 0", a2.Utilization())
	}

	// Test with full utilization
	a3 := NewArena(100)
	a3.AllocBytes(a3.Capacity()) // Allocate all available space
	util := a3.Utilization()
	if util < 0.9 { // Should be close to 1.0, allowing for alignment overhead
		t.Errorf("Full arena Utilization = %f, want close to 1.0", util)
	}
}

func BenchmarkMetrics(b *testing.B) {
	a := NewArena(1024 * 1024)
	// Pre-allocate some data
	for i := 0; i < 100; i++ {
		a.AllocBytes(1000)
	}

	b.Run("SizeInUse", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			a.SizeInUse()
		}
	})

	b.Run("NumChunks", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			a.NumChunks()
		}
	})

	b.Run("Capacity", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			a.Capacity()
		}
	})

	b.Run("Utilization", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			a.Utilization()
		}
	})

	b.Run("Metrics", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			a.Metrics()
		}
	})
}

func BenchmarkSafeArenaMetrics(b *testing.B) {
	s := NewSafeArena(1024 * 1024)
	// Pre-allocate some data
	for i := 0; i < 100; i++ {
		s.AllocBytes(1000)
	}

	b.Run("SafeSizeInUse", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			s.SizeInUse()
		}
	})

	b.Run("SafeMetrics", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			s.Metrics()
		}
	})
}
