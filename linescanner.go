package linescanner

import (
	"errors"
	"io"
)

var (
	ErrReadFailure          = errors.New("read failure")
	ErrNilReader            = errors.New("reader is nil")
	ErrInvalidPosition      = errors.New("invalid position")
	ErrInvalidMaxChunkSize  = errors.New("max chunk size is invalid")
	ErrInvalidMaxBufferSize = errors.New("max buffer size is invalid")
	ErrGreaterBufferSize    = errors.New("buffer size must be greater than chunk size")
	ErrBufferOverflow       = errors.New("buffer is overflow")
)

const (
	defaultMaxChunkSize  = 4096
	defaultMaxBufferSize = 1 << 20
	endPosition          = -1
)

type Direction int

const (
	Forward Direction = iota
	Backward
)

type LineScanner interface {
	Line() (line string, err error)
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
