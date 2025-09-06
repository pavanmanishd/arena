package arena

// SizeInUse returns the total number of bytes currently allocated in the arena.
// This includes internal fragmentation due to alignment.
func (a *Arena) SizeInUse() int {
	if a.chunks == nil {
		return 0
	}
	sum := 0
	for _, c := range a.chunks {
		sum += int(c.offset)
	}
	return sum
}

// NumChunks returns the number of chunks currently allocated by the arena.
func (a *Arena) NumChunks() int {
	if a.chunks == nil {
		return 0
	}
	return len(a.chunks)
}

// Capacity returns the total capacity (in bytes) of all chunks in the arena.
func (a *Arena) Capacity() int {
	if a.chunks == nil {
		return 0
	}
	sum := 0
	for _, c := range a.chunks {
		sum += len(c.buf)
	}
	return sum
}

// Utilization returns the ratio of bytes in use to total capacity (0.0 to 1.0).
// Returns 0.0 if the arena has no capacity.
func (a *Arena) Utilization() float64 {
	capacity := a.Capacity()
	if capacity == 0 {
		return 0
	}
	return float64(a.SizeInUse()) / float64(capacity)
}

// ChunkSize returns the default chunk size used by this arena.
func (a *Arena) ChunkSize() int {
	return a.chunkSize
}

// Metrics returns a snapshot of arena statistics.
func (a *Arena) Metrics() ArenaMetrics {
	return ArenaMetrics{
		SizeInUse:   a.SizeInUse(),
		Capacity:    a.Capacity(),
		NumChunks:   a.NumChunks(),
		ChunkSize:   a.ChunkSize(),
		Utilization: a.Utilization(),
	}
}

// ArenaMetrics contains statistical information about an arena.
type ArenaMetrics struct {
	SizeInUse   int     // Bytes currently allocated
	Capacity    int     // Total capacity in bytes
	NumChunks   int     // Number of chunks
	ChunkSize   int     // Default chunk size
	Utilization float64 // Ratio of used to total capacity (0.0-1.0)
}

// Thread-safe metrics for SafeArena

// SizeInUse thread-safely returns the total number of bytes currently allocated.
func (s *SafeArena) SizeInUse() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.a.SizeInUse()
}

// NumChunks thread-safely returns the number of chunks currently allocated.
func (s *SafeArena) NumChunks() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.a.NumChunks()
}

// Capacity thread-safely returns the total capacity of all chunks.
func (s *SafeArena) Capacity() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.a.Capacity()
}

// Utilization thread-safely returns the ratio of bytes in use to total capacity.
func (s *SafeArena) Utilization() float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.a.Utilization()
}

// ChunkSize thread-safely returns the default chunk size.
func (s *SafeArena) ChunkSize() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.a.ChunkSize()
}

// Metrics thread-safely returns a snapshot of arena statistics.
func (s *SafeArena) Metrics() ArenaMetrics {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.a.Metrics()
}
