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

	endOfFile bool
	endOfScan bool
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

func (b *backward) read() error {
	// initialize chunk size.
	chunkSize := minInt(b.readerPos, b.maxChunkSize)

	// allocate chunk.
	if b.chunk == nil {
		b.chunk = make([]byte, chunkSize)
	}

	// reconcile chunk len
	b.chunk = b.chunk[:chunkSize]

	// reconcile reader position
	b.readerPos -= chunkSize

	// read
	n, err := b.reader.ReadAt(b.chunk, int64(b.readerPos))
	if err != nil {
		return err
	}

	if b.readerPos == 0 {
		b.endOfFile = true
	}

	// check read success
	if n != chunkSize {
		return errors.New("") // FIXME:
	}

	// reconcile buffer
	if b.buffer == nil {
		b.buffer = make([]byte, 0, n)
		b.buffer = append(b.buffer, b.chunk...)
	} else {
		// check buffer overflow
		if len(b.buffer)+n > b.maxBufferSize {
			return ErrBufferOverflow
		}

		// need buffer expand?
		if len(b.buffer)+n > cap(b.buffer) {
			expandedBuffer := make([]byte, 0, len(b.buffer)+n)
			expandedBuffer = append(expandedBuffer, b.chunk...)
			expandedBuffer = append(expandedBuffer, b.buffer...)
			b.buffer = expandedBuffer
		} else {
			prevBufferSize := len(b.buffer)
			b.buffer = b.buffer[:prevBufferSize+n] // expand length
			copy(b.buffer[n:], b.buffer[:prevBufferSize])
			copy(b.buffer, b.chunk)
		}
	}

	return nil
}

func (b *backward) Line() (string, error) {
	if b.endOfScan {
		return "", io.EOF
	}
	b.backupPosition()
	for {
		lineStartPos := bytes.LastIndexByte(b.buffer, '\n')
		if lineStartPos >= 0 {
			line := removeCarrageReturn(b.buffer[lineStartPos+1:])
			b.buffer = b.buffer[:lineStartPos]
			b.readerLineEndPos -= len(line) + 1
			return line, nil
		} else {
			if b.endOfFile {
				b.endOfScan = true
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
