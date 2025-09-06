// Package arena implements a chunked bump allocator (memory arena).
// Typical usage: create one arena per request, allocate many temporary
// objects from it, then Reset() at the end of the request for O(1) cleanup.
package arena

import "unsafe"

// DefaultChunkSize is the default chunk size for new arenas (64 KiB).
const DefaultChunkSize = 1 << 16

// chunk represents a single memory chunk within an arena.
type chunk struct {
	buf    []byte  // backing memory
	offset uintptr // allocation offset within buf
}

// Arena is a chunked bump allocator. Not goroutine-safe by default.
// Use SafeArena for concurrent access.
type Arena struct {
	chunks    []chunk
	chunkSize int
}

// NewArena creates a new Arena with the specified chunk size.
// If chunkSize <= 0, DefaultChunkSize is used.
func NewArena(chunkSize int) *Arena {
	if chunkSize <= 0 {
		chunkSize = DefaultChunkSize
	}
	a := &Arena{chunkSize: chunkSize}
	a.grow(chunkSize)
	return a
}

// AllocBytes returns a []byte slice pointing into the arena's backing chunk.
// The caller must ensure the arena remains reachable while the returned slice is in use.
// Returns nil if n <= 0.
func (a *Arena) AllocBytes(n int) []byte {
	if n <= 0 {
		return nil
	}
	a.panicIfReleased()

	ci := len(a.chunks) - 1
	if ci < 0 {
		a.grow(n)
		ci = len(a.chunks) - 1
	}
	c := &a.chunks[ci]

	off := alignPtr(c.offset)

	if uintptr(n)+off > uintptr(len(c.buf)) {
		a.grow(n)
		ci = len(a.chunks) - 1
		c = &a.chunks[ci]
		off = alignPtr(c.offset)
	}

	start := int(off)
	end := start + n
	c.offset = uintptr(end)
	return c.buf[start:end]
}

// EnsureCapacity ensures the current chunk has at least n free bytes.
// If not, it grows the arena with a new chunk.
func (a *Arena) EnsureCapacity(n int) {
	a.panicIfReleased()
	ci := len(a.chunks) - 1
	if ci < 0 {
		a.grow(n)
		return
	}
	c := &a.chunks[ci]
	off := alignPtr(c.offset)
	if uintptr(n)+off > uintptr(len(c.buf)) {
		a.grow(n)
	}
}

// Reset resets allocation offsets to zero but keeps allocated chunks for reuse.
// This provides O(1) cleanup for arena reuse.
func (a *Arena) Reset() {
	a.panicIfReleased()
	for i := range a.chunks {
		a.chunks[i].offset = 0
	}
}

// Release drops all chunks and makes the arena unusable.
// Any subsequent operations will panic.
func (a *Arena) Release() {
	a.chunks = nil
}

// grow appends a new chunk of at least min bytes.
func (a *Arena) grow(min int) {
	size := a.chunkSize
	if min > size {
		size = min
	}
	buf := make([]byte, size)
	a.chunks = append(a.chunks, chunk{buf: buf, offset: 0})
}

// panicIfReleased panics if the arena has been released.
func (a *Arena) panicIfReleased() {
	if a.chunks == nil {
		panic("arena: use after Release()")
	}
}

// alignPtr aligns the offset up to pointer size alignment.
func alignPtr(off uintptr) uintptr {
	const align = unsafe.Sizeof(uintptr(0))
	mask := align - 1
	return (off + mask) & ^mask
}
