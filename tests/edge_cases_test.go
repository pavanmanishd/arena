package arena_test

import (
	"fmt"
	"math"
	"runtime"
	"sync"
	"testing"
	"time"
	"unsafe"

	"github.com/pavanmanishd/arena"
)

// TestEdgeCases covers all edge cases and potential issues
func TestEdgeCases(t *testing.T) {
	t.Run("ZeroAndNegativeChunkSizes", func(t *testing.T) {
		testCases := []struct {
			size     int
			expected int
		}{
			{0, arena.DefaultChunkSize},
			{-1, arena.DefaultChunkSize},
			{-1000, arena.DefaultChunkSize},
			{1, 1},
			{math.MaxInt32, math.MaxInt32},
		}

		for _, tc := range testCases {
			a := arena.NewArena(tc.size)
			if a.ChunkSize() != tc.expected {
				t.Errorf("NewArena(%d): got chunkSize %d, want %d", tc.size, a.ChunkSize(), tc.expected)
			}
			a.Release()
		}
	})

	t.Run("LargeAllocations", func(t *testing.T) {
		a := arena.NewArena(1024)
		defer a.Release()

		// Test allocation larger than chunk size
		large := a.AllocBytes(2048)
		if len(large) != 2048 {
			t.Errorf("Large allocation failed: got %d, want 2048", len(large))
		}

		// Test very large allocation
		veryLarge := a.AllocBytes(1024 * 1024) // 1MB
		if len(veryLarge) != 1024*1024 {
			t.Errorf("Very large allocation failed: got %d, want %d", len(veryLarge), 1024*1024)
		}
	})

	t.Run("IntegerOverflowProtection", func(t *testing.T) {
		a := arena.NewArena(1024)
		defer a.Release()

		// Test potential overflow scenarios
		defer func() {
			if r := recover(); r != nil {
				// Expected for very large allocations
				t.Logf("Recovered from panic (expected): %v", r)
			}
		}()

		// This might cause issues on 32-bit systems
		if unsafe.Sizeof(int(0)) == 8 { // 64-bit system
			// Test allocation that could overflow
			_ = a.AllocBytes(math.MaxInt32)
		}
	})

	t.Run("AlignmentEdgeCases", func(t *testing.T) {
		a := arena.NewArena(1024)
		defer a.Release()

		// Test alignment with various types
		type AlignTest1 struct{ a int8 }
		type AlignTest2 struct{ a int64 }
		type AlignTest3 struct {
			a int8
			b int64
		}

		p1 := arena.Alloc[AlignTest1](a)
		p2 := arena.Alloc[AlignTest2](a)
		p3 := arena.Alloc[AlignTest3](a)

		// Check alignment
		addr1 := uintptr(unsafe.Pointer(p1))
		addr2 := uintptr(unsafe.Pointer(p2))
		addr3 := uintptr(unsafe.Pointer(p3))

		ptrAlign := unsafe.Sizeof(uintptr(0))
		if addr1%ptrAlign != 0 {
			t.Errorf("AlignTest1 not properly aligned: %x", addr1)
		}
		if addr2%ptrAlign != 0 {
			t.Errorf("AlignTest2 not properly aligned: %x", addr2)
		}
		if addr3%ptrAlign != 0 {
			t.Errorf("AlignTest3 not properly aligned: %x", addr3)
		}
	})

	t.Run("UseAfterRelease", func(t *testing.T) {
		a := arena.NewArena(1024)
		a.Release()

		testPanic := func(name string, fn func()) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("%s: expected panic after Release()", name)
				}
			}()
			fn()
		}

		testPanic("AllocBytes", func() { a.AllocBytes(100) })
		testPanic("EnsureCapacity", func() { a.EnsureCapacity(100) })
		testPanic("Reset", func() { a.Reset() })
		testPanic("Alloc", func() { arena.Alloc[int](a) })
		testPanic("AllocSlice", func() { arena.AllocSlice[int](a, 10) })
	})

	t.Run("MultipleReleases", func(t *testing.T) {
		a := arena.NewArena(1024)
		a.Release()
		// Multiple releases should be safe
		a.Release()
		a.Release()
	})

	t.Run("EmptySliceAllocations", func(t *testing.T) {
		a := arena.NewArena(1024)
		defer a.Release()

		// Test zero and negative slice allocations
		s1 := arena.AllocSlice[int](a, 0)
		s2 := arena.AllocSlice[int](a, -1)
		s3 := arena.AllocSliceZeroed[int](a, 0)
		s4 := arena.AllocSliceZeroed[int](a, -1)

		if s1 != nil || s2 != nil || s3 != nil || s4 != nil {
			t.Error("Empty slice allocations should return nil")
		}
	})
}

// TestMemoryCorruption checks for memory corruption issues
func TestMemoryCorruption(t *testing.T) {
	a := arena.NewArena(1024)
	defer a.Release()

	// Allocate multiple objects and verify they don't overlap
	ptrs := make([]*[64]byte, 100)
	for i := range ptrs {
		ptrs[i] = arena.Alloc[[64]byte](a)
		// Fill with pattern
		for j := range ptrs[i] {
			ptrs[i][j] = byte(i)
		}
	}

	// Verify patterns are intact
	for i, ptr := range ptrs {
		for j, b := range ptr {
			if b != byte(i) {
				t.Errorf("Memory corruption detected at ptr[%d][%d]: got %d, want %d", i, j, b, byte(i))
			}
		}
	}
}

// TestBoundaryConditions tests boundary conditions
func TestBoundaryConditions(t *testing.T) {
	t.Run("ExactChunkSizeAllocation", func(t *testing.T) {
		chunkSize := 1024
		a := arena.NewArena(chunkSize)
		defer a.Release()

		// Allocate exactly chunk size
		buf := a.AllocBytes(chunkSize)
		if len(buf) != chunkSize {
			t.Errorf("Exact chunk size allocation failed: got %d, want %d", len(buf), chunkSize)
		}

		// This should trigger a new chunk
		buf2 := a.AllocBytes(1)
		if len(buf2) != 1 {
			t.Errorf("Small allocation after full chunk failed: got %d, want 1", len(buf2))
		}

		if a.NumChunks() < 2 {
			t.Errorf("Expected at least 2 chunks, got %d", a.NumChunks())
		}
	})

	t.Run("AlignmentBoundaries", func(t *testing.T) {
		a := arena.NewArena(1024)
		defer a.Release()

		// Allocate sizes that test alignment boundaries
		sizes := []int{1, 2, 3, 4, 5, 7, 8, 9, 15, 16, 17}
		for _, size := range sizes {
			buf := a.AllocBytes(size)
			if len(buf) != size {
				t.Errorf("Allocation of size %d failed: got %d", size, len(buf))
			}

			// Check alignment
			addr := uintptr(unsafe.Pointer(&buf[0]))
			align := unsafe.Sizeof(uintptr(0))
			if addr%align != 0 {
				t.Errorf("Buffer of size %d not properly aligned: %x", size, addr)
			}
		}
	})
}

// TestTypeSpecificAllocations tests allocation of various Go types
func TestTypeSpecificAllocations(t *testing.T) {
	a := arena.NewArena(4096)
	defer a.Release()

	// Test basic types
	t.Run("BasicTypes", func(t *testing.T) {
		pBool := arena.Alloc[bool](a)
		pInt8 := arena.Alloc[int8](a)
		pInt16 := arena.Alloc[int16](a)
		pInt32 := arena.Alloc[int32](a)
		pInt64 := arena.Alloc[int64](a)
		pUint8 := arena.Alloc[uint8](a)
		pUint16 := arena.Alloc[uint16](a)
		pUint32 := arena.Alloc[uint32](a)
		pUint64 := arena.Alloc[uint64](a)
		pFloat32 := arena.Alloc[float32](a)
		pFloat64 := arena.Alloc[float64](a)

		// Verify zero initialization
		if *pBool != false || *pInt8 != 0 || *pInt16 != 0 || *pInt32 != 0 || *pInt64 != 0 ||
			*pUint8 != 0 || *pUint16 != 0 || *pUint32 != 0 || *pUint64 != 0 ||
			*pFloat32 != 0 || *pFloat64 != 0 {
			t.Error("Basic types not properly zero-initialized")
		}

		// Verify writability
		*pBool = true
		*pInt64 = 12345
		*pFloat64 = 3.14159

		if *pBool != true || *pInt64 != 12345 || *pFloat64 != 3.14159 {
			t.Error("Could not write to allocated basic types")
		}
	})

	// Test complex types
	t.Run("ComplexTypes", func(t *testing.T) {
		type ComplexStruct struct {
			A int64
			B string
			C []int
			D map[string]int
			E *int
		}

		pStruct := arena.Alloc[ComplexStruct](a)
		if pStruct.A != 0 || pStruct.B != "" || pStruct.C != nil || pStruct.D != nil || pStruct.E != nil {
			t.Error("Complex struct not properly zero-initialized")
		}

		// Initialize and test
		pStruct.A = 100
		pStruct.B = "test"
		pStruct.C = []int{1, 2, 3}
		pStruct.D = make(map[string]int)
		pStruct.D["key"] = 42

		if pStruct.A != 100 || pStruct.B != "test" || len(pStruct.C) != 3 || pStruct.D["key"] != 42 {
			t.Error("Could not properly initialize complex struct")
		}
	})

	// Test arrays and slices
	t.Run("ArraysAndSlices", func(t *testing.T) {
		// Fixed arrays
		pArray := arena.Alloc[[10]int](a)
		for i := range pArray {
			if pArray[i] != 0 {
				t.Errorf("Array element %d not zero-initialized: %d", i, pArray[i])
			}
			pArray[i] = i * 2
		}

		// Slices
		slice := arena.AllocSlice[int](a, 20)
		if len(slice) != 20 || cap(slice) != 20 {
			t.Errorf("Slice allocation failed: len=%d, cap=%d", len(slice), cap(slice))
		}

		for i := range slice {
			slice[i] = i * 3
		}

		// Verify values
		for i := range slice {
			if slice[i] != i*3 {
				t.Errorf("Slice element %d: got %d, want %d", i, slice[i], i*3)
			}
		}
	})
}

// TestResetBehavior thoroughly tests Reset functionality
func TestResetBehavior(t *testing.T) {
	a := arena.NewArena(1024)
	defer a.Release()

	// Allocate across multiple chunks
	for i := 0; i < 5; i++ {
		a.AllocBytes(512) // This should create multiple chunks
	}

	initialChunks := a.NumChunks()
	initialCapacity := a.Capacity()

	a.Reset()

	// After reset
	if a.SizeInUse() != 0 {
		t.Errorf("SizeInUse after Reset: got %d, want 0", a.SizeInUse())
	}
	if a.NumChunks() != initialChunks {
		t.Errorf("NumChunks changed after Reset: got %d, want %d", a.NumChunks(), initialChunks)
	}
	if a.Capacity() != initialCapacity {
		t.Errorf("Capacity changed after Reset: got %d, want %d", a.Capacity(), initialCapacity)
	}
	if a.Utilization() != 0 {
		t.Errorf("Utilization after Reset: got %f, want 0", a.Utilization())
	}

	// Verify we can still allocate after reset
	buf := a.AllocBytes(100)
	if len(buf) != 100 {
		t.Errorf("Allocation after Reset failed: got %d, want 100", len(buf))
	}
}

// TestMemoryLeaks checks for potential memory leaks
func TestMemoryLeaks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory leak test in short mode")
	}

	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	// Create and destroy many arenas
	for i := 0; i < 1000; i++ {
		a := arena.NewArena(1024)
		for j := 0; j < 100; j++ {
			a.AllocBytes(64)
		}
		a.Release()
	}

	runtime.GC()
	runtime.ReadMemStats(&m2)

	// Check if memory usage increased significantly
	if m2.Alloc > m1.Alloc*2 {
		t.Errorf("Potential memory leak: before=%d, after=%d", m1.Alloc, m2.Alloc)
	}
}

// TestKeepAlive tests the PtrAndKeepAlive functionality
func TestKeepAlive(t *testing.T) {
	var ptr *int

	func() {
		a := arena.NewArena(1024)
		p := arena.Alloc[int](a)
		*p = 42
		ptr = arena.PtrAndKeepAlive(a, p)
		// Arena should be kept alive by PtrAndKeepAlive call
	}()

	// This is a best-effort test - hard to guarantee GC behavior
	runtime.GC()

	if *ptr != 42 {
		t.Errorf("PtrAndKeepAlive failed: got %d, want 42", *ptr)
	}
}

// TestConcurrencyStress performs stress testing on SafeArena
func TestConcurrencyStress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	s := arena.NewSafeArena(64 * 1024)
	defer s.Release()

	const (
		numWorkers      = 20
		numOpsPerWorker = 1000
	)

	var wg sync.WaitGroup
	errors := make(chan error, numWorkers)

	// Start workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < numOpsPerWorker; j++ {
				switch j % 6 {
				case 0:
					buf := s.AllocBytes(64)
					if len(buf) != 64 {
						errors <- fmt.Errorf("worker %d: AllocBytes failed", workerID)
						return
					}
				case 1:
					ptr := arena.SafeAlloc[int64](s)
					*ptr = int64(workerID*1000 + j)
				case 2:
					slice := arena.SafeAllocSlice[int32](s, 10)
					if len(slice) != 10 {
						errors <- fmt.Errorf("worker %d: AllocSlice failed", workerID)
						return
					}
				case 3:
					s.EnsureCapacity(128)
				case 4:
					_ = s.SizeInUse()
					_ = s.Utilization()
				case 5:
					if j%100 == 0 {
						s.Reset()
					}
				}

				// Yield occasionally
				if j%50 == 0 {
					runtime.Gosched()
				}
			}
		}(i)
	}

	// Wait for completion
	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Error(err)
	}
}

// TestSafeArenaDeadlock tests for potential deadlocks in SafeArena
func TestSafeArenaDeadlock(t *testing.T) {
	s := arena.NewSafeArena(1024)
	defer s.Release()

	done := make(chan bool, 2)
	timeout := time.After(5 * time.Second)

	// Goroutine 1: Continuous allocations
	go func() {
		for i := 0; i < 1000; i++ {
			s.AllocBytes(32)
			if i%100 == 0 {
				runtime.Gosched()
			}
		}
		done <- true
	}()

	// Goroutine 2: Continuous metrics reading
	go func() {
		for i := 0; i < 1000; i++ {
			_ = s.Metrics()
			if i%100 == 0 {
				runtime.Gosched()
			}
		}
		done <- true
	}()

	// Wait for completion or timeout
	completed := 0
	for completed < 2 {
		select {
		case <-done:
			completed++
		case <-timeout:
			t.Fatal("Test timed out - possible deadlock")
		}
	}
}
