package arena

type Chunk struct {
	buffer []byte
	offset uintptr
}

type Arena struct {
	chunks []Chunk
	chunkSize int
}

const defaultChunkSize = 1024 * 1024 // 1MB

func NewArena(chunkSize int) *Arena {
	if chunkSize <= 0 {
		chunkSize = defaultChunkSize
	}

	a := &Arena{chunkSize: chunkSize}
	a.grow(chunkSize)

	return a
}

func (a *Arena) grow(minSize int) {
	size := a.chunkSize
	if minSize > size {
		size = minSize
	}

	chunk := make([]byte, size)
	a.chunks = append(a.chunks, Chunk{buffer: chunk, offset: 0})
}