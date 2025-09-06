package arena

import (
	"fmt"
	"sync"
	"unsafe"
)

// Example demonstrates basic arena usage
func Example() {
	// Create a new arena with default chunk size
	a := NewArena(0)
	defer a.Release() // Always clean up

	// Allocate raw bytes
	buf := a.AllocBytes(1024)
	fmt.Printf("Allocated buffer of size: %d\n", len(buf))

	// Allocate a typed value (zeroed)
	ptr := Alloc[int](a)
	*ptr = 42
	fmt.Printf("Allocated int with value: %d\n", *ptr)

	// Allocate a slice
	slice := AllocSlice[int](a, 5)
	for i := range slice {
		slice[i] = i * 2
	}
	fmt.Printf("Allocated slice: %v\n", slice)

	// Check memory usage
	fmt.Printf("Memory in use: %d bytes\n", a.SizeInUse())
	fmt.Printf("Utilization: %.2f%%\n", a.Utilization()*100)

	// Reset for reuse (O(1) operation)
	a.Reset()
	fmt.Printf("After reset, memory in use: %d bytes\n", a.SizeInUse())

	// Output:
	// Allocated buffer of size: 1024
	// Allocated int with value: 42
	// Allocated slice: [0 2 4 6 8]
	// Memory in use: 1072 bytes
	// Utilization: 1.64%
	// After reset, memory in use: 0 bytes
}

// ExampleSafeArena demonstrates thread-safe arena usage
func ExampleSafeArena() {
	// Create a thread-safe arena
	s := NewSafeArena(1024)
	defer s.Release()

	var wg sync.WaitGroup
	const numWorkers = 3

	// Launch concurrent workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Each worker allocates some memory
			buf := s.AllocBytes(100)
			ptr := SafeAlloc[int](s)
			*ptr = id

			fmt.Printf("Worker %d allocated %d bytes\n", id, len(buf))
		}(i)
	}

	wg.Wait()
	fmt.Printf("Total memory in use: %d bytes\n", s.SizeInUse())
	// Output varies due to goroutine scheduling, but shows concurrent allocation
}

// ExampleArena_webServer demonstrates arena usage in a web server context
func ExampleArena_webServer() {
	// Simulate a request handler that uses arena for temporary allocations
	handleRequest := func(requestID int) {
		// Create arena for this request
		a := NewArena(4096) // 4KB chunks
		defer a.Release()

		// Allocate temporary objects for request processing
		requestData := AllocSlice[byte](a, 1024)
		responseBuffer := AllocSlice[byte](a, 2048)

		// Simulate processing
		copy(requestData, []byte("request data"))
		copy(responseBuffer, []byte("response data"))

		fmt.Printf("Request %d processed\n", requestID)
		fmt.Printf("Arena utilization: %.1f%%\n", a.Utilization()*100)
	}

	// Simulate multiple requests
	for i := 1; i <= 3; i++ {
		handleRequest(i)
	}

	// Output:
	// Request 1 processed
	// Arena utilization: 75.0%
	// Request 2 processed
	// Arena utilization: 75.0%
	// Request 3 processed
	// Arena utilization: 75.0%
}

// ExampleArena_Reset demonstrates arena reuse with Reset
func ExampleArena_Reset() {
	a := NewArena(1024)
	defer a.Release()

	for round := 1; round <= 3; round++ {
		// Allocate memory for this round
		for i := 0; i < 5; i++ {
			Alloc[int64](a)
		}

		fmt.Printf("Round %d - Memory in use: %d bytes\n", round, a.SizeInUse())

		// Reset arena for next round (O(1) operation)
		a.Reset()
	}

	// Output:
	// Round 1 - Memory in use: 40 bytes
	// Round 2 - Memory in use: 40 bytes
	// Round 3 - Memory in use: 40 bytes
}

// ExampleArenaMetrics demonstrates monitoring arena performance
func ExampleArenaMetrics() {
	a := NewArena(1024)
	defer a.Release()

	// Allocate various sizes to see metrics
	a.AllocBytes(100)
	Alloc[int64](a)
	AllocSlice[int32](a, 50)

	// Get detailed metrics
	metrics := a.Metrics()
	fmt.Printf("Metrics:\n")
	fmt.Printf("  Size in use: %d bytes\n", metrics.SizeInUse)
	fmt.Printf("  Capacity: %d bytes\n", metrics.Capacity)
	fmt.Printf("  Chunks: %d\n", metrics.NumChunks)
	fmt.Printf("  Chunk size: %d bytes\n", metrics.ChunkSize)
	fmt.Printf("  Utilization: %.1f%%\n", metrics.Utilization*100)

	// Output:
	// Metrics:
	//   Size in use: 312 bytes
	//   Capacity: 1024 bytes
	//   Chunks: 1
	//   Chunk size: 1024 bytes
	//   Utilization: 30.5%
}

// ExampleArena_alignment demonstrates that allocations are properly aligned
func ExampleArena_alignment() {
	a := NewArena(1024)
	defer a.Release()

	// Allocate different types to show alignment
	ptr1 := Alloc[int8](a)
	ptr2 := Alloc[int64](a) // Should be 8-byte aligned
	ptr3 := Alloc[int32](a) // Should be 4-byte aligned

	fmt.Printf("int8 address alignment: %d\n", uintptr(unsafe.Pointer(ptr1))%8)
	fmt.Printf("int64 address alignment: %d\n", uintptr(unsafe.Pointer(ptr2))%8)
	fmt.Printf("int32 address alignment: %d\n", uintptr(unsafe.Pointer(ptr3))%8)

	// Output:
	// int8 address alignment: 0
	// int64 address alignment: 0
	// int32 address alignment: 0
}
