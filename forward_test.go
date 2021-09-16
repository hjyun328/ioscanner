package linescanner

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io"
	"strings"
	"testing"
)

func TestForward_NewForward(t *testing.T) {
	// given
	reader := strings.NewReader("")
	position := 100

	// when
	scanner := NewForward(reader, position)

	// then
	assert.Equal(t, scanner.reader, reader)
	assert.Equal(t, scanner.readerPos, position)
	assert.Equal(t, scanner.readerLineStartPos, position)
	assert.Equal(t, scanner.bufferLineStartPos, 0)
	assert.Equal(t, scanner.maxChunkSize, defaultMaxChunkSize)
	assert.Equal(t, scanner.maxBufferSize, defaultMaxBufferSize)
}

func TestForward_NewForward_ErrNilReader(t *testing.T) {
	assert.PanicsWithValue(t, ErrNilReader, func() {
		NewForward(nil, endPosition)
	})
}

func TestForward_NewForward_ErrInvalidMaxChunkSize(t *testing.T) {
	assert.PanicsWithValue(t, ErrInvalidMaxChunkSize, func() {
		NewForwardWithSize(strings.NewReader(""), endPosition, 0, 100)
	})
}

func TestForward_NewForward_ErrInvalidMaxBufferSize(t *testing.T) {
	assert.PanicsWithValue(t, ErrInvalidMaxBufferSize, func() {
		NewForwardWithSize(strings.NewReader(""), endPosition, 100, 0)
	})
}

func TestForward_NewForward_ErrGreaterBufferSize(t *testing.T) {
	assert.PanicsWithValue(t, ErrGreaterBufferSize, func() {
		NewForwardWithSize(strings.NewReader(""), endPosition, 100, 10)
	})
}

func TestForward_EndOfFile_False(t *testing.T) {
	// given
	forward := NewForward(strings.NewReader(""), 0)

	// when
	forward.readerPos = 0

	// then
	assert.False(t, forward.endOfFile())
}

func TestForward_EndOfFile_True(t *testing.T) {
	// given
	forward := NewForward(strings.NewReader(""), 0)

	// when
	forward.readerPos = endPosition

	// then
	assert.True(t, forward.endOfFile())
}

func TestForward_EndOfScan_False(t *testing.T) {
	// given
	forward := NewForward(strings.NewReader(""), 0)

	// when
	forward.readerPos = -1
	forward.readerLineStartPos = 0

	// then
	assert.False(t, forward.endOfScan())
}

func TestForward_EndOfScan_True(t *testing.T) {
	// given
	forward := NewForward(strings.NewReader(""), 0)

	// when
	forward.readerPos = endPosition
	forward.readerLineStartPos = endPosition

	// then
	assert.True(t, forward.endOfScan())
}

func TestForward_AllocateChunk(t *testing.T) {
	// given
	forward := NewForwardWithSize(strings.NewReader("abcdefgh"), 0, 4, 4)
	forward.readerPos = 4

	// when
	err := forward.allocateChunk()

	// then
	assert.Nil(t, err)
	assert.Equal(t, forward.chunk, []byte("efgh"))
	assert.Equal(t, cap(forward.chunk), 4)
	assert.Equal(t, forward.readerPos, 8)
}

func TestForward_AllocateChunk_AlreadyAllocated(t *testing.T) {
	// given
	forward := NewForwardWithSize(strings.NewReader("abcdefgh"), 0, 4, 4)
	forward.readerPos = 4

	// when
	forward.chunk = make([]byte, 2, 4)
	err := forward.allocateChunk()

	// then
	assert.Nil(t, err)
	assert.Equal(t, forward.chunk, []byte("efgh"))
	assert.Equal(t, cap(forward.chunk), 4)
	assert.Equal(t, forward.readerPos, 8)
}

func TestForward_AllocateChunk_WithPosition(t *testing.T) {
	// given
	forward := NewForwardWithSize(strings.NewReader("abcdefg"), 2, 4, 4)

	// when
	err := forward.allocateChunk()

	// then
	assert.Nil(t, err)
	assert.Equal(t, forward.chunk, []byte("cdef"))
	assert.Equal(t, cap(forward.chunk), 4)
	assert.Equal(t, forward.readerPos, 6)
}

func TestForward_AllocateChunk_ReadError(t *testing.T) {
	// given
	readErr := errors.New("")
	reader := new(ReaderMock)
	reader.On("ReadAt", mock.Anything, mock.Anything).Return(0, readErr)
	forward := NewForward(reader, 10)

	// when
	err := forward.allocateChunk()

	// then
	assert.Equal(t, err, readErr)
	assert.Equal(t, forward.readerPos, 10)
}

func TestForward_AllocateChunk_EndOfFile(t *testing.T) {
	// given
	forward := NewForwardWithSize(strings.NewReader("ab"), 0, 4, 4)

	// when
	err := forward.allocateChunk()

	// then
	assert.Nil(t, err)
	assert.True(t, forward.endOfFile())
}

func TestForward_AllocateBuffer(t *testing.T) {
	// given
	chunk := []byte("efgh")
	buffer := make([]byte, 0, 8)
	buffer = append(buffer, []byte("abcd")...)
	lineStartPos := 4
	forward := NewForwardWithSize(strings.NewReader(""), 0, len(chunk), cap(buffer))
	forward.chunk = chunk
	forward.buffer = buffer
	forward.bufferLineStartPos = 4

	// when
	err := forward.allocateBuffer()

	// then
	assert.Nil(t, err)
	assert.Equal(t, forward.buffer, append(buffer, chunk...))
	assert.Equal(t, cap(forward.buffer), cap(buffer))
	assert.Equal(t, forward.bufferLineStartPos, lineStartPos)
}

func TestForward_AllocateBuffer_EmptyChunk(t *testing.T) {
	// given
	lineStartPos := 4
	forward := NewForward(strings.NewReader(""), 0)
	forward.bufferLineStartPos = 4

	// when
	err := forward.allocateBuffer()

	// then
	assert.Nil(t, err)
	assert.Empty(t, forward.buffer)
	assert.Equal(t, cap(forward.buffer), 0)
	assert.Equal(t, forward.bufferLineStartPos, lineStartPos)
}

func TestForward_AllocateBuffer_BufferOverflow(t *testing.T) {
	// given
	chunk := []byte("ab")
	buffer := make([]byte, 4)
	lineStartPos := 1
	forward := NewForwardWithSize(strings.NewReader(""), 0, len(chunk), cap(buffer))
	forward.chunk = chunk
	forward.buffer = buffer
	forward.bufferLineStartPos = lineStartPos

	// when
	err := forward.allocateBuffer()

	// then
	assert.Equal(t, err, ErrBufferOverflow)
	assert.Equal(t, forward.buffer, buffer)
	assert.Equal(t, forward.bufferLineStartPos, lineStartPos)
}

func TestForward_AllocateBuffer_BufferExpanded(t *testing.T) {
	// given
	chunk := []byte("fgh")
	buffer := []byte("a\nbcde")
	bufferLineStartPos := 2
	forward := NewForwardWithSize(strings.NewReader(""), 0, len(chunk), 1024)
	forward.chunk = chunk
	forward.buffer = buffer
	forward.bufferLineStartPos = bufferLineStartPos

	// when
	err := forward.allocateBuffer()

	// then
	assert.Nil(t, err)
	assert.Equal(t, forward.buffer, []byte("bcdefgh"))
	assert.Equal(t, cap(forward.buffer), 7)
	assert.Equal(t, forward.bufferLineStartPos, 0)
}

func TestForward_AllocateBuffer_BufferReused(t *testing.T) {
	// given
	chunk := []byte("fg")
	buffer := []byte("a\nbcde")
	bufferLineStartPos := 2
	forward := NewForwardWithSize(strings.NewReader(""), 0, len(chunk), 1024)
	forward.chunk = chunk
	forward.buffer = buffer
	forward.bufferLineStartPos = bufferLineStartPos

	// when
	err := forward.allocateBuffer()

	// then
	assert.Nil(t, err)
	assert.Equal(t, forward.buffer, []byte("bcdefg"))
	assert.Equal(t, cap(forward.buffer), cap(buffer))
	assert.Equal(t, forward.bufferLineStartPos, 0)
}

func TestForward_RemoveLineFromBuffer(t *testing.T) {
	// given
	readerLineStartPos := 100
	bufferLineStartPos := 3
	lineSize := 6
	forward := NewForwardWithSize(strings.NewReader(""), 0, 4, 4)
	forward.buffer = []byte("ab\ncdefg\r\n")
	forward.readerLineStartPos = readerLineStartPos
	forward.bufferLineStartPos = bufferLineStartPos

	// when
	line := forward.removeLineFromBuffer(lineSize)

	// then
	assert.Equal(t, line, "cdefg")
	assert.Equal(t, forward.readerLineStartPos, readerLineStartPos+lineSize+1)
	assert.Equal(t, forward.bufferLineStartPos, bufferLineStartPos+lineSize+1)
}

func TestForward_Read(t *testing.T) {
	// given
	forward := NewForwardWithSize(strings.NewReader("abcd\nefgh\nijkl"), 0, 4, 14)

	// when
	err := forward.read()

	// then
	assert.Nil(t, err)
	assert.Equal(t, forward.chunk, []byte("abcd"))
	assert.Equal(t, cap(forward.chunk), 4)
	assert.Equal(t, forward.buffer, []byte("abcd"))
	assert.Equal(t, cap(forward.buffer), 4)
	assert.Equal(t, forward.readerPos, 4)
	assert.False(t, forward.endOfFile())

	// when
	err = forward.read()

	// then
	assert.Nil(t, err)
	assert.Equal(t, forward.chunk, []byte("\nefg"))
	assert.Equal(t, cap(forward.chunk), 4)
	assert.Equal(t, forward.buffer, []byte("abcd\nefg"))
	assert.Equal(t, cap(forward.buffer), 8)
	assert.Equal(t, forward.readerPos, 8)
	assert.False(t, forward.endOfFile())

	// when
	err = forward.read()

	// then
	assert.Nil(t, err)
	assert.Equal(t, forward.chunk, []byte("h\nij"))
	assert.Equal(t, cap(forward.chunk), 4)
	assert.Equal(t, forward.buffer, []byte("abcd\nefgh\nij"))
	assert.Equal(t, cap(forward.buffer), 12)
	assert.Equal(t, forward.readerPos, 12)
	assert.False(t, forward.endOfFile())

	// when
	err = forward.read()

	// then
	assert.Nil(t, err)
	assert.Equal(t, forward.chunk, []byte("kl"))
	assert.Equal(t, forward.buffer, []byte("abcd\nefgh\nijkl"))
	assert.Equal(t, cap(forward.buffer), 14)
	assert.True(t, forward.endOfFile())
}

func TestForward_Read_AllocateChunkError(t *testing.T) {
	// given
	readErr := errors.New("")
	reader := new(ReaderMock)
	reader.On("ReadAt", mock.Anything, mock.Anything).Return(0, readErr)
	forward := NewForward(reader, 10)

	// when
	err := forward.read()

	// then
	assert.Equal(t, err, readErr)
	assert.Equal(t, forward.readerPos, 10)
}

func TestForward_Read_AllocateChunkEndOfFile(t *testing.T) {
	// given
	data := "abcd\nefgh\nijkl"
	forward := NewForward(strings.NewReader("abcd\nefgh\nijkl"), 0)

	// when
	err := forward.read()

	// then
	assert.Nil(t, err)
	assert.Equal(t, forward.chunk, []byte(data))
	assert.Equal(t, forward.buffer, []byte(data))
	assert.True(t, forward.endOfFile())
}

func TestForward_Read_AllocateBufferError(t *testing.T) {
	// given
	data := "abcdefg"
	buffer := "wxyz"
	forward := NewForwardWithSize(strings.NewReader(data), 0, 4, len(buffer))
	forward.buffer = []byte(buffer)
	forward.readerPos = 6

	// when
	err := forward.read()

	// then
	assert.Equal(t, err, ErrBufferOverflow)
	assert.Equal(t, forward.chunk, []byte("g"))
	assert.Equal(t, forward.buffer, []byte(buffer))
	assert.True(t, forward.endOfFile())
}

func TestForward_Line(t *testing.T) {
	// given
	data := "a\nb\r\ncdef\ngh\ni"
	forward := NewForwardWithSize(strings.NewReader(data), 0, 4, 8)

	// when
	line, err := forward.Line()

	// then
	assert.Nil(t, err)
	assert.Equal(t, line, "a")
	assert.Equal(t, forward.chunk, []byte("a\nb\r"))
	assert.Equal(t, cap(forward.chunk), 4)
	assert.Equal(t, forward.buffer, []byte("a\nb\r"))
	assert.Equal(t, cap(forward.buffer), 4)
	assert.Equal(t, forward.readerPos, 4)
	assert.Equal(t, forward.readerLineStartPos, 2)
	assert.Equal(t, forward.bufferLineStartPos, 2)
	assert.False(t, forward.endOfFile())
	assert.False(t, forward.endOfScan())

	// when
	line, err = forward.Line()

	// then
	assert.Nil(t, err)
	assert.Equal(t, line, "b")
	assert.Equal(t, forward.chunk, []byte("\ncde"))
	assert.Equal(t, cap(forward.chunk), 4)
	assert.Equal(t, forward.buffer, []byte("b\r\ncde"))
	assert.Equal(t, cap(forward.buffer), 6)
	assert.Equal(t, forward.readerPos, 8)
	assert.Equal(t, forward.readerLineStartPos, 5)
	assert.Equal(t, forward.bufferLineStartPos, 3)
	assert.False(t, forward.endOfFile())
	assert.False(t, forward.endOfScan())

	// when
	line, err = forward.Line()

	// then
	assert.Nil(t, err)
	assert.Equal(t, line, "cdef")
	assert.Equal(t, forward.chunk, []byte("f\ngh"))
	assert.Equal(t, cap(forward.chunk), 4)
	assert.Equal(t, forward.buffer, []byte("cdef\ngh"))
	assert.Equal(t, cap(forward.buffer), 7)
	assert.Equal(t, forward.readerPos, 12)
	assert.Equal(t, forward.readerLineStartPos, 10)
	assert.Equal(t, forward.bufferLineStartPos, 5)
	assert.False(t, forward.endOfFile())
	assert.False(t, forward.endOfScan())

	// when
	line, err = forward.Line()

	// then
	assert.Nil(t, err)
	assert.Equal(t, line, "gh")
	assert.Equal(t, forward.chunk, []byte("\ni"))
	assert.Equal(t, cap(forward.chunk), 4)
	assert.Equal(t, forward.buffer, []byte("gh\ni"))
	assert.Equal(t, cap(forward.buffer), 7)
	assert.Equal(t, forward.readerPos, endPosition)
	assert.Equal(t, forward.readerLineStartPos, 13)
	assert.Equal(t, forward.bufferLineStartPos, 3)
	assert.True(t, forward.endOfFile())
	assert.False(t, forward.endOfScan())

	// when
	line, err = forward.Line()

	// then
	assert.Equal(t, err, io.EOF)
	assert.Equal(t, line, "i")
	assert.Equal(t, forward.chunk, []byte("\ni"))
	assert.Equal(t, cap(forward.chunk), 4)
	assert.Equal(t, forward.buffer, []byte("gh\ni"))
	assert.Equal(t, cap(forward.buffer), 7)
	assert.Equal(t, forward.readerPos, endPosition)
	assert.Equal(t, forward.readerLineStartPos, endPosition)
	assert.Equal(t, forward.bufferLineStartPos, 5)
	assert.True(t, forward.endOfFile())
	assert.True(t, forward.endOfScan())
}

func TestForward_Line_Error(t *testing.T) {
	// given
	readErr := errors.New("")
	reader := new(ReaderMock)
	reader.On("ReadAt", mock.Anything, mock.Anything).Return(0, readErr)
	forward := NewForward(reader, 0)

	// when
	line, err := forward.Line()

	// then
	assert.Equal(t, err, readErr)
	assert.Empty(t, line)

	// when
	line, err = forward.Line()

	// then
	assert.Equal(t, err, readErr)
	assert.Empty(t, line)
}

func TestForward_Line_AlreadyEndOfScan(t *testing.T) {
	// given
	forward := NewForward(strings.NewReader(""), endPosition)

	// when
	line, err := forward.Line()

	// then
	assert.Equal(t, err, io.EOF)
	assert.Equal(t, line, "")
	assert.True(t, forward.endOfFile())
	assert.True(t, forward.endOfScan())
}

func TestForward_Line_LineFeedOnly(t *testing.T) {
	// given
	data := "\n\r\n\n"
	forward := NewForward(strings.NewReader(data), 0)

	// when
	line, err := forward.Line()

	// then
	assert.Nil(t, err)
	assert.Empty(t, line)

	// when
	line, err = forward.Line()

	// then
	assert.Nil(t, err)
	assert.Empty(t, line)

	// when
	line, err = forward.Line()

	// then
	assert.Nil(t, err)
	assert.Empty(t, line)

	// when
	line, err = forward.Line()

	// then
	assert.Equal(t, err, io.EOF)
	assert.Empty(t, line)
	assert.True(t, forward.endOfFile())
	assert.True(t, forward.endOfScan())
}

func TestForward_Position(t *testing.T) {
	// given
	data := "abcdefgh\r\nhij"
	forward := NewForward(strings.NewReader(data), 0)

	// when
	line, err := forward.Line()

	// then
	assert.Nil(t, err)
	assert.Equal(t, line, "abcdefgh")
	assert.Equal(t, forward.Position(), 10)

	// given
	forward = NewForward(strings.NewReader(data), forward.Position())

	// when
	line, err = forward.Line()

	// then
	assert.Equal(t, err, io.EOF)
	assert.Equal(t, line, "hij")
	assert.Equal(t, forward.Position(), endPosition)
}
