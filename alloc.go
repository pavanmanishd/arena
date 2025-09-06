package arena

import (
	"runtime"
	"unsafe"
)

// Alloc returns a pointer to a T stored inside the arena with zeroed memory.
// The returned pointer is valid as long as the arena hasn't been released.
func Alloc[T any](a *Arena) *T {
	var zero T
	size := int(unsafe.Sizeof(zero))
	b := a.AllocBytes(size)
	// Zero the memory
	if len(b) > 0 {
		clear(b)
	}
	return (*T)(unsafe.Pointer(&b[0]))
}

// AllocZeroed is identical to Alloc - provided for API consistency.
func AllocZeroed[T any](a *Arena) *T {
	return Alloc[T](a)
}

// AllocUninitialized returns a *T located in the arena without zeroing memory.
// This is faster than Alloc but the memory contents are undefined.
// Use with caution - ensure proper initialization before use.
func AllocUninitialized[T any](a *Arena) *T {
	var zero T
	size := int(unsafe.Sizeof(zero))
	b := a.AllocBytes(size)
	return (*T)(unsafe.Pointer(&b[0]))
}

// AllocSlice allocates a slice of n elements of type T inside the arena.
// The slice elements are not initialized (contain garbage data).
// Returns nil if n <= 0.
func AllocSlice[T any](a *Arena, n int) []T {
	if n <= 0 {
		return nil
	}
	var zero T
	elemSize := int(unsafe.Sizeof(zero))
	total := elemSize * n
	b := a.AllocBytes(total)
	return unsafe.Slice((*T)(unsafe.Pointer(&b[0])), n)
}

// AllocSliceZeroed allocates a slice of n elements of type T with zeroed memory.
// This is slower than AllocSlice but ensures clean initialization.
func AllocSliceZeroed[T any](a *Arena, n int) []T {
	if n <= 0 {
		return nil
	}
	var zero T
	elemSize := int(unsafe.Sizeof(zero))
	total := elemSize * n
	b := a.AllocBytes(total)
	// Zero the memory
	if len(b) > 0 {
		clear(b)
	}
	return unsafe.Slice((*T)(unsafe.Pointer(&b[0])), n)
}

// PtrAndKeepAlive returns t and calls runtime.KeepAlive on the arena.
// This is useful to prevent the arena from being garbage collected
// while the pointer is still in use in unsafe code.
func PtrAndKeepAlive[T any](a *Arena, t *T) *T {
	runtime.KeepAlive(a)
	return t
}
