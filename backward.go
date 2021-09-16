package linescanner

import (
	"bytes"
	"io"
)

type backward struct {
	reader io.ReaderAt

	maxChunkSize int
	chunk        []byte

	maxBufferSize int
	buffer        []byte

	readerPos        int
	readerLineEndPos int

	err error
}

func NewBackward(reader io.ReaderAt, position int) *backward {
	return NewBackwardWithSize(reader, position, defaultMaxChunkSize, defaultMaxBufferSize)
}

func NewBackwardWithSize(reader io.ReaderAt, position, maxChunkSize int, maxBufferSize int) *backward {
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
	return &backward{
		reader:           reader,
		maxChunkSize:     maxChunkSize,
		maxBufferSize:    maxBufferSize,
		readerPos:        position,
		readerLineEndPos: position,
	}
}

func (b *backward) endOfFile() bool {
	return b.readerPos <= 0
}

func (b *backward) endOfScan() bool {
	return b.endOfFile() && b.readerLineEndPos <= 0
}

func (b *backward) allocateChunk() error {
	chunkSize := minInt(b.readerPos, b.maxChunkSize)
	if b.chunk == nil {
		b.chunk = make([]byte, chunkSize)
	}
	b.chunk = b.chunk[:chunkSize]
	n, err := b.reader.ReadAt(b.chunk, int64(b.readerPos-chunkSize))
	if err != nil {
		if err == io.EOF {
			return ErrInvalidPosition
		}
		return err
	}
	if n != chunkSize {
		return ErrReadFailure
	}
	b.readerPos -= chunkSize
	return nil
}

func (b *backward) allocateBuffer() error {
	chunkSize := len(b.chunk)
	bufferSize := len(b.buffer)
	if chunkSize+bufferSize > b.maxBufferSize {
		return ErrBufferOverflow
	}
	if chunkSize+bufferSize > cap(b.buffer) {
		expandedBuffer := make([]byte, 0, chunkSize+bufferSize)
		expandedBuffer = append(expandedBuffer, b.chunk...)
		expandedBuffer = append(expandedBuffer, b.buffer...)
		b.buffer = expandedBuffer
	} else {
		prevBufferSize := bufferSize
		b.buffer = b.buffer[:chunkSize+prevBufferSize]
		copy(b.buffer[chunkSize:], b.buffer[:prevBufferSize])
		copy(b.buffer, b.chunk)
	}
	return nil
}

func (b *backward) removeLineFromBuffer(lineFeedStartPos int) string {
	lineWithCR := b.buffer[lineFeedStartPos+1:]
	line := removeCarriageReturn(lineWithCR)
	b.buffer = b.buffer[:maxInt(lineFeedStartPos, 0)]
	b.readerLineEndPos -= len(lineWithCR)
	if lineFeedStartPos >= 0 {
		b.readerLineEndPos--
	}
	return line
}

func (b *backward) read() error {
	if err := b.allocateChunk(); err != nil {
		return err
	}
	if err := b.allocateBuffer(); err != nil {
		return err
	}
	return nil
}

func (b *backward) Line() (string, error) {
	if b.err != nil {
		return "", b.err
	}
	if b.endOfScan() {
		return "", io.EOF
	}
	for {
		lineFeedStartPos := bytes.LastIndexByte(b.buffer, '\n')
		if lineFeedStartPos >= 0 {
			return b.removeLineFromBuffer(lineFeedStartPos), nil
		} else {
			if b.endOfFile() {
				return b.removeLineFromBuffer(-1), io.EOF
			}
			if b.err = b.read(); b.err != nil {
				return "", b.err
			}
		}
	}
}

func (b *backward) Position() int {
	if b.readerLineEndPos <= 0 {
		return endPosition
	}
	return b.readerLineEndPos
}
