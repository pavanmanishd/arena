package arena

import (
	"runtime"
	"sync"
	"testing"
)

func TestNewSafeArena(t *testing.T) {
	s := NewSafeArena(1024)
	if s == nil {
		t.Fatal("NewSafeArena returned nil")
	}
	if s.a == nil {
		t.Fatal("SafeArena.a is nil")
	}
}

func TestSafeArenaAllocBytes(t *testing.T) {
	s := NewSafeArena(1024)

	b := s.AllocBytes(100)
	if len(b) != 100 {
		t.Errorf("AllocBytes(100) length = %d, want 100", len(b))
	}

	// Test nil for zero/negative size
	if s.AllocBytes(0) != nil {
		t.Error("AllocBytes(0) should return nil")
	}
	if s.AllocBytes(-1) != nil {
		t.Error("AllocBytes(-1) should return nil")
	}
}

func TestSafeArenaOperations(t *testing.T) {
	s := NewSafeArena(1024)

	// Test basic operations
	s.AllocBytes(100)
	if s.SizeInUse() == 0 {
		t.Error("Expected non-zero size in use")
	}

	s.EnsureCapacity(200)
	s.Reset()
	if s.SizeInUse() != 0 {
		t.Error("Expected zero size in use after Reset")
	}

	s.Release()
	// After release, operations should panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic after Release")
		}
	}()
	s.AllocBytes(100)
}

func TestSafeAllocFunctions(t *testing.T) {
	s := NewSafeArena(1024)

	// Test SafeAlloc
	ptr := SafeAlloc[int](s)
	if ptr == nil {
		t.Fatal("SafeAlloc[int] returned nil")
	}
	if *ptr != 0 {
		t.Errorf("SafeAlloc[int] value = %d, want 0", *ptr)
	}

	// Test SafeAllocZeroed
	ptr2 := SafeAllocZeroed[int64](s)
	if ptr2 == nil {
		t.Fatal("SafeAllocZeroed[int64] returned nil")
	}
	if *ptr2 != 0 {
		t.Errorf("SafeAllocZeroed[int64] value = %d, want 0", *ptr2)
	}

	// Test SafeAllocUninitialized
	ptr3 := SafeAllocUninitialized[int](s)
	if ptr3 == nil {
		t.Fatal("SafeAllocUninitialized[int] returned nil")
	}
	*ptr3 = 42 // Should be writable

	// Test SafeAllocSlice
	slice := SafeAllocSlice[int](s, 5)
	if len(slice) != 5 {
		t.Errorf("SafeAllocSlice length = %d, want 5", len(slice))
	}

	// Test SafeAllocSliceZeroed
	slice2 := SafeAllocSliceZeroed[int](s, 3)
	if len(slice2) != 3 {
		t.Errorf("SafeAllocSliceZeroed length = %d, want 3", len(slice2))
	}
	for i, v := range slice2 {
		if v != 0 {
			t.Errorf("slice2[%d] = %d, want 0", i, v)
		}
	}

	// Test SafePtrAndKeepAlive
	result := SafePtrAndKeepAlive(s, ptr)
	if result != ptr {
		t.Error("SafePtrAndKeepAlive returned different pointer")
	}
}

func TestSafeArenaMetrices(t *testing.T) {
	s := NewSafeArena(1024)

	// Initial state
	if s.NumChunks() == 0 {
		t.Error("Expected at least one chunk initially")
	}
	if s.Capacity() == 0 {
		t.Error("Expected non-zero capacity")
	}
	if s.ChunkSize() != 1024 {
		t.Errorf("ChunkSize = %d, want 1024", s.ChunkSize())
	}

	// After allocation
	s.AllocBytes(100)
	if s.SizeInUse() == 0 {
		t.Error("Expected non-zero size in use after allocation")
	}

	util := s.Utilization()
	if util <= 0 || util > 1 {
		t.Errorf("Utilization = %f, want 0 < x <= 1", util)
	}

	// Test Metrics method
	metrics := s.Metrics()
	if metrics.SizeInUse != s.SizeInUse() {
		t.Error("Metrics.SizeInUse mismatch")
	}
	if metrics.Capacity != s.Capacity() {
		t.Error("Metrics.Capacity mismatch")
	}
	if metrics.NumChunks != s.NumChunks() {
		t.Error("Metrics.NumChunks mismatch")
	}
}

func TestSafeArenaConcurrency(t *testing.T) {
	s := NewSafeArena(1024)
	const numGoroutines = 10
	const numAllocsPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Launch multiple goroutines doing allocations
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numAllocsPerGoroutine; j++ {
				// Mix different allocation types
				switch j % 4 {
				case 0:
					s.AllocBytes(64)
				case 1:
					SafeAlloc[int](s)
				case 2:
					SafeAllocSlice[byte](s, 32)
				case 3:
					s.EnsureCapacity(128)
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify arena is still functional
	if s.SizeInUse() == 0 {
		t.Error("Expected non-zero size in use after concurrent operations")
	}
	if s.NumChunks() == 0 {
		t.Error("Expected at least one chunk after concurrent operations")
	}
}

func TestSafeArenaConcurrentResetRelease(t *testing.T) {
	s := NewSafeArena(1024)
	const numWorkers = 5

	var wg sync.WaitGroup
	wg.Add(numWorkers)

	// Workers doing allocations
	for i := 0; i < numWorkers-2; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				s.AllocBytes(32)
				runtime.Gosched() // Yield to allow other goroutines to run
			}
		}()
	}

	// Worker doing periodic resets
	go func() {
		defer wg.Done()
		for i := 0; i < 5; i++ {
			runtime.Gosched()
			s.Reset()
		}
	}()

	// Worker doing metrics reads
	go func() {
		defer wg.Done()
		for i := 0; i < 20; i++ {
			_ = s.SizeInUse()
			_ = s.Utilization()
			_ = s.Metrics()
			runtime.Gosched()
		}
	}()

	wg.Wait()
}

func BenchmarkSafeArena(b *testing.B) {
	s := NewSafeArena(1024 * 1024)

	b.Run("AllocBytes", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			s.AllocBytes(64)
			if i%1000 == 999 {
				s.Reset()
			}
		}
	})

	b.Run("SafeAlloc", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			SafeAlloc[int](s)
			if i%1000 == 999 {
				s.Reset()
			}
		}
	})
}

func BenchmarkSafeArenaConcurrent(b *testing.B) {
	s := NewSafeArena(1024 * 1024)

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
}
