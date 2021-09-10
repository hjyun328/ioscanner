package linescanner

import (
	"bytes"
	"io"
)

type forward struct {
	reader io.ReaderAt

	chunk  []byte
	buffer []byte

	bufferLineStartPos int
	readerPos          int
	readerLineStartPos int

	backupBufferLineStartPos int
	backupReaderPos          int
	backupReaderLineStartPos int

	endOfFile bool
	endOfScan bool
}

func newForward(reader io.ReaderAt, position int) *forward {
	return newForwardWithSize(reader, position, defaultChunkSize, defaultBufferSize)
}

func newForwardWithSize(reader io.ReaderAt, position int, chunkSize int, bufferSize int) *forward {
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
	return &forward{
		reader:             reader,
		chunk:              make([]byte, chunkSize),
		buffer:             make([]byte, 0, bufferSize),
		readerPos:          position,
		readerLineStartPos: position,
	}
}

func (s *forward) backupPosition() {
	s.backupBufferLineStartPos = s.bufferLineStartPos
	s.backupReaderPos = s.readerPos
	s.backupReaderLineStartPos = s.readerLineStartPos
}

func (s *forward) recoverPosition() {
	s.bufferLineStartPos = s.backupBufferLineStartPos
	s.readerPos = s.backupReaderPos
	s.readerLineStartPos = s.backupReaderLineStartPos
}

func (s *forward) endPosition() {
	s.bufferLineStartPos = endPosition
	s.readerPos = endPosition
	s.readerLineStartPos = endPosition
}

func (s *forward) getLineSizeExcludingLineFeed() int {
	lineSize := bytes.IndexByte(s.buffer[s.bufferLineStartPos:], '\n')
	if lineSize < 0 && s.endOfFile {
		s.endOfScan = true
		return len(s.buffer[s.bufferLineStartPos:])
	}
	return lineSize
}

func (s *forward) getLineExcludingCarrageReturn(lineSize int) string {
	return removeCarrageReturn(s.buffer[s.bufferLineStartPos : s.bufferLineStartPos+lineSize])
}

func (s *forward) arrangeBuffer(n int) error {
	lineSize := len(s.buffer[s.bufferLineStartPos:])
	if lineSize+n > cap(s.buffer) {
		return ErrBufferOverflow
	}
	if s.bufferLineStartPos+lineSize+n > cap(s.buffer) {
		copy(s.buffer, s.buffer[s.bufferLineStartPos:])
		s.buffer = s.buffer[:lineSize]
		s.bufferLineStartPos = 0
	}
	return nil
}

func (s *forward) read() error {
	n, err := s.reader.ReadAt(s.chunk, int64(s.readerPos))
	if err != nil && err != io.EOF {
		return err
	}
	if n > 0 {
		if err := s.arrangeBuffer(n); err != nil {
			return err
		}
		s.buffer = append(s.buffer, s.chunk[:n]...)
		s.readerPos += n
	}
	if err == io.EOF {
		s.endOfFile = true
		s.readerPos = -1
	}
	return nil
}

func (s *forward) Line() (string, error) {
	if s.endOfScan {
		return "", io.EOF
	}
	s.backupPosition()
	for {
		lineSize := s.getLineSizeExcludingLineFeed()
		if lineSize < 0 {
			if err := s.read(); err != nil {
				s.recoverPosition()
				return "", err
			}
			continue
		}
		line := s.getLineExcludingCarrageReturn(lineSize)
		s.bufferLineStartPos += lineSize + 1
		s.readerLineStartPos += lineSize + 1
		if s.endOfScan {
			s.endPosition()
			return line, io.EOF
		}
		return line, nil
	}
}

func (s *forward) Position() int {
	return s.readerLineStartPos
}
