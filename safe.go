package arena

import (
	"runtime"
	"sync"
)

// SafeArena is a mutex-protected wrapper around Arena for concurrent access.
// All operations are thread-safe but come with the overhead of mutex locking.
type SafeArena struct {
	mu sync.Mutex
	a  *Arena
}

// NewSafeArena creates a new thread-safe arena with the specified chunk size.
// If chunkSize <= 0, DefaultChunkSize is used.
func NewSafeArena(chunkSize int) *SafeArena {
	return &SafeArena{a: NewArena(chunkSize)}
}

// AllocBytes thread-safely allocates n bytes and returns a slice pointing to them.
// Returns nil if n <= 0.
func (s *SafeArena) AllocBytes(n int) []byte {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.a.AllocBytes(n)
}

// EnsureCapacity thread-safely ensures the current chunk has at least n free bytes.
func (s *SafeArena) EnsureCapacity(n int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.a.EnsureCapacity(n)
}

// Reset thread-safely resets allocation offsets to zero for arena reuse.
func (s *SafeArena) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.a.Reset()
}

// Release thread-safely drops all chunks and makes the arena unusable.
func (s *SafeArena) Release() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.a.Release()
}

// Generic allocation functions for SafeArena

// SafeAlloc thread-safely returns a pointer to a T stored inside the arena with zeroed memory.
func SafeAlloc[T any](s *SafeArena) *T {
	s.mu.Lock()
	defer s.mu.Unlock()
	return Alloc[T](s.a)
}

// SafeAllocZeroed is identical to SafeAlloc - provided for API consistency.
func SafeAllocZeroed[T any](s *SafeArena) *T {
	return SafeAlloc[T](s)
}

// SafeAllocUninitialized thread-safely returns a *T without zeroing memory.
func SafeAllocUninitialized[T any](s *SafeArena) *T {
	s.mu.Lock()
	defer s.mu.Unlock()
	return AllocUninitialized[T](s.a)
}

// SafeAllocSlice thread-safely allocates a slice of n elements of type T.
func SafeAllocSlice[T any](s *SafeArena, n int) []T {
	s.mu.Lock()
	defer s.mu.Unlock()
	return AllocSlice[T](s.a, n)
}

// SafeAllocSliceZeroed thread-safely allocates a slice of n elements with zeroed memory.
func SafeAllocSliceZeroed[T any](s *SafeArena, n int) []T {
	s.mu.Lock()
	defer s.mu.Unlock()
	return AllocSliceZeroed[T](s.a, n)
}

// SafePtrAndKeepAlive thread-safely returns t and calls runtime.KeepAlive on the arena.
func SafePtrAndKeepAlive[T any](s *SafeArena, t *T) *T {
	s.mu.Lock()
	defer s.mu.Unlock()
	runtime.KeepAlive(s.a)
	return t
}
