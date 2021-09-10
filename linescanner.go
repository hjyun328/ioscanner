package linescanner

import (
	"errors"
	"io"
)

var (
	ErrNilReader         = errors.New("reader is nil")
	ErrInvalidPosition   = errors.New("invalid position")
	ErrInvalidChunkSize  = errors.New("chunk size is invalid")
	ErrInvalidBufferSize = errors.New("buffer size is invalid")
	ErrGreaterBufferSize = errors.New("buffer size must be greater than chunk size")
	ErrBufferOverflow    = errors.New("buffer is overflow")
	ErrInvalidLineCount  = errors.New("line count is invalid")
)

const (
	defaultChunkSize  = 4096
	defaultBufferSize = 1 << 20
	endPosition       = -1
)

type Direction int

const (
	Forward Direction = iota
	Backward
)

type LineScanner interface {
	Lines(count int) (lines []string, err error)
	Position() int
}

func New(direction Direction, reader io.ReaderAt, position int) LineScanner {
	switch direction {
	case Forward:
		return newForward(reader, position)
	case Backward:
		return newBackward(reader, position)
	}
	return nil
}

func NewWithSize(direction Direction, reader io.ReaderAt, position int, chunkSize, bufferSize int) LineScanner {
	switch direction {
	case Forward:
		return newForwardWithSize(reader, position, chunkSize, bufferSize)
	case Backward:
		return newBackward(reader, position)
	}
	return nil
}
