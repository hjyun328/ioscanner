package linescanner

import (
	"bytes"
	"errors"
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

	backupReaderPos        int
	backupReaderLineEndPos int
}

func newBackward(reader io.ReaderAt, position int) *backward {
	return newBackwardWithSize(reader, position, defaultChunkSize, defaultBufferSize)
}

func newBackwardWithSize(reader io.ReaderAt, position, maxChunkSize int, maxBufferSize int) *backward {
	if reader == nil {
		panic(ErrNilReader)
	}
	if position < 0 {
		panic(ErrInvalidPosition)
	}
	if maxChunkSize <= 0 {
		panic(ErrInvalidChunkSize)
	}
	if maxBufferSize <= 0 {
		panic(ErrInvalidBufferSize)
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

func (b *backward) backupPosition() {
	b.backupReaderPos = b.readerPos
	b.backupReaderLineEndPos = b.readerLineEndPos
}

func (b *backward) recoverPosition() {
	b.readerPos = b.backupReaderPos
	b.readerLineEndPos = b.backupReaderLineEndPos
}

func (b *backward) endPosition() {
	b.readerPos = endPosition
	b.readerLineEndPos = endPosition
}

func (b *backward) endOfFile() bool {
	return b.readerPos == endPosition
}

func (b *backward) endOfScan() bool {
	return b.endOfFile() && b.readerLineEndPos == endPosition
}

func (b *backward) allocateChunk() {
	chunkSize := minInt(b.readerPos, b.maxChunkSize)
	if b.chunk == nil {
		b.chunk = make([]byte, chunkSize)
	}
	b.chunk = b.chunk[:chunkSize]
}

func (b *backward) allocateBuffer(chunkSize int) error {
	if b.buffer == nil {
		b.buffer = make([]byte, 0, chunkSize)
		b.buffer = append(b.buffer, b.chunk...)
	} else {
		if len(b.buffer)+chunkSize > b.maxBufferSize {
			return ErrBufferOverflow
		}
		if len(b.buffer)+chunkSize > cap(b.buffer) {
			expandedBuffer := make([]byte, 0, len(b.buffer)+chunkSize)
			expandedBuffer = append(expandedBuffer, b.chunk...)
			expandedBuffer = append(expandedBuffer, b.buffer...)
			b.buffer = expandedBuffer
		} else {
			prevBufferSize := len(b.buffer)
			b.buffer = b.buffer[:prevBufferSize+chunkSize]
			copy(b.buffer[chunkSize:], b.buffer[:prevBufferSize])
			copy(b.buffer, b.chunk)
		}
	}
	return nil
}

func (b *backward) removeLineFromBuffer(lineStartPos int) string {
	line := removeCarrageReturn(b.buffer[lineStartPos+1:])
	b.buffer = b.buffer[:lineStartPos]
	b.readerLineEndPos -= len(line) + 1
	return line
}

func (b *backward) read() error {
	b.allocateChunk()
	n, err := b.reader.ReadAt(b.chunk, int64(b.readerPos-len(b.chunk)))
	if err != nil {
		if err == io.EOF {
			return errors.New("") // FIXME: invalid reader position.
		}
		return err
	}
	if n != len(b.chunk) {
		return errors.New("") // FIXME: read failure chunk size.
	}
	b.readerPos -= n
	if err := b.allocateBuffer(n); err != nil {
		return err
	}
	return nil
}

func (b *backward) Line() (string, error) {
	if b.endOfScan() {
		return "", io.EOF
	}
	b.backupPosition()
	for {
		lineStartPos := bytes.LastIndexByte(b.buffer, '\n')
		if lineStartPos >= 0 {
			return b.removeLineFromBuffer(lineStartPos), nil
		} else {
			if b.endOfFile() {
				b.endPosition()
				return string(b.buffer), io.EOF
			}
			if err := b.read(); err != nil {
				b.recoverPosition()
				return "", err
			}
		}
	}
}
func (b *backward) Position() int {
	return b.readerLineEndPos
}
