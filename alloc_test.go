package arena

import (
	"fmt"
	"testing"
	"unsafe"
)

type testStruct struct {
	a int64
	b int32
	c int16
	d int8
}

func TestAlloc(t *testing.T) {
	a := NewArena(1024)

	// Test basic allocation
	ptr := Alloc[int](a)
	if ptr == nil {
		t.Fatal("Alloc[int] returned nil")
	}
	if *ptr != 0 {
		t.Errorf("Alloc[int] value = %d, want 0 (zeroed)", *ptr)
	}

	// Test struct allocation
	s := Alloc[testStruct](a)
	if s == nil {
		t.Fatal("Alloc[testStruct] returned nil")
	}
	if s.a != 0 || s.b != 0 || s.c != 0 || s.d != 0 {
		t.Errorf("Alloc[testStruct] not properly zeroed: %+v", *s)
	}

	// Verify we can write to allocated memory
	*ptr = 42
	s.a = 100
	if *ptr != 42 || s.a != 100 {
		t.Error("Could not write to allocated memory")
	}
}

func TestAllocZeroed(t *testing.T) {
	a := NewArena(1024)
	ptr := AllocZeroed[int64](a)

	if ptr == nil {
		t.Fatal("AllocZeroed[int64] returned nil")
	}
	if *ptr != 0 {
		t.Errorf("AllocZeroed[int64] value = %d, want 0", *ptr)
	}
}

func TestAllocUninitialized(t *testing.T) {
	a := NewArena(1024)
	ptr := AllocUninitialized[int](a)

	if ptr == nil {
		t.Fatal("AllocUninitialized[int] returned nil")
	}

	// We can't test the value since it's uninitialized,
	// but we can verify we can write to it
	*ptr = 123
	if *ptr != 123 {
		t.Error("Could not write to uninitialized memory")
	}
}

func TestAllocSlice(t *testing.T) {
	a := NewArena(1024)

	// Test normal slice allocation
	slice := AllocSlice[int](a, 10)
	if len(slice) != 10 {
		t.Errorf("AllocSlice[int](10) length = %d, want 10", len(slice))
	}
	if cap(slice) != 10 {
		t.Errorf("AllocSlice[int](10) capacity = %d, want 10", cap(slice))
	}

	// Test zero size
	empty := AllocSlice[int](a, 0)
	if empty != nil {
		t.Errorf("AllocSlice[int](0) = %v, want nil", empty)
	}

	// Test negative size
	negative := AllocSlice[int](a, -1)
	if negative != nil {
		t.Errorf("AllocSlice[int](-1) = %v, want nil", negative)
	}

	// Verify we can write to slice
	for i := range slice {
		slice[i] = i * 2
	}
	for i := range slice {
		if slice[i] != i*2 {
			t.Errorf("slice[%d] = %d, want %d", i, slice[i], i*2)
		}
	}
}

func TestAllocSliceZeroed(t *testing.T) {
	a := NewArena(1024)
	slice := AllocSliceZeroed[int](a, 5)

	if len(slice) != 5 {
		t.Errorf("AllocSliceZeroed[int](5) length = %d, want 5", len(slice))
	}

	// Verify all elements are zeroed
	for i, v := range slice {
		if v != 0 {
			t.Errorf("slice[%d] = %d, want 0 (zeroed)", i, v)
		}
	}
}

func TestPtrAndKeepAlive(t *testing.T) {
	a := NewArena(1024)
	ptr := Alloc[int](a)
	*ptr = 42

	result := PtrAndKeepAlive(a, ptr)
	if result != ptr {
		t.Errorf("PtrAndKeepAlive returned different pointer")
	}
	if *result != 42 {
		t.Errorf("PtrAndKeepAlive value = %d, want 42", *result)
	}
}

func TestAllocAlignment(t *testing.T) {
	a := NewArena(1024)

	// Allocate several pointers and verify they're properly aligned
	ptrs := make([]*int64, 10)
	for i := range ptrs {
		ptrs[i] = Alloc[int64](a)
		addr := uintptr(unsafe.Pointer(ptrs[i]))
		if addr%unsafe.Alignof(int64(0)) != 0 {
			t.Errorf("Pointer %d not properly aligned: %x", i, addr)
		}
	}
}

func BenchmarkAlloc(b *testing.B) {
	a := NewArena(1024 * 1024)

	b.Run("Alloc[int]", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			Alloc[int](a)
			if i%1000 == 999 {
				a.Reset()
			}
		}
	})

	b.Run("AllocUninitialized[int]", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			AllocUninitialized[int](a)
			if i%1000 == 999 {
				a.Reset()
			}
		}
	})
}

func BenchmarkAllocSlice(b *testing.B) {
	a := NewArena(1024 * 1024)
	sizes := []int{10, 100, 1000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("AllocSlice-%d", size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				AllocSlice[int](a, size)
				if i%100 == 99 {
					a.Reset()
				}
			}
		})

		b.Run(fmt.Sprintf("AllocSliceZeroed-%d", size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				AllocSliceZeroed[int](a, size)
				if i%100 == 99 {
					a.Reset()
				}
			}
		})
	}
}
