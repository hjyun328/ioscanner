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

	backupReaderPos          int
	backupReaderLineStartPos int
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

func (f *forward) backupPosition() {
	f.backupReaderPos = f.readerPos
	f.backupReaderLineStartPos = f.readerLineStartPos
}

func (f *forward) recoverPosition() {
	f.readerPos = f.backupReaderPos
	f.readerLineStartPos = f.backupReaderLineStartPos
}

func (f *forward) endOfFile() bool {
	return f.readerPos < 0
}

func (f *forward) endOfScan() bool {
	return f.endOfFile() && f.readerLineStartPos < 0
}

func (f *forward) endPosition() {
	f.readerPos = endPosition
	f.readerLineStartPos = endPosition
	f.bufferLineStartPos = endPosition
}

func (f *forward) allocateChunk() error {
	if f.endOfFile() {
		return nil
	}
	if f.chunk == nil {
		f.chunk = make([]byte, f.maxChunkSize)
	} else {
		f.chunk = f.chunk[:f.maxChunkSize]
	}
	n, err := f.reader.ReadAt(f.chunk, int64(f.readerPos))
	if err != nil {
		if err != io.EOF {
			return err
		}
		f.readerPos = endPosition
	}
	f.chunk = f.chunk[:n]
	return nil
}

func (f *forward) allocateBuffer() error {
	chunkSize := len(f.chunk)
	if chunkSize <= 0 {
		return nil
	}
	lineSize := len(f.buffer[f.bufferLineStartPos:])
	if lineSize+chunkSize > f.maxBufferSize {
		return ErrBufferOverflow
	}
	if lineSize+chunkSize > cap(f.buffer) {
		expandedBuffer := make([]byte, 0, lineSize+chunkSize)
		expandedBuffer = append(expandedBuffer, f.buffer[f.bufferLineStartPos:]...)
		expandedBuffer = append(expandedBuffer, f.chunk...)
		f.buffer = expandedBuffer
	} else if f.bufferLineStartPos+lineSize+chunkSize > cap(f.buffer) {
		copy(f.buffer, f.buffer[f.bufferLineStartPos:])
		f.buffer = f.buffer[:lineSize]
		f.bufferLineStartPos = 0
	}
	f.buffer = append(f.buffer, f.chunk...)
	return nil
}

func (f *forward) arrangeBuffer(n int) error {
	lineSize := len(f.buffer[f.bufferLineStartPos:])
	if lineSize+n > cap(f.buffer) {
		return ErrBufferOverflow
	}
	if f.bufferLineStartPos+lineSize+n > cap(f.buffer) {
		copy(f.buffer, f.buffer[f.bufferLineStartPos:])
		f.buffer = f.buffer[:lineSize]
		f.bufferLineStartPos = 0
	}
	return nil
}

func (f *forward) removeLineFromBuffer(lineSize int) string {
	line := removeCarrageReturn(f.buffer[f.bufferLineStartPos : f.bufferLineStartPos+lineSize])
	f.readerLineStartPos += lineSize + 1
	return line
}

func (f *forward) read() (err error) {
	if err = f.allocateChunk(); err != nil {
		return err
	}
	if err = f.allocateBuffer(); err != nil {
		return err
	}
	f.readerPos += len(f.chunk)
	return nil
}

func (f *forward) Line() (string, error) {
	if f.endOfScan() {
		return "", io.EOF
	}
	f.backupPosition()
	for {
		lineSize := bytes.LastIndexByte(f.buffer[f.bufferLineStartPos:], '\n')
		if lineSize >= 0 {
			return f.removeLineFromBuffer(lineSize), nil
		} else {
			if f.endOfFile() {
				line, err := f.removeLineFromBuffer(lineSize), io.EOF
				f.endPosition()
				return line, err
			}
			if err := f.read(); err != nil {
				f.recoverPosition()
				return "", err
			}
		}
	}
}

func (f *forward) Position() int {
	return f.readerLineStartPos
}
