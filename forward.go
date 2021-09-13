package linescanner

import (
	"bytes"
	"io"
)

type forward struct {
	reader io.ReaderAt

	maxChunkSize int
	chunk        []byte

	maxBufferSize int
	buffer        []byte

	readerPos          int
	readerLineStartPos int
	bufferLineStartPos int

	err error
}

func newForward(reader io.ReaderAt, position int) *forward {
	return newForwardWithSize(reader, position, defaultMaxChunkSize, defaultMaxBufferSize)
}

func newForwardWithSize(reader io.ReaderAt, position int, maxChunkSize int, maxBufferSize int) *forward {
	if reader == nil {
		panic(ErrNilReader)
	}
	if maxChunkSize <= 0 {
		panic(ErrInvalidMaxChunkSize)
	}
	if maxBufferSize <= 0 {
		panic(ErrInvalidMaxBufferSize)
	}
	if maxChunkSize > maxBufferSize {
		panic(ErrGreaterBufferSize)
	}
	return &forward{
		reader:             reader,
		maxChunkSize:       maxChunkSize,
		maxBufferSize:      maxBufferSize,
		readerPos:          position,
		readerLineStartPos: position,
	}
}

func (f *forward) endOfFile() bool {
	return f.readerPos < 0
}

func (f *forward) endOfScan() bool {
	return f.endOfFile() && f.readerLineStartPos < 0
}

func (f *forward) allocateChunk() error {
	if f.chunk == nil {
		f.chunk = make([]byte, f.maxChunkSize)
	} else {
		f.chunk = f.chunk[:f.maxChunkSize]
	}
	n, err := f.reader.ReadAt(f.chunk, int64(f.readerPos))
	if err == nil {
		f.readerPos += len(f.chunk)
	} else {
		if err != io.EOF {
			return err
		}
		f.readerPos = endPosition
	}
	f.chunk = f.chunk[:n]
	return err
}

func (f *forward) allocateBuffer() error {
	chunkSize := len(f.chunk)
	if chunkSize <= 0 {
		return nil
	}
	bufferLineSize := len(f.buffer[f.bufferLineStartPos:])
	if bufferLineSize+chunkSize > f.maxBufferSize {
		return ErrBufferOverflow
	}
	if bufferLineSize+chunkSize > cap(f.buffer) {
		// FIXME: do not allocate buffer if position is less than or equal to maxChunkSize for reusing chunk buffer.
		expandedBuffer := make([]byte, 0, bufferLineSize+chunkSize)
		expandedBuffer = append(expandedBuffer, f.buffer[f.bufferLineStartPos:]...)
		f.buffer = expandedBuffer
		f.bufferLineStartPos = 0
	} else if f.bufferLineStartPos+bufferLineSize+chunkSize > cap(f.buffer) {
		copy(f.buffer, f.buffer[f.bufferLineStartPos:])
		f.buffer = f.buffer[:bufferLineSize]
		f.bufferLineStartPos = 0
	}
	f.buffer = append(f.buffer, f.chunk...)
	return nil
}

func (f *forward) removeLineFromBuffer(lineSize int) string {
	line := removeCarrageReturn(f.buffer[f.bufferLineStartPos : f.bufferLineStartPos+lineSize])
	f.readerLineStartPos += lineSize + 1
	f.bufferLineStartPos += lineSize + 1
	return line
}

func (f *forward) read() (err error) {
	if err = f.allocateChunk(); err != nil && err != io.EOF {
		return err
	}
	if err := f.allocateBuffer(); err != nil {
		return err
	}
	return nil
}

func (f *forward) Line() (string, error) {
	if f.err != nil {
		return "", f.err
	}
	if f.endOfScan() {
		return "", io.EOF
	}
	for {
		lineSize := bytes.IndexByte(f.buffer[f.bufferLineStartPos:], '\n')
		if lineSize >= 0 {
			return f.removeLineFromBuffer(lineSize), nil
		} else {
			if f.endOfFile() {
				line := f.removeLineFromBuffer(len(f.buffer[f.bufferLineStartPos:]))
				f.readerLineStartPos = endPosition
				return line, io.EOF
			}
			if f.err = f.read(); f.err != nil {
				return "", f.err
			}
		}
	}
}

func (f *forward) Position() int {
	return f.readerLineStartPos
}
