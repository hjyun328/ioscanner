package linescanner

import (
	"errors"
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

type LineScanner interface {
	Line() (line string, err error)
	Position() int
}
