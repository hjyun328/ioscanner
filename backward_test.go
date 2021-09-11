package linescanner

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io"
	"strings"
	"testing"
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
	scanner := newBackward(reader, position)

	// then
	assert.Equal(t, scanner.reader, reader)
	assert.Equal(t, scanner.readerPos, position)
	assert.Equal(t, scanner.readerLineEndPos, position)
	assert.Equal(t, scanner.maxChunkSize, defaultChunkSize)
	assert.Equal(t, scanner.maxBufferSize, defaultBufferSize)
}

func TestBackward_NewBackward_ErrNilReader(t *testing.T) {
	assert.PanicsWithValue(t, ErrNilReader, func() {
		newBackward(nil, endPosition)
	})
}

func TestBackward_NewBackward_ErrInvalidMaxChunkSize(t *testing.T) {
	assert.PanicsWithValue(t, ErrInvalidChunkSize, func() {
		newBackwardWithSize(strings.NewReader(""), endPosition, 0, 100)
	})
}

func TestBackward_NewBackward_ErrInvalidMaxBufferSize(t *testing.T) {
	assert.PanicsWithValue(t, ErrInvalidBufferSize, func() {
		newBackwardWithSize(strings.NewReader(""), endPosition, 100, 0)
	})
}

func TestBackward_NewBackward_ErrGreatorBufferSize(t *testing.T) {
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
	scanner := newBackwardWithSize(reader, position, maxChunkSize, maxBufferSize)

	// then
	assert.Equal(t, scanner.reader, reader)
	assert.Equal(t, scanner.readerPos, position)
	assert.Equal(t, scanner.readerLineEndPos, position)
	assert.Equal(t, scanner.maxChunkSize, maxChunkSize)
	assert.Equal(t, scanner.maxBufferSize, maxBufferSize)
}

func TestBackward_BackupPostiion(t *testing.T) {
	// given
	scanner := newBackward(strings.NewReader(""), 0)
	scanner.readerPos = 1
	scanner.readerLineEndPos = 2

	// when
	scanner.backupPosition()

	// then
	assert.Equal(t, scanner.readerPos, scanner.backupReaderPos)
}

func TestBackward_RecoverPosition(t *testing.T) {
	// given
	scanner := newBackward(strings.NewReader(""), 0)
	scanner.readerPos = 1

	// when
	scanner.backupPosition()

	// then
	assert.Equal(t, scanner.readerPos, scanner.backupReaderPos)

	// given
	scanner.readerPos = 10

	// when
	scanner.recoverPosition()

	// then
	assert.Equal(t, scanner.readerPos, 1)
}

func TestBackward_EndOfFile_False(t *testing.T) {
	// given
	scanner := newBackward(strings.NewReader(""), 0)

	// when
	scanner.readerPos = 1

	// then
	assert.False(t, scanner.endOfFile())
}

func TestBackward_EndOfFile_True(t *testing.T) {
	// given
	scanner := newBackward(strings.NewReader(""), 0)

	// when
	scanner.readerPos = -1

	// then
	assert.True(t, scanner.endOfFile())

	// when
	scanner.readerPos = 0

	// then
	assert.True(t, scanner.endOfFile())
}

func TestBackward_EndOfScan_False(t *testing.T) {
	// given
	scanner := newBackward(strings.NewReader(""), 0)

	// when
	scanner.readerPos = -1
	scanner.readerLineEndPos = 1

	// then
	assert.False(t, scanner.endOfScan())
}

func TestBackward_EndOfScan_True(t *testing.T) {
	// given
	scanner := newBackward(strings.NewReader(""), 0)

	// when
	scanner.readerPos = -1
	scanner.readerLineEndPos = -1

	// then
	assert.True(t, scanner.endOfScan())
}

func TestBackward_AllocateChunk_GreatorThanReaderPosWhenFirstAllocated(t *testing.T) {
	// given
	scanner := newBackwardWithSize(strings.NewReader(""), 0, 4, 4)
	scanner.readerPos = 2

	// when
	scanner.allocateChunk()

	// then
	assert.Equal(t, len(scanner.chunk), scanner.readerPos)
	assert.Equal(t, cap(scanner.chunk), scanner.readerPos)
}

func TestBackward_AllocateChunk_GreatorThanReaderPosWhenAlreadyAllocated(t *testing.T) {
	// given
	scanner := newBackwardWithSize(strings.NewReader(""), 0, 4, 4)
	scanner.readerPos = 6

	// when
	scanner.allocateChunk()

	// then
	assert.Equal(t, len(scanner.chunk), scanner.maxChunkSize)
	assert.Equal(t, cap(scanner.chunk), scanner.maxChunkSize)

	// given
	scanner.readerPos = 2

	// when
	scanner.allocateChunk()

	// then
	assert.Equal(t, len(scanner.chunk), scanner.readerPos)
	assert.Equal(t, cap(scanner.chunk), scanner.maxChunkSize)
}

func TestBackward_AllocateChunk_LessThanReaderPos(t *testing.T) {
	// given
	scanner := newBackwardWithSize(strings.NewReader(""), 0, 4, 4)
	scanner.readerPos = 6

	// when
	scanner.allocateChunk()

	// then
	assert.Equal(t, len(scanner.chunk), scanner.maxChunkSize)
	assert.Equal(t, cap(scanner.chunk), scanner.maxChunkSize)
}

func TestBackward_AllocateBuffer_FirstAllocation(t *testing.T) {
	// given
	chunk := []byte("abcd")
	scanner := newBackwardWithSize(strings.NewReader(""), 0, len(chunk), len(chunk))
	scanner.chunk = chunk

	// when
	err := scanner.allocateBuffer()

	// then
	assert.Nil(t, err)
	assert.Equal(t, scanner.buffer, scanner.chunk)
	assert.Equal(t, cap(scanner.buffer), len(scanner.chunk))
}

func TestBackward_AllocateBuffer_BufferOverflow(t *testing.T) {
	// given
	chunk := []byte("abcd")
	buffer := make([]byte, 1, len(chunk))
	scanner := newBackwardWithSize(strings.NewReader(""), 0, len(chunk), cap(buffer))
	scanner.chunk = chunk
	scanner.buffer = buffer

	// when
	err := scanner.allocateBuffer()

	// then
	assert.Equal(t, err, ErrBufferOverflow)
}

func TestBackward_AllocateBuffer_BufferExpanded(t *testing.T) {
	// given
	chunk := []byte("abcd")
	buffer := make([]byte, 1, len(chunk)+1)
	buffer[0] = 'e'
	scanner := newBackwardWithSize(strings.NewReader(""), 0, len(chunk), cap(buffer))
	scanner.chunk = chunk
	scanner.buffer = buffer

	// when
	err := scanner.allocateBuffer()

	// then
	assert.Nil(t, err)
	assert.Equal(t, scanner.buffer, []byte("abcde"))
	assert.Equal(t, cap(scanner.buffer), len(scanner.buffer))
}

func TestBackward_AllocateBuffer_BufferReused(t *testing.T) {
	// given
	chunk := []byte("abcd")
	buffer := make([]byte, 1, 10)
	buffer[0] = 'e'
	scanner := newBackwardWithSize(strings.NewReader(""), 0, len(chunk), cap(buffer))
	scanner.chunk = chunk
	scanner.buffer = buffer

	// when
	err := scanner.allocateBuffer()

	// then
	assert.Nil(t, err)
	assert.Equal(t, scanner.buffer, []byte("abcde"))
	assert.Equal(t, cap(scanner.buffer), 10)
}

func TestBackward_RemoveLineFromBuffer(t *testing.T) {
	// given
	scanner := newBackwardWithSize(strings.NewReader(""), 16, 4, 8)
	scanner.buffer = []byte("a\r\ndefg\r")

	// when
	line := scanner.removeLineFromBuffer(2)

	// then
	assert.Equal(t, line, "defg")
	assert.Equal(t, len(scanner.buffer), 2)
	assert.Equal(t, cap(scanner.buffer), 8)
	assert.Equal(t, scanner.readerLineEndPos, 16-(len(line) /* line feed */ +1 /* carrage return */ +1))
}

func TestBackward_Read(t *testing.T) {
	// given
	data := "abcd\nefgh\nijkl"
	scanner := newBackwardWithSize(strings.NewReader(data), len(data), 4, 14)

	// when
	err := scanner.read()

	// then
	assert.Nil(t, err)
	assert.Equal(t, scanner.chunk, []byte("ijkl"))
	assert.Equal(t, cap(scanner.chunk), 4)
	assert.Equal(t, scanner.buffer, []byte("ijkl"))
	assert.Equal(t, cap(scanner.buffer), 4)
	assert.Equal(t, scanner.readerPos, 10)
	assert.False(t, scanner.endOfFile())

	// when
	err = scanner.read()

	// then
	assert.Nil(t, err)
	assert.Equal(t, scanner.chunk, []byte("fgh\n"))
	assert.Equal(t, cap(scanner.chunk), 4)
	assert.Equal(t, scanner.buffer, []byte("fgh\nijkl"))
	assert.Equal(t, cap(scanner.buffer), 8)
	assert.Equal(t, scanner.readerPos, 6)
	assert.False(t, scanner.endOfFile())

	// when
	err = scanner.read()

	// then
	assert.Nil(t, err)
	assert.Equal(t, scanner.chunk, []byte("cd\ne"))
	assert.Equal(t, cap(scanner.chunk), 4)
	assert.Equal(t, scanner.buffer, []byte("cd\nefgh\nijkl"))
	assert.Equal(t, cap(scanner.buffer), 12)
	assert.Equal(t, scanner.readerPos, 2)
	assert.False(t, scanner.endOfFile())

	// when
	err = scanner.read()

	// then
	assert.Nil(t, err)
	assert.Equal(t, scanner.chunk, []byte("ab"))
	assert.Equal(t, cap(scanner.chunk), 4)
	assert.Equal(t, scanner.buffer, []byte("abcd\nefgh\nijkl"))
	assert.Equal(t, cap(scanner.buffer), 14)
	assert.Equal(t, scanner.readerPos, 0)
	assert.True(t, scanner.endOfFile())
}

func TestBackward_Read_WithPosition(t *testing.T) {
	// given
	data := "abcd\nefgh\nijkl"
	scanner := newBackwardWithSize(strings.NewReader(data), len(data)-2, 4, 14)

	// when
	err := scanner.read()

	// then
	assert.Nil(t, err)
	assert.Equal(t, scanner.chunk, []byte("h\nij"))
	assert.Equal(t, cap(scanner.chunk), 4)
	assert.Equal(t, scanner.buffer, []byte("h\nij"))
	assert.Equal(t, cap(scanner.buffer), 4)
	assert.Equal(t, scanner.readerPos, 8)
	assert.False(t, scanner.endOfFile())
}

func TestBackward_Read_InvalidPosition(t *testing.T) {
	// given
	data := "abcd\nefgh\nijkl"
	scanner := newBackwardWithSize(strings.NewReader(data), len(data)+1, 4, 14)

	// when
	err := scanner.read()

	// then
	assert.Equal(t, err, ErrInvalidReaderPosition)
}

func TestBackward_Read_ReadFailure(t *testing.T) {
	// given
	readErr := errors.New("")
	reader := new(ReaderMock)
	reader.On("ReadAt", mock.Anything, mock.Anything).Return(0, readErr)
	scanner := newBackward(reader, 10)

	// when
	err := scanner.read()

	// then
	assert.Equal(t, err, readErr)
}

func TestBackward_Read_ReadFailureChunkSize(t *testing.T) {
	// given
	position := 20
	readSize := 10
	reader := new(ReaderMock)
	reader.On("ReadAt", mock.Anything, mock.Anything).Return(readSize, nil)
	scanner := newBackward(reader, position)

	// when
	err := scanner.read()

	// then
	assert.Equal(t, err, ErrReadFailureChunkSize)
	assert.Equal(t, scanner.readerPos, position)
}

func TestBackward_Read_BufferOverflow(t *testing.T) {
	// given
	chunk := []byte("abcd")
	buffer := make([]byte, 1, len(chunk))
	scanner := newBackwardWithSize(strings.NewReader("abcd"), 4, len(chunk), cap(buffer))
	scanner.chunk = chunk
	scanner.buffer = buffer

	// when
	err := scanner.read()

	// then
	assert.Equal(t, err, ErrBufferOverflow)
}

func TestBackward_Line(t *testing.T) {
	// given
	data := "a\nb\r\ncdef\nghij"
	scanner := newBackwardWithSize(strings.NewReader(data), len(data), 4, 8)

	// when
	line, err := scanner.Line()

	// then
	assert.Nil(t, err)
	assert.Equal(t, line, "ghij")
	assert.Equal(t, scanner.chunk, []byte("def\n"))
	assert.Equal(t, cap(scanner.chunk), 4)
	assert.Equal(t, scanner.buffer, []byte("def"))
	assert.Equal(t, cap(scanner.buffer), 8)
	assert.Equal(t, scanner.readerPos, 6)
	assert.Equal(t, scanner.readerLineEndPos, 9)
	assert.False(t, scanner.endOfFile())
	assert.False(t, scanner.endOfScan())

	// when
	line, err = scanner.Line()

	// then
	assert.Nil(t, err)
	assert.Equal(t, line, "cdef")
	assert.Equal(t, scanner.chunk, []byte("b\r\nc"))
	assert.Equal(t, cap(scanner.chunk), 4)
	assert.Equal(t, scanner.buffer, []byte("b\r"))
	assert.Equal(t, cap(scanner.buffer), 8)
	assert.Equal(t, scanner.readerPos, 2)
	assert.Equal(t, scanner.readerLineEndPos, 4)
	assert.False(t, scanner.endOfFile())
	assert.False(t, scanner.endOfScan())

	// when
	line, err = scanner.Line()

	// then
	assert.Nil(t, err)
	assert.Equal(t, line, "b")
	assert.Equal(t, scanner.chunk, []byte("a\n"))
	assert.Equal(t, cap(scanner.chunk), 4)
	assert.Equal(t, scanner.buffer, []byte("a"))
	assert.Equal(t, cap(scanner.buffer), 8)
	assert.Equal(t, scanner.readerPos, 0)
	assert.Equal(t, scanner.readerLineEndPos, 1)
	assert.True(t, scanner.endOfFile())
	assert.False(t, scanner.endOfScan())

	// when
	line, err = scanner.Line()

	// then
	assert.Equal(t, err, io.EOF)
	assert.Equal(t, line, "a")
	assert.Equal(t, scanner.chunk, []byte("a\n"))
	assert.Equal(t, cap(scanner.chunk), 4)
	assert.Empty(t, scanner.buffer)
	assert.Equal(t, cap(scanner.buffer), 8)
	assert.Equal(t, scanner.readerPos, 0)
	assert.Equal(t, scanner.readerLineEndPos, 0)
	assert.True(t, scanner.endOfFile())
	assert.True(t, scanner.endOfScan())
}

func TestBackward_Line_Error(t *testing.T) {
	// given
	data := "abcdefgh\nhij"
	scanner := newBackwardWithSize(strings.NewReader(data), len(data), 4, 5)

	// when
	line, err := scanner.Line()

	// then
	assert.Nil(t, err)
	assert.Equal(t, line, "hij")
	assert.Equal(t, scanner.chunk, []byte("\nhij"))
	assert.Equal(t, cap(scanner.chunk), 4)
	assert.Equal(t, scanner.buffer, []byte(""))
	assert.Equal(t, cap(scanner.buffer), 4)
	assert.Equal(t, scanner.readerPos, len(data)-scanner.maxChunkSize)
	assert.Equal(t, scanner.readerLineEndPos, len(data)-scanner.maxChunkSize)
	assert.False(t, scanner.endOfFile())
	assert.False(t, scanner.endOfScan())

	// when
	line, err = scanner.Line()

	// then
	assert.Equal(t, err, ErrBufferOverflow)
	assert.Equal(t, line, "")
	assert.Equal(t, scanner.chunk, []byte("abcd"))
	assert.Equal(t, cap(scanner.chunk), 4)
	assert.Equal(t, scanner.buffer, []byte("efgh"))
	assert.Equal(t, cap(scanner.buffer), 4)
	assert.Equal(t, scanner.readerPos, len(data)-scanner.maxChunkSize)
	assert.Equal(t, scanner.readerLineEndPos, len(data)-scanner.maxChunkSize)
	assert.False(t, scanner.endOfFile())
	assert.False(t, scanner.endOfScan())
}

func TestBackward_Line_AlreadyEndOfScan(t *testing.T) {
	// given
	scanner := newBackward(strings.NewReader(""), endPosition)

	// when
	line, err := scanner.Line()

	// then
	assert.Equal(t, err, io.EOF)
	assert.Equal(t, line, "")
}

func TestBackward_Line_LineFeedOnly(t *testing.T) {
	// given
	data := "\n\r\n\n"
	scanner := newBackward(strings.NewReader(data), len(data))

	// when
	line, err := scanner.Line()

	// then
	assert.Nil(t, err)
	assert.Empty(t, line)

	// when
	line, err = scanner.Line()

	// then
	assert.Nil(t, err)
	assert.Empty(t, line)

	// when
	line, err = scanner.Line()

	// then
	assert.Nil(t, err)
	assert.Empty(t, line)

	// when
	line, err = scanner.Line()

	// then
	assert.Equal(t, err, io.EOF)
	assert.Empty(t, line)
}

func TestBackward_Position(t *testing.T) {
	// given
	data := "abcdefgh\r\nhij"
	scanner := newBackward(strings.NewReader(data), len(data))

	// when
	line, err := scanner.Line()

	// then
	assert.Nil(t, err, nil)
	assert.Equal(t, line, "hij")

	// given
	scanner = newBackward(strings.NewReader(data), scanner.Position())

	// when
	line, err = scanner.Line()

	// then
	assert.Equal(t, err, io.EOF)
	assert.Equal(t, line, "abcdefgh")
	assert.Equal(t, scanner.Position(), endPosition)
}
