package linescanner

import (
	"github.com/stretchr/testify/assert"
	"io"
	"strings"
	"testing"
)

func TestNewForward_InvalidPosition(t *testing.T) {
	assert.PanicsWithValue(t, ErrInvalidPosition, func() {
		newForward(strings.NewReader(""), -1)
	})
}

func TestNewForward_InvalidReader(t *testing.T) {
	assert.PanicsWithValue(t, ErrNilReader, func() {
		newForward(nil, -1)
	})
}

func TestNewForwardWithSize_InvalidChunkSize(t *testing.T) {
	assert.PanicsWithValue(t, ErrInvalidChunkSize, func() {
		newForwardWithSize(strings.NewReader(""), 0, -1, 100)
	})
	assert.PanicsWithValue(t, ErrInvalidChunkSize, func() {
		newForwardWithSize(strings.NewReader(""), 0, 0, 100)
	})
}

func TestNewForwardWithSize_InvalidBufferSize(t *testing.T) {
	assert.PanicsWithValue(t, ErrInvalidBufferSize, func() {
		newForwardWithSize(strings.NewReader(""), 0, 1, 0)
	})
	assert.PanicsWithValue(t, ErrInvalidBufferSize, func() {
		newForwardWithSize(strings.NewReader(""), 0, 1, -1)
	})
}

func TestNewForwardWithSize_GreaterBufferSize(t *testing.T) {
	assert.PanicsWithValue(t, ErrGreaterBufferSize, func() {
		newForwardWithSize(strings.NewReader(""), 0, 2, 1)
	})
}

func TestBackupPosition(t *testing.T) {
	// given
	forward := newForward(strings.NewReader(""), 0)
	forward.bufferLineStartPos = 1
	forward.readerPos = 2
	forward.readerLineStartPos = 3

	// when
	forward.backupPosition()

	// then
	assert.Equal(t, forward.bufferLineStartPos, forward.backupBufferLineStartPos)
	assert.Equal(t, forward.readerPos, forward.backupReaderPos)
	assert.Equal(t, forward.readerLineStartPos, forward.backupReaderLineStartPos)
}

func TestRecoverPosition(t *testing.T) {
	// given
	forward := newForward(strings.NewReader(""), 0)
	forward.bufferLineStartPos = 1
	forward.readerPos = 2
	forward.readerLineStartPos = 3
	forward.backupPosition()
	forward.bufferLineStartPos = 4
	forward.readerPos = 5
	forward.readerLineStartPos = 6

	// when
	forward.recoverPosition()

	// then
	assert.Equal(t, forward.bufferLineStartPos, 1)
	assert.Equal(t, forward.readerPos, 2)
	assert.Equal(t, forward.readerLineStartPos, 3)
}

func TestEndPosition(t *testing.T) {
	// given
	forward := newForward(strings.NewReader(""), 0)
	forward.bufferLineStartPos = 1
	forward.readerPos = 2
	forward.readerLineStartPos = 3

	// when
	forward.endPosition()

	// then
	assert.Equal(t, forward.bufferLineStartPos, endPosition)
	assert.Equal(t, forward.readerPos, endPosition)
	assert.Equal(t, forward.readerLineStartPos, endPosition)
}

func TestGetLineSizeExcludingLineFeed(t *testing.T) {
	// given
	forward := newForward(strings.NewReader(""), 0)
	forward.buffer = append(forward.buffer, "abcdefg\n"...)

	// when
	lineSize := forward.getLineSizeExcludingLineFeed()

	// then
	assert.Equal(t, lineSize, 7)
}

func TestGetLineSizeExcludingLineFeed_OnlyLinesFeed(t *testing.T) {
	// given
	forward := newForward(strings.NewReader(""), 0)
	forward.buffer = append(forward.buffer, "\n"...)

	// when
	lineSize := forward.getLineSizeExcludingLineFeed()

	// then
	assert.Equal(t, lineSize, 0)
}

func TestGetLineSizeExcludingLineFeed_WithCarrageReturn(t *testing.T) {
	// given
	forward := newForward(strings.NewReader(""), 0)
	forward.buffer = append(forward.buffer, "abcdefg\r\n"...)

	// when
	lineSize := forward.getLineSizeExcludingLineFeed()

	// then
	assert.Equal(t, lineSize, 8)
}

func TestGetLineSizeExcludingLineFeed_WithoutLinesFeed(t *testing.T) {
	// given
	forward := newForward(strings.NewReader(""), 0)
	forward.buffer = append(forward.buffer, "abcdefg"...)

	// when
	lineSize := forward.getLineSizeExcludingLineFeed()

	// then
	assert.Equal(t, lineSize, -1)
}

func TestGetLineSizeExcludingLineFeed_Empty(t *testing.T) {
	// given
	forward := newForward(strings.NewReader(""), 0)
	forward.buffer = append(forward.buffer, ""...)

	// when
	lineSize := forward.getLineSizeExcludingLineFeed()

	// then
	assert.Equal(t, lineSize, -1)
}

func TestGetLineSizeExcludingLineFeed_EndOfFile(t *testing.T) {
	// given
	forward := newForward(strings.NewReader(""), 0)
	forward.buffer = append(forward.buffer, "abcdefg"...)
	forward.endOfFile = true

	// when
	lineSize := forward.getLineSizeExcludingLineFeed()

	// then
	assert.Equal(t, lineSize, 7)
	assert.True(t, forward.endOfScan)
}

func TestGetLineExcludingCarrageReturn(t *testing.T) {
	// given
	forward := newForward(strings.NewReader(""), 0)
	forward.buffer = append(forward.buffer, "abcdefg\r"...)

	// when
	Lines := forward.getLineExcludingCarrageReturn(8)

	// then
	assert.Equal(t, Lines, "abcdefg")
}

func TestGetLineExcludingCarrageReturn_WithoutCarrageReturn(t *testing.T) {
	// given
	forward := newForward(strings.NewReader(""), 0)
	forward.buffer = append(forward.buffer, "abcdefg"...)

	// when
	Lines := forward.getLineExcludingCarrageReturn(7)

	// then
	assert.Equal(t, Lines, "abcdefg")
}

func TestGetLineExcludingCarrageReturn_Empty(t *testing.T) {
	// given
	forward := newForward(strings.NewReader(""), 0)
	forward.buffer = append(forward.buffer, ""...)

	// when
	Lines := forward.getLineExcludingCarrageReturn(0)

	// then
	assert.Equal(t, Lines, "")
}

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
