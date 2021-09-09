package ioscanner

import (
	"bytes"
	"errors"
	"io"
)

var ErrBufferOverflow = errors.New("overflow buffer")

type Scanner struct {
	reader  io.ReaderAt
	eof     bool
	eob     bool

	chunk     []byte
	buffer    []byte
	bufferPos int

	filePos          int
	filePosLineStart int
}

func New(reader io.ReaderAt, position int, chunkSize int, bufferSize int) *Scanner {
	return &Scanner{
		reader:           reader,
		eof: false,
		eob: false,
		chunk:            make([]byte, chunkSize),
		buffer:           make([]byte, 0, bufferSize),
		bufferPos:        0,
		filePos:          position,
		filePosLineStart: position,
	}
}

func (s *Scanner) getLineSizeExcludingLF() int {
	lineSize := bytes.IndexByte(s.buffer[s.bufferPos:], '\n')
	if lineSize < 0 && s.eof {
		s.eob = true
		lineSize = len(s.buffer[s.bufferPos:])
	}
	return lineSize
}

func (s *Scanner) getLineExcludingCR(lineSize int) string {
	line := s.buffer[s.bufferPos : s.bufferPos+lineSize]
	if len(line) > 0 && line[len(line)-1] == '\r' {
		return string(line[:len(line)-1])
	}
	return string(line)
}

func (s *Scanner) read() error {
	n, err := s.reader.ReadAt(s.chunk, int64(s.filePos))
	if err != nil && err != io.EOF {
		return err
	}
	if err == io.EOF {
		s.eof = true
	}
	if n > 0 {
		s.filePos += n
		if len(s.buffer)-s.bufferPos+n >= cap(s.buffer) {
			return ErrBufferOverflow
		}
		if s.bufferPos+n >= cap(s.buffer) {
			s.buffer = s.buffer[:0]
			s.bufferPos = 0
		}
		s.buffer = append(s.buffer, s.chunk[:n]...)
	}
	return nil
}

func (s *Scanner) Line(lineCount int) (lines []string, err error) {
	for {
		lineSize := s.getLineSizeExcludingLF()
		if lineSize < 0 {
			if err := s.read(); err != nil {
				return nil, err
			}
			continue
		}
		lines = append(lines, s.getLineExcludingCR(lineSize))
		s.bufferPos += lineSize + 1
		s.filePosLineStart += lineSize + 1
		if s.eob || len(lines) == lineCount {
			return lines, nil
		}
	}
}

func (s *Scanner) Position() int {
	return s.filePosLineStart
}
