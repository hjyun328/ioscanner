package linescanner

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestForward_NewForward(t *testing.T) {
	// given
	reader := strings.NewReader("")
	position := 100

	// when
	scanner := newForward(reader, position)

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
		newBackward(nil, endPosition)
	})
}

func TestForward_NewForward_ErrInvalidMaxChunkSize(t *testing.T) {
	assert.PanicsWithValue(t, ErrInvalidMaxChunkSize, func() {
		newForwardWithSize(strings.NewReader(""), endPosition, 0, 100)
	})
}

func TestForward_NewForward_ErrInvalidMaxBufferSize(t *testing.T) {
	assert.PanicsWithValue(t, ErrInvalidMaxBufferSize, func() {
		newForwardWithSize(strings.NewReader(""), endPosition, 100, 0)
	})
}

func TestForward_NewForward_ErrGreaterBufferSize(t *testing.T) {
	assert.PanicsWithValue(t, ErrGreaterBufferSize, func() {
		newForwardWithSize(strings.NewReader(""), endPosition, 100, 10)
	})
}

func TestForward_BackupPosition(t *testing.T) {
	// given
	forward := newForward(strings.NewReader(""), 0)
	forward.readerPos = 1
	forward.readerLineStartPos = 2

	// when
	forward.backupPosition()

	// then
	assert.Equal(t, forward.readerPos, forward.backupReaderPos)
	assert.Equal(t, forward.readerLineStartPos, forward.backupReaderLineStartPos)
}

func TestForward_RecoverPosition(t *testing.T) {
	// given
	forward := newForward(strings.NewReader(""), 0)
	forward.readerPos = 2
	forward.readerLineStartPos = 3

	// when
	forward.backupPosition()

	// then
	assert.Equal(t, forward.readerPos, forward.backupReaderPos)
	assert.Equal(t, forward.readerLineStartPos, forward.backupReaderLineStartPos)

	// given
	forward.readerPos = 4
	forward.readerLineStartPos = 5

	// when
	forward.recoverPosition()

	// then
	assert.Equal(t, forward.readerPos, forward.backupReaderPos)
	assert.Equal(t, forward.readerLineStartPos, forward.backupReaderLineStartPos)
}

func TestForward_EndOfFile_False(t *testing.T) {
	// given
	forward := newForward(strings.NewReader(""), 0)

	// when
	forward.readerPos = 0

	// then
	assert.False(t, forward.endOfFile())
}

func TestForward_EndOfFile_True(t *testing.T) {
	// given
	forward := newForward(strings.NewReader(""), 0)

	// when
	forward.readerPos = endPosition

	// then
	assert.True(t, forward.endOfFile())
}

func TestForward_EndOfScan_False(t *testing.T) {
	// given
	forward := newForward(strings.NewReader(""), 0)

	// when
	forward.readerPos = -1
	forward.readerLineStartPos = 0

	// then
	assert.False(t, forward.endOfScan())
}

func TestForward_EndOfScan_True(t *testing.T) {
	// given
	forward := newForward(strings.NewReader(""), 0)

	// when
	forward.readerPos = endPosition
	forward.readerLineStartPos = endPosition

	// then
	assert.True(t, forward.endOfScan())
}

/*
func TestArrangeBuffer(t *testing.T) {
	// given
	forward := newForwardWithSize(strings.NewReader(""), 0, 4, 7)
	forward.buffer = append(forward.buffer, "abcdefg"...)
	forward.bufferLineStartPos = 3

	// when
	err := forward.arrangeBuffer(3)

	// then
	assert.Nil(t, err)
	assert.Equal(t, forward.buffer, []byte("defg"))
	assert.Equal(t, forward.bufferLineStartPos, 0)
}

func TestArrangeBuffer_BufferOverflow(t *testing.T) {
	// given
	forward := newForwardWithSize(strings.NewReader(""), 0, 4, 7)
	forward.buffer = append(forward.buffer, "abcdefg"...)
	forward.bufferLineStartPos = 3

	// when
	err := forward.arrangeBuffer(4)

	// then
	assert.Equal(t, err, ErrBufferOverflow)
	assert.Equal(t, forward.buffer, []byte("abcdefg"))
	assert.Equal(t, forward.bufferLineStartPos, 3)
}

func TestArrangeBuffer_NotArranged(t *testing.T) {
	// given
	forward := newForwardWithSize(strings.NewReader(""), 0, 4, 7)
	forward.buffer = append(forward.buffer, "abcde"...)
	forward.bufferLineStartPos = 3

	// when
	err := forward.arrangeBuffer(2)

	// then
	assert.Nil(t, err)
	assert.Equal(t, forward.buffer, []byte("abcde"))
	assert.Equal(t, forward.bufferLineStartPos, 3)
}

func TestRead(t *testing.T) {
	// given
	forward := newForwardWithSize(strings.NewReader("abcdefghijklmnop"), 0, 4, 8)

	// when
	err := forward.read()

	// then
	assert.Nil(t, err)
	assert.Equal(t, forward.chunk, []byte("abcd"))
	assert.Equal(t, forward.buffer, []byte("abcd"))
	assert.Equal(t, forward.bufferLineStartPos, 0)
	assert.Equal(t, forward.readerPos, 4)
	assert.False(t, forward.endOfFile)

	// when
	err = forward.read()

	// then
	assert.Nil(t, err)
	assert.Equal(t, forward.chunk, []byte("efgh"))
	assert.Equal(t, forward.buffer, []byte("abcdefgh"))
	assert.Equal(t, forward.bufferLineStartPos, 0)
	assert.Equal(t, forward.readerPos, 8)
	assert.False(t, forward.endOfFile)
}

func TestRead_EndOfFile(t *testing.T) {
	// given
	forward := newForwardWithSize(strings.NewReader("abc"), 0, 4, 7)

	// when
	err := forward.read()

	// then
	assert.Nil(t, err)
	assert.Equal(t, forward.chunk, append([]byte("abc"), 0x00))
	assert.Equal(t, forward.buffer, []byte("abc"))
	assert.Equal(t, forward.bufferLineStartPos, 0)
	assert.Equal(t, forward.readerPos, endPosition)
	assert.True(t, forward.endOfFile)
}

func TestRead_Arranged(t *testing.T) {
	// given
	forward := newForwardWithSize(strings.NewReader("abcdefghijk"), 7, 3, 7)
	forward.buffer = append(forward.buffer, "abcdefg"...)
	forward.bufferLineStartPos = 3

	// when
	err := forward.read()

	// then
	assert.Nil(t, err)
	assert.Equal(t, forward.chunk, []byte("hij"))
	assert.Equal(t, forward.buffer, []byte("defghij"))
	assert.Equal(t, forward.bufferLineStartPos, 0)
	assert.Equal(t, forward.readerPos, 10)
	assert.False(t, forward.endOfFile)
}

func TestRead_BufferOverflow(t *testing.T) {
	// given
	forward := newForwardWithSize(strings.NewReader("abcdefghijk"), 7, 4, 7)
	forward.buffer = append(forward.buffer, "abcdefg"...)
	forward.bufferLineStartPos = 3

	// when
	err := forward.read()

	// then
	assert.Equal(t, err, ErrBufferOverflow)
	assert.Equal(t, forward.chunk, []byte("hijk"))
	assert.Equal(t, forward.buffer, []byte("abcdefg"))
	assert.Equal(t, forward.bufferLineStartPos, 3)
	assert.Equal(t, forward.readerPos, 7)
	assert.False(t, forward.endOfFile)
}

func TestLine(t *testing.T) {
	// given
	reader := strings.NewReader("ab\ncdefg")
	forward := newForwardWithSize(reader, 0, 4, 5)

	// when
	line, err := forward.Line()

	// then
	assert.Nil(t, err)
	assert.Equal(t, line, "ab")
	assert.Equal(t, forward.Position(), 3)
	assert.Equal(t, forward.Position(), forward.readerLineStartPos)
	assert.Equal(t, forward.bufferLineStartPos, 3)
	assert.Equal(t, forward.readerPos, 4)
	assert.False(t, forward.endOfFile)
	assert.False(t, forward.endOfScan)
}

func TestLine_EndOfFileButNotEndOfScan(t *testing.T) {
	// given
	reader := strings.NewReader("ab\ncdefg\nhijk")
	forward := newForward(reader, 0)

	// when
	line, err := forward.Line()

	// then
	assert.Nil(t, err)
	assert.Equal(t, line, "ab")
	assert.Equal(t, forward.Position(), 3)
	assert.Equal(t, forward.Position(), forward.readerLineStartPos)
	assert.Equal(t, forward.bufferLineStartPos, 3)
	assert.Equal(t, forward.readerPos, endPosition)
	assert.True(t, forward.endOfFile)
	assert.False(t, forward.endOfScan)

	// when
	line, err = forward.Line()

	// then
	assert.Nil(t, err)
	assert.Equal(t, line, "cdefg")
	assert.Equal(t, forward.Position(), 9)
	assert.Equal(t, forward.Position(), forward.readerLineStartPos)
	assert.Equal(t, forward.bufferLineStartPos, 9)
	assert.Equal(t, forward.readerPos, endPosition)
	assert.True(t, forward.endOfFile)
	assert.False(t, forward.endOfScan)
}

func TestLine_BufferOverflow(t *testing.T) {
	// given
	reader := strings.NewReader("ab\ncdefg")
	forward := newForwardWithSize(reader, 0, 4, 4)

	// when
	line, err := forward.Line()

	// then
	assert.Nil(t, err)
	assert.Equal(t, line, "ab")
	assert.Equal(t, forward.Position(), 3)
	assert.Equal(t, forward.Position(), forward.readerLineStartPos)
	assert.Equal(t, forward.bufferLineStartPos, 3)
	assert.Equal(t, forward.readerPos, 4)
	assert.False(t, forward.endOfFile)
	assert.False(t, forward.endOfScan)

	// when
	line, err = forward.Line()

	// then
	assert.Equal(t, err, ErrBufferOverflow)
	assert.Empty(t, line)
	assert.Equal(t, forward.Position(), 3)
	assert.Equal(t, forward.Position(), forward.readerLineStartPos)
	assert.Equal(t, forward.bufferLineStartPos, 3)
	assert.Equal(t, forward.readerPos, 4)
	assert.False(t, forward.endOfFile)
	assert.False(t, forward.endOfScan)
}

func TestLine_WithEmpty(t *testing.T) {
	// given
	reader := strings.NewReader("")
	forward := newForwardWithSize(reader, 0, 4, 4)

	// when
	line, err := forward.Line()

	// then
	assert.Equal(t, err, io.EOF)
	assert.Empty(t, line)
	assert.Equal(t, forward.Position(), endPosition)
	assert.Equal(t, forward.Position(), forward.readerLineStartPos)
	assert.Equal(t, forward.bufferLineStartPos, endPosition)
	assert.Equal(t, forward.readerPos, endPosition)
	assert.True(t, forward.endOfFile)
	assert.True(t, forward.endOfScan)
}

func TestLine_WithTextOnly(t *testing.T) {
	// given
	reader := strings.NewReader("abcdefg")
	forward := newForward(reader, 0)

	// when
	lines, err := forward.Line()

	// then
	assert.Equal(t, err, io.EOF)
	assert.Equal(t, lines, "abcdefg")
	assert.Equal(t, forward.Position(), endPosition)
	assert.Equal(t, forward.Position(), forward.readerLineStartPos)
	assert.Equal(t, forward.bufferLineStartPos, endPosition)
	assert.Equal(t, forward.readerPos, endPosition)
	assert.True(t, forward.endOfFile)
	assert.True(t, forward.endOfScan)
}

func TestLine_WithLinesFeedOnly(t *testing.T) {
	// given
	reader := strings.NewReader("\n\n")
	forward := newForward(reader, 0)

	// when
	line, err := forward.Line()

	// then
	assert.Nil(t, err)
	assert.Equal(t, line, "")
	assert.Equal(t, forward.Position(), 1)
	assert.Equal(t, forward.Position(), forward.readerLineStartPos)
	assert.Equal(t, forward.bufferLineStartPos, 1)
	assert.Equal(t, forward.readerPos, endPosition)
	assert.True(t, forward.endOfFile)
	assert.False(t, forward.endOfScan)

	// when
	line, err = forward.Line()

	// then
	assert.Nil(t, err)
	assert.Equal(t, line, "")
	assert.Equal(t, forward.Position(), 2)
	assert.Equal(t, forward.Position(), forward.readerLineStartPos)
	assert.Equal(t, forward.bufferLineStartPos, 2)
	assert.Equal(t, forward.readerPos, endPosition)
	assert.True(t, forward.endOfFile)
	assert.False(t, forward.endOfScan)

	// when
	line, err = forward.Line()

	// then
	assert.Equal(t, err, io.EOF)
	assert.Equal(t, line, "")
	assert.Equal(t, forward.Position(), endPosition)
	assert.Equal(t, forward.Position(), forward.readerLineStartPos)
	assert.Equal(t, forward.bufferLineStartPos, endPosition)
	assert.Equal(t, forward.readerPos, endPosition)
	assert.True(t, forward.endOfFile)
	assert.True(t, forward.endOfScan)
}

func TestLine_WithCarrageReturn(t *testing.T) {
	// given
	reader := strings.NewReader("ab\r\ncdefg")
	forward := newForwardWithSize(reader, 0, 4, 5)

	// when
	line, err := forward.Line()

	// then
	assert.Nil(t, err)
	assert.Equal(t, line, "ab")
	assert.Equal(t, forward.Position(), 4)
	assert.Equal(t, forward.Position(), forward.readerLineStartPos)
	assert.Equal(t, forward.bufferLineStartPos, 4)
	assert.Equal(t, forward.readerPos, 4)
	assert.False(t, forward.endOfFile)
	assert.False(t, forward.endOfScan)

	// when
	line, err = forward.Line()

	// then
	assert.Equal(t, err, io.EOF)
	assert.Equal(t, line, "cdefg")
	assert.Equal(t, forward.Position(), endPosition)
	assert.Equal(t, forward.Position(), forward.readerLineStartPos)
	assert.Equal(t, forward.bufferLineStartPos, endPosition)
	assert.Equal(t, forward.readerPos, endPosition)
	assert.True(t, forward.endOfFile)
	assert.True(t, forward.endOfScan)
}

func TestLine_WithPosition(t *testing.T) {
	// given
	reader := strings.NewReader("ab\ncdefg")
	forward := newForward(reader, 3)

	// when
	line, err := forward.Line()

	// then
	assert.Equal(t, err, io.EOF)
	assert.Equal(t, line, "cdefg")
	assert.Equal(t, forward.Position(), endPosition)
	assert.Equal(t, forward.Position(), forward.readerLineStartPos)
	assert.Equal(t, forward.bufferLineStartPos, endPosition)
	assert.Equal(t, forward.readerPos, endPosition)
	assert.True(t, forward.endOfFile)
	assert.True(t, forward.endOfScan)
}

func TestLine_WithExceedPosition(t *testing.T) {
	// given
	reader := strings.NewReader("ab\ncdefg")
	forward := newForward(reader, 100)

	// when
	line, err := forward.Line()

	// then
	assert.Equal(t, err, io.EOF)
	assert.Empty(t, line, "")
	assert.Equal(t, forward.Position(), endPosition)
	assert.Equal(t, forward.Position(), forward.readerLineStartPos)
	assert.Equal(t, forward.bufferLineStartPos, endPosition)
	assert.Equal(t, forward.readerPos, endPosition)
	assert.True(t, forward.endOfFile)
	assert.True(t, forward.endOfScan)
}

func TestLine_WithReaderLineStartPosition(t *testing.T) {
	// given
	reader := strings.NewReader("ab\ncdefg\nhijk")
	forward := newForward(reader, 0)

	// when
	line, err := forward.Line()

	// then
	assert.Nil(t, err)
	assert.Equal(t, line, "ab")
	assert.Equal(t, forward.Position(), 3)
	assert.Equal(t, forward.Position(), forward.readerLineStartPos)
	assert.Equal(t, forward.bufferLineStartPos, 3)
	assert.Equal(t, forward.readerPos, endPosition)
	assert.True(t, forward.endOfFile)
	assert.False(t, forward.endOfScan)

	// when
	forward = newForward(reader, forward.Position())
	line, err = forward.Line()

	// then
	assert.Nil(t, err)
	assert.Equal(t, line, "cdefg")
	assert.Equal(t, forward.Position(), 9)
	assert.Equal(t, forward.Position(), forward.readerLineStartPos)
	assert.Equal(t, forward.bufferLineStartPos, 6)
	assert.Equal(t, forward.readerPos, endPosition)
	assert.True(t, forward.endOfFile)
	assert.False(t, forward.endOfScan)
}
*/
