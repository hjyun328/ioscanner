package linescanner

import "io"

type backward struct {
	reader io.ReaderAt

	chunk  []byte
	buffer []byte
}

func newBackward(reader io.ReaderAt, position int) *backward {
	return newBackwardWithSize(reader, position, defaultChunkSize, defaultBufferSize)
}

func newBackwardWithSize(reader io.ReaderAt, position, chunkSize int, bufferSize int) *backward {
	if reader == nil {
		panic(ErrNilReader)
	}
	if position < 0 {
		panic(ErrInvalidPosition)
	}
	if chunkSize <= 0 {
		panic(ErrInvalidChunkSize)
	}
	if bufferSize <= 0 {
		panic(ErrInvalidBufferSize)
	}
	if chunkSize > bufferSize {
		panic(ErrGreaterBufferSize)
	}
	return &backward{}
}

func (b *backward) Lines(count int) (lines []string, err error) {
	return nil, nil
}

func (b *backward) Position() int {
	return 0
}
