package linescanner

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type ReaderMock struct {
	io.ReaderAt
	mock.Mock
}

func (r *ReaderMock) ReadAt(p []byte, off int64) (n int, err error) {
	args := r.Called(p, off)
	return args.Int(0), args.Error(1)
}

func TestBackward_NewBackward(t *testing.T) {
	// given
	reader := strings.NewReader("")
	position := 100

	// when
	backward := newBackward(reader, position)

	// then
	assert.Equal(t, backward.reader, reader)
	assert.Equal(t, backward.readerPos, position)
	assert.Equal(t, backward.readerLineEndPos, position)
	assert.Equal(t, backward.maxChunkSize, defaultMaxChunkSize)
	assert.Equal(t, backward.maxBufferSize, defaultMaxBufferSize)
}

func TestBackward_NewBackward_ErrNilReader(t *testing.T) {
	assert.PanicsWithValue(t, ErrNilReader, func() {
		newBackward(nil, endPosition)
	})
}

func TestBackward_NewBackward_ErrInvalidMaxChunkSize(t *testing.T) {
	assert.PanicsWithValue(t, ErrInvalidMaxChunkSize, func() {
		newBackwardWithSize(strings.NewReader(""), endPosition, 0, 100)
	})
}

func TestBackward_NewBackward_ErrInvalidMaxBufferSize(t *testing.T) {
	assert.PanicsWithValue(t, ErrInvalidMaxBufferSize, func() {
		newBackwardWithSize(strings.NewReader(""), endPosition, 100, 0)
	})
}

func TestBackward_NewBackward_ErrGreaterBufferSize(t *testing.T) {
	assert.PanicsWithValue(t, ErrGreaterBufferSize, func() {
		newBackwardWithSize(strings.NewReader(""), endPosition, 100, 10)
	})
}

func TestBackward_NewBackwardWithSize(t *testing.T) {
	// given
	reader := strings.NewReader("")
	position := 100
	maxChunkSize := 1024
	maxBufferSize := 4096

	// when
	backward := newBackwardWithSize(reader, position, maxChunkSize, maxBufferSize)

	// then
	assert.Equal(t, backward.reader, reader)
	assert.Equal(t, backward.readerPos, position)
	assert.Equal(t, backward.readerLineEndPos, position)
	assert.Equal(t, backward.maxChunkSize, maxChunkSize)
	assert.Equal(t, backward.maxBufferSize, maxBufferSize)
}

func TestBackward_EndOfFile_False(t *testing.T) {
	// given
	backward := newBackward(strings.NewReader(""), 0)

	// when
	backward.readerPos = 1

	// then
	assert.False(t, backward.endOfFile())
}

func TestBackward_EndOfFile_True(t *testing.T) {
	// given
	backward := newBackward(strings.NewReader(""), 0)

	// when
	backward.readerPos = -1

	// then
	assert.True(t, backward.endOfFile())

	// when
	backward.readerPos = 0

	// then
	assert.True(t, backward.endOfFile())
}

func TestBackward_EndOfScan_False(t *testing.T) {
	// given
	backward := newBackward(strings.NewReader(""), 0)

	// when
	backward.readerPos = -1
	backward.readerLineEndPos = 1

	// then
	assert.False(t, backward.endOfScan())
}

func TestBackward_EndOfScan_True(t *testing.T) {
	// given
	backward := newBackward(strings.NewReader(""), 0)

	// when
	backward.readerPos = -1
	backward.readerLineEndPos = -1

	// then
	assert.True(t, backward.endOfScan())
}

func TestBackward_AllocateChunk(t *testing.T) {
	// given
	backward := newBackwardWithSize(strings.NewReader("abcdefgh"), 8, 4, 4)

	// when
	err := backward.allocateChunk()

	// then
	assert.Nil(t, err)
	assert.Equal(t, backward.chunk, []byte("efgh"))
	assert.Equal(t, cap(backward.chunk), 4)
}

func TestBackward_AllocateChunk_GreaterThanReaderPosWhenFirstAllocated(t *testing.T) {
	// given
	backward := newBackwardWithSize(strings.NewReader("abcd"), 4, 4, 4)
	backward.readerPos = 2

	// when
	err := backward.allocateChunk()

	// then
	assert.Nil(t, err)
	assert.Equal(t, len(backward.chunk), 2)
	assert.Equal(t, cap(backward.chunk), 2)
}

func TestBackward_AllocateChunk_GreaterThanReaderPosWhenAlreadyAllocated(t *testing.T) {
	// given
	backward := newBackwardWithSize(strings.NewReader("abcdef"), 6, 4, 4)

	// when
	err := backward.allocateChunk()

	// then
	assert.Nil(t, err)
	assert.Equal(t, len(backward.chunk), backward.maxChunkSize)
	assert.Equal(t, cap(backward.chunk), backward.maxChunkSize)

	// given
	backward.readerPos = 2

	// when
	err = backward.allocateChunk()

	// then
	assert.Nil(t, err)
	assert.Equal(t, len(backward.chunk), 2)
	assert.Equal(t, cap(backward.chunk), backward.maxChunkSize)
}

func TestBackward_AllocateChunk_WithPosition(t *testing.T) {
	// given
	data := "abcd\nefgh\nijkl"
	backward := newBackwardWithSize(strings.NewReader(data), len(data)-2, 4, 14)

	// when
	err := backward.allocateChunk()

	// then
	assert.Nil(t, err)
	assert.Equal(t, backward.chunk, []byte("h\nij"))
}

func TestBackward_AllocateChunk_InvalidPosition(t *testing.T) {
	// given
	data := "abcd\nefgh\nijkl"
	backward := newBackwardWithSize(strings.NewReader(data), len(data)+1, 4, 14)

	// when
	err := backward.allocateChunk()

	// then
	assert.Equal(t, err, ErrInvalidPosition)
	assert.Equal(t, backward.chunk, append([]byte("jkl"), 0x00))
}

func TestBackward_AllocateChunk_ReadError(t *testing.T) {
	// given
	readErr := errors.New("")
	reader := new(ReaderMock)
	reader.On("ReadAt", mock.Anything, mock.Anything).Return(0, readErr)
	backward := newBackward(reader, 10)

	// when
	err := backward.allocateChunk()

	// then
	assert.Equal(t, err, readErr)
}

func TestBackward_AllocateChunk_ReadFailure(t *testing.T) {
	// given
	reader := new(ReaderMock)
	reader.On("ReadAt", mock.Anything, mock.Anything).Return(10, nil)
	backward := newBackward(reader, 20)

	// when
	err := backward.allocateChunk()

	// then
	assert.Equal(t, err, ErrReadFailure)
}

func TestBackward_AllocateChunk_LessThanReaderPos(t *testing.T) {
	// given
	backward := newBackwardWithSize(strings.NewReader("abcdef"), 6, 4, 4)

	// when
	err := backward.allocateChunk()

	// then
	assert.Nil(t, err)
	assert.Equal(t, len(backward.chunk), backward.maxChunkSize)
	assert.Equal(t, cap(backward.chunk), backward.maxChunkSize)
}

func TestBackward_AllocateBuffer(t *testing.T) {
	// given
	chunk := []byte("abcd")
	backward := newBackwardWithSize(strings.NewReader(""), 0, len(chunk), len(chunk))
	backward.chunk = chunk

	// when
	err := backward.allocateBuffer()

	// then
	assert.Nil(t, err)
	assert.Equal(t, backward.buffer, backward.chunk)
	assert.Equal(t, cap(backward.buffer), len(backward.chunk))
}

func TestBackward_AllocateBuffer_BufferOverflow(t *testing.T) {
	// given
	chunk := []byte("abcd")
	buffer := make([]byte, 1, len(chunk))
	backward := newBackwardWithSize(strings.NewReader(""), 0, len(chunk), cap(buffer))
	backward.chunk = chunk
	backward.buffer = buffer

	// when
	err := backward.allocateBuffer()

	// then
	assert.Equal(t, err, ErrBufferOverflow)
	assert.Equal(t, backward.buffer, buffer)
	assert.Equal(t, backward.readerPos, 0)
}

func TestBackward_AllocateBuffer_BufferExpanded(t *testing.T) {
	// given
	chunk := []byte("abcd")
	buffer := make([]byte, 1, len(chunk)+1)
	buffer[0] = 'e'
	backward := newBackwardWithSize(strings.NewReader(""), 0, len(chunk), cap(buffer))
	backward.chunk = chunk
	backward.buffer = buffer

	// when
	err := backward.allocateBuffer()

	// then
	assert.Nil(t, err)
	assert.Equal(t, backward.buffer, []byte("abcde"))
	assert.Equal(t, cap(backward.buffer), len(backward.buffer))
}

func TestBackward_AllocateBuffer_BufferReused(t *testing.T) {
	// given
	chunk := []byte("abcd")
	buffer := make([]byte, 1, 10)
	buffer[0] = 'e'
	backward := newBackwardWithSize(strings.NewReader(""), 0, len(chunk), cap(buffer))
	backward.chunk = chunk
	backward.buffer = buffer

	// when
	err := backward.allocateBuffer()

	// then
	assert.Nil(t, err)
	assert.Equal(t, backward.buffer, []byte("abcde"))
	assert.Equal(t, cap(backward.buffer), 10)
}

func TestBackward_RemoveLineFromBuffer(t *testing.T) {
	// given
	backward := newBackwardWithSize(strings.NewReader(""), 16, 4, 8)
	backward.buffer = []byte("a\r\ndefg\r")

	// when
	line := backward.removeLineFromBuffer(2)

	// then
	assert.Equal(t, line, "defg")
	assert.Equal(t, len(backward.buffer), 2)
	assert.Equal(t, cap(backward.buffer), 8)
	assert.Equal(t, backward.readerLineEndPos, 16-(len(line) /* line feed */ +1 /* carrage return */ +1))
}

func TestBackward_Read(t *testing.T) {
	// given
	data := "abcd\nefgh\nijkl"
	backward := newBackwardWithSize(strings.NewReader(data), len(data), 4, 14)

	// when
	err := backward.read()

	// then
	assert.Nil(t, err)
	assert.Equal(t, backward.chunk, []byte("ijkl"))
	assert.Equal(t, cap(backward.chunk), 4)
	assert.Equal(t, backward.buffer, []byte("ijkl"))
	assert.Equal(t, cap(backward.buffer), 4)
	assert.Equal(t, backward.readerPos, 10)
	assert.False(t, backward.endOfFile())

	// when
	err = backward.read()

	// then
	assert.Nil(t, err)
	assert.Equal(t, backward.chunk, []byte("fgh\n"))
	assert.Equal(t, cap(backward.chunk), 4)
	assert.Equal(t, backward.buffer, []byte("fgh\nijkl"))
	assert.Equal(t, cap(backward.buffer), 8)
	assert.Equal(t, backward.readerPos, 6)
	assert.False(t, backward.endOfFile())

	// when
	err = backward.read()

	// then
	assert.Nil(t, err)
	assert.Equal(t, backward.chunk, []byte("cd\ne"))
	assert.Equal(t, cap(backward.chunk), 4)
	assert.Equal(t, backward.buffer, []byte("cd\nefgh\nijkl"))
	assert.Equal(t, cap(backward.buffer), 12)
	assert.Equal(t, backward.readerPos, 2)
	assert.False(t, backward.endOfFile())

	// when
	err = backward.read()

	// then
	assert.Nil(t, err)
	assert.Equal(t, backward.chunk, []byte("ab"))
	assert.Equal(t, cap(backward.chunk), 4)
	assert.Equal(t, backward.buffer, []byte("abcd\nefgh\nijkl"))
	assert.Equal(t, cap(backward.buffer), 14)
	assert.Equal(t, backward.readerPos, 0)
	assert.True(t, backward.endOfFile())
}

func TestBackward_Read_AllocateChunkError(t *testing.T) {
	// given
	readErr := errors.New("")
	reader := new(ReaderMock)
	reader.On("ReadAt", mock.Anything, mock.Anything).Return(0, readErr)
	backward := newBackward(reader, 10)

	// when
	err := backward.read()

	// then
	assert.Equal(t, err, readErr)
	assert.Equal(t, backward.readerPos, 10)
}

func TestBackward_Read_AllocateBufferError(t *testing.T) {
	// given
	data := "abcd"
	buffer := make([]byte, 1, len(data))
	readerPos := 4
	backward := newBackwardWithSize(strings.NewReader(data), readerPos, len(data), cap(buffer))
	backward.buffer = buffer

	// when
	err := backward.read()

	// then
	assert.Equal(t, err, ErrBufferOverflow)
	assert.Equal(t, backward.readerPos, 0)
}

func TestBackward_Line(t *testing.T) {
	// given
	data := "a\nb\r\ncdef\nghij"
	backward := newBackwardWithSize(strings.NewReader(data), len(data), 4, 8)

	// when
	line, err := backward.Line()

	// then
	assert.Nil(t, err)
	assert.Equal(t, line, "ghij")
	assert.Equal(t, backward.chunk, []byte("def\n"))
	assert.Equal(t, cap(backward.chunk), 4)
	assert.Equal(t, backward.buffer, []byte("def"))
	assert.Equal(t, cap(backward.buffer), 8)
	assert.Equal(t, backward.readerPos, 6)
	assert.Equal(t, backward.readerLineEndPos, 9)
	assert.False(t, backward.endOfFile())
	assert.False(t, backward.endOfScan())

	// when
	line, err = backward.Line()

	// then
	assert.Nil(t, err)
	assert.Equal(t, line, "cdef")
	assert.Equal(t, backward.chunk, []byte("b\r\nc"))
	assert.Equal(t, cap(backward.chunk), 4)
	assert.Equal(t, backward.buffer, []byte("b\r"))
	assert.Equal(t, cap(backward.buffer), 8)
	assert.Equal(t, backward.readerPos, 2)
	assert.Equal(t, backward.readerLineEndPos, 4)
	assert.False(t, backward.endOfFile())
	assert.False(t, backward.endOfScan())

	// when
	line, err = backward.Line()

	// then
	assert.Nil(t, err)
	assert.Equal(t, line, "b")
	assert.Equal(t, backward.chunk, []byte("a\n"))
	assert.Equal(t, cap(backward.chunk), 4)
	assert.Equal(t, backward.buffer, []byte("a"))
	assert.Equal(t, cap(backward.buffer), 8)
	assert.Equal(t, backward.readerPos, 0)
	assert.Equal(t, backward.readerLineEndPos, 1)
	assert.True(t, backward.endOfFile())
	assert.False(t, backward.endOfScan())

	// when
	line, err = backward.Line()

	// then
	assert.Equal(t, err, io.EOF)
	assert.Equal(t, line, "a")
	assert.Equal(t, backward.chunk, []byte("a\n"))
	assert.Equal(t, cap(backward.chunk), 4)
	assert.Empty(t, backward.buffer)
	assert.Equal(t, cap(backward.buffer), 8)
	assert.Equal(t, backward.readerPos, 0)
	assert.Equal(t, backward.readerLineEndPos, 0)
	assert.True(t, backward.endOfFile())
	assert.True(t, backward.endOfScan())
}

func TestBackward_Line_Error(t *testing.T) {
	// given
	readErr := errors.New("")
	reader := new(ReaderMock)
	reader.On("ReadAt", mock.Anything, mock.Anything).Return(0, readErr)
	backward := newBackward(reader, 100)

	// when
	line, err := backward.Line()

	// then
	assert.Equal(t, err, readErr)
	assert.Empty(t, line)

	// when
	line, err = backward.Line()

	// then
	assert.Equal(t, err, readErr)
	assert.Empty(t, line)
}

func TestBackward_Line_AlreadyEndOfScan(t *testing.T) {
	// given
	backward := newBackward(strings.NewReader(""), endPosition)

	// when
	line, err := backward.Line()

	// then
	assert.Equal(t, err, io.EOF)
	assert.Equal(t, line, "")
	assert.True(t, backward.endOfFile())
	assert.True(t, backward.endOfScan())
}

func TestBackward_Line_LineFeedOnly(t *testing.T) {
	// given
	data := "\n\r\n\n"
	backward := newBackward(strings.NewReader(data), len(data))

	// when
	line, err := backward.Line()

	// then
	assert.Nil(t, err)
	assert.Empty(t, line)
	assert.True(t, backward.endOfFile())
	assert.False(t, backward.endOfScan())

	// when
	line, err = backward.Line()

	// then
	assert.Nil(t, err)
	assert.Empty(t, line)
	assert.True(t, backward.endOfFile())
	assert.False(t, backward.endOfScan())

	// when
	line, err = backward.Line()

	// then
	assert.Nil(t, err)
	assert.Empty(t, line)
	assert.True(t, backward.endOfFile())
	assert.True(t, backward.endOfScan())

	// when
	line, err = backward.Line()

	// then
	assert.Equal(t, err, io.EOF)
	assert.Empty(t, line)
	assert.True(t, backward.endOfFile())
	assert.True(t, backward.endOfScan())
}

func TestBackward_Position(t *testing.T) {
	// given
	data := "abcdefgh\r\nhij"
	backward := newBackward(strings.NewReader(data), len(data))

	// when
	line, err := backward.Line()

	// then
	assert.Nil(t, err)
	assert.Equal(t, line, "hij")
	assert.Equal(t, backward.Position(), 9)

	// given
	backward = newBackward(strings.NewReader(data), backward.Position())

	// when
	line, err = backward.Line()

	// then
	assert.Equal(t, err, io.EOF)
	assert.Equal(t, line, "abcdefgh")
	assert.Equal(t, backward.Position(), endPosition)
}
