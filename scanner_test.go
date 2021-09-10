package linescanner

import (
	"github.com/stretchr/testify/assert"
	"io"
	"strings"
	"testing"
)

func TestNew_InvalidPosition(t *testing.T) {
	assert.PanicsWithValue(t, ErrInvalidPosition, func() {
		New(strings.NewReader(""), -1)
	})
}

func TestNew_InvalidReader(t *testing.T) {
	assert.PanicsWithValue(t, ErrNilReader, func() {
		New(nil, -1)
	})
}

func TestNew_InvalidChunkSize(t *testing.T) {
	assert.PanicsWithValue(t, ErrInvalidChunkSize, func() {
		NewWithSize(strings.NewReader(""), 0, -1, 100)
	})
	assert.PanicsWithValue(t, ErrInvalidChunkSize, func() {
		NewWithSize(strings.NewReader(""), 0, 0, 100)
	})
}

func TestNew_InvalidBufferSize(t *testing.T) {
	assert.PanicsWithValue(t, ErrInvalidBufferSize, func() {
		NewWithSize(strings.NewReader(""), 0, 1, 0)
	})
	assert.PanicsWithValue(t, ErrInvalidBufferSize, func() {
		NewWithSize(strings.NewReader(""), 0, 1, -1)
	})
}

func TestNew_GreaterBufferSize(t *testing.T) {
	assert.PanicsWithValue(t, ErrGreaterBufferSize, func() {
		NewWithSize(strings.NewReader(""), 0, 2, 1)
	})
}

func TestLine_SingleLineCount(t *testing.T) {
	// given
	reader := strings.NewReader("ab\ncdefg")
	scanner := NewWithSize(reader, 0, 4, 5)

	// when
	line, err := scanner.Line(1)

	// then
	assert.Nil(t, err)
	assert.Equal(t, line, []string{"ab"})
	assert.Equal(t, scanner.Position(), 3)
	assert.Equal(t, scanner.Position(), scanner.readerLineStartPos)
	assert.Equal(t, scanner.bufferLineStartPos, 3)
	assert.Equal(t, scanner.readerPos, 4)
	assert.False(t, scanner.endOfFile)
	assert.False(t, scanner.endOfScan)
}

func TestLine_EndOfFileBufferRemains(t *testing.T) {
	// given
	reader := strings.NewReader("ab\ncdefg\nhijk")
	scanner := New(reader, 0)

	// when
	line, err := scanner.Line(2)

	// then
	assert.Nil(t, err)
	assert.Equal(t, line, []string{"ab", "cdefg"})
	assert.Equal(t, scanner.Position(), 9)
	assert.Equal(t, scanner.Position(), scanner.readerLineStartPos)
	assert.Equal(t, scanner.bufferLineStartPos, 9)
	assert.Equal(t, scanner.readerPos, 13)
	assert.True(t, scanner.endOfFile)
	assert.False(t, scanner.endOfScan)
}

func TestLine_BufferOverflow(t *testing.T) {
	// given
	reader := strings.NewReader("ab\ncdefg")
	scanner := NewWithSize(reader, 0, 4, 4)

	// when
	line, err := scanner.Line(2)

	// then
	assert.Equal(t, err, ErrBufferOverflow)
	assert.Nil(t, line)
	assert.Equal(t, scanner.Position(), 0)
	assert.Equal(t, scanner.Position(), scanner.readerLineStartPos)
	assert.Equal(t, scanner.bufferLineStartPos, 0)
	assert.Equal(t, scanner.readerPos, 0)
	assert.False(t, scanner.endOfFile)
	assert.False(t, scanner.endOfScan)
}

func TestLine_ExceedLineCount(t *testing.T) {
	// given
	reader := strings.NewReader("ab\ncdefg")
	scanner := NewWithSize(reader, 0, 4, 5)

	// when
	line, err := scanner.Line(100)

	// then
	assert.NotNil(t, err)
	assert.Equal(t, line, []string{"ab", "cdefg"})
	assert.Equal(t, scanner.Position(), endPosition)
	assert.Equal(t, scanner.Position(), scanner.readerLineStartPos)
	assert.Equal(t, scanner.bufferLineStartPos, endPosition)
	assert.Equal(t, scanner.readerPos, endPosition)
	assert.True(t, scanner.endOfFile)
	assert.True(t, scanner.endOfScan)
}

func TestLine_Empty(t *testing.T) {
	// given
	reader := strings.NewReader("")
	scanner := NewWithSize(reader, 0, 4, 4)

	// when
	line, err := scanner.Line(100)

	// then
	assert.Equal(t, err, io.EOF)
	assert.Equal(t, line, []string{""})
	assert.Equal(t, scanner.Position(), endPosition)
	assert.Equal(t, scanner.Position(), scanner.readerLineStartPos)
	assert.Equal(t, scanner.bufferLineStartPos, endPosition)
	assert.Equal(t, scanner.readerPos, endPosition)
	assert.True(t, scanner.endOfFile)
	assert.True(t, scanner.endOfScan)
}

func TestLine_TextOnly(t *testing.T) {
	// given
	reader := strings.NewReader("abcdefg")
	scanner := New(reader, 0)

	// when
	line, err := scanner.Line(100)

	// then
	assert.Equal(t, err, io.EOF)
	assert.Equal(t, line, []string{"abcdefg"})
	assert.Equal(t, scanner.Position(), endPosition)
	assert.Equal(t, scanner.Position(), scanner.readerLineStartPos)
	assert.Equal(t, scanner.bufferLineStartPos, endPosition)
	assert.Equal(t, scanner.readerPos, endPosition)
	assert.True(t, scanner.endOfFile)
	assert.True(t, scanner.endOfScan)
}

func TestLine_LineFeedOnly(t *testing.T) {
	// given
	reader := strings.NewReader("\n\n\n")
	scanner := New(reader, 0)

	// when
	line, err := scanner.Line(100)

	// then
	assert.Equal(t, err, io.EOF)
	assert.Equal(t, line, []string{"", "", "", ""})
	assert.Equal(t, scanner.Position(), endPosition)
	assert.Equal(t, scanner.Position(), scanner.readerLineStartPos)
	assert.Equal(t, scanner.bufferLineStartPos, endPosition)
	assert.Equal(t, scanner.readerPos, endPosition)
	assert.True(t, scanner.endOfFile)
	assert.True(t, scanner.endOfScan)
}

func TestLine_WithCarrageReturn(t *testing.T) {
	// given
	reader := strings.NewReader("ab\r\ncdefg")
	scanner := NewWithSize(reader, 0, 4, 5)

	// when
	line, err := scanner.Line(100)

	// then
	assert.NotNil(t, err)
	assert.Equal(t, len(line), 2)
	assert.Equal(t, line[0], "ab")
	assert.Equal(t, line[1], "cdefg")
	assert.Equal(t, scanner.Position(), endPosition)
	assert.Equal(t, scanner.Position(), scanner.readerLineStartPos)
	assert.Equal(t, scanner.bufferLineStartPos, endPosition)
	assert.Equal(t, scanner.readerPos, endPosition)
	assert.True(t, scanner.endOfFile)
	assert.True(t, scanner.endOfScan)
}

func TestLine_WithPosition(t *testing.T) {
	// given
	reader := strings.NewReader("ab\ncdefg")
	scanner := New(reader, 3)

	// when
	line, err := scanner.Line(100)

	// then
	assert.NotNil(t, err)
	assert.Equal(t, line, []string{"cdefg"})
	assert.Equal(t, scanner.Position(), endPosition)
	assert.Equal(t, scanner.Position(), scanner.readerLineStartPos)
	assert.Equal(t, scanner.bufferLineStartPos, endPosition)
	assert.Equal(t, scanner.readerPos, endPosition)
	assert.True(t, scanner.endOfFile)
	assert.True(t, scanner.endOfScan)
}

func TestLine_ExceedPosition(t *testing.T) {
	// given
	reader := strings.NewReader("ab\ncdefg")
	scanner := New(reader, 100)

	// when
	line, err := scanner.Line(100)

	// then
	assert.Equal(t, err, io.EOF)
	assert.Equal(t, line, []string{""})
	assert.Equal(t, scanner.Position(), endPosition)
	assert.Equal(t, scanner.Position(), scanner.readerLineStartPos)
	assert.Equal(t, scanner.bufferLineStartPos, endPosition)
	assert.Equal(t, scanner.readerPos, endPosition)
	assert.True(t, scanner.endOfFile)
	assert.True(t, scanner.endOfScan)
}

func TestGetLineSizeExcludingLF(t *testing.T) {
	// given
	scanner := New(strings.NewReader(""), 0)
	scanner.buffer = append(scanner.buffer, "abcdefg\n"...)

	// when
	lineSize := scanner.getLineSizeExcludingLF()

	// then
	assert.Equal(t, lineSize, 7)
}

func TestGetLineSizeExcludingLF_OnlyLineFeed(t *testing.T) {
	// given
	scanner := New(strings.NewReader(""), 0)
	scanner.buffer = append(scanner.buffer, "\n"...)

	// when
	lineSize := scanner.getLineSizeExcludingLF()

	// then
	assert.Equal(t, lineSize, 0)
}

func TestGetLineSizeExcludingLF_WithCarrageReturn(t *testing.T) {
	// given
	scanner := New(strings.NewReader(""), 0)
	scanner.buffer = append(scanner.buffer, "abcdefg\r\n"...)

	// when
	lineSize := scanner.getLineSizeExcludingLF()

	// then
	assert.Equal(t, lineSize, 8)
}

func TestGetLineSizeExcludingLF_WithoutLineFeed(t *testing.T) {
	// given
	scanner := New(strings.NewReader(""), 0)
	scanner.buffer = append(scanner.buffer, "abcdefg"...)

	// when
	lineSize := scanner.getLineSizeExcludingLF()

	// then
	assert.Equal(t, lineSize, -1)
}

func TestGetLineSizeExcludingLF_Empty(t *testing.T) {
	// given
	scanner := New(strings.NewReader(""), 0)
	scanner.buffer = append(scanner.buffer, ""...)

	// when
	lineSize := scanner.getLineSizeExcludingLF()

	// then
	assert.Equal(t, lineSize, -1)
}

func TestGetLineSizeExcludingLF_EndOfFile(t *testing.T) {
	// given
	scanner := New(strings.NewReader(""), 0)
	scanner.buffer = append(scanner.buffer, "abcdefg"...)
	scanner.endOfFile = true

	// when
	lineSize := scanner.getLineSizeExcludingLF()

	// then
	assert.Equal(t, lineSize, 7)
	assert.True(t, scanner.endOfScan)
}

func TestGetLineExcludingCR(t *testing.T) {
	// given
	scanner := New(strings.NewReader(""), 0)
	scanner.buffer = append(scanner.buffer, "abcdefg\r"...)

	// when
	line := scanner.getLineExcludingCR(8)

	// then
	assert.Equal(t, line, "abcdefg")
}

func TestGetLineExcludingCR_WithoutCarrageReturn(t *testing.T) {
	// given
	scanner := New(strings.NewReader(""), 0)
	scanner.buffer = append(scanner.buffer, "abcdefg"...)

	// when
	line := scanner.getLineExcludingCR(7)

	// then
	assert.Equal(t, line, "abcdefg")
}

func TestGetLineExcludingCR_Empty(t *testing.T) {
	// given
	scanner := New(strings.NewReader(""), 0)
	scanner.buffer = append(scanner.buffer, ""...)

	// when
	line := scanner.getLineExcludingCR(0)

	// then
	assert.Equal(t, line, "")
}

func TestArrangeBuffer(t *testing.T) {
	// given
	scanner := NewWithSize(strings.NewReader(""), 0, 4, 7)
	scanner.buffer = append(scanner.buffer, "abcdefg"...)
	scanner.bufferLineStartPos = 3

	// when
	err := scanner.arrangeBuffer(3)

	// then
	assert.Nil(t, err)
	assert.Equal(t, scanner.buffer, []byte("defg"))
	assert.Equal(t, scanner.bufferLineStartPos, 0)
}

func TestArrangeBuffer_BufferOverflow(t *testing.T) {
	// given
	scanner := NewWithSize(strings.NewReader(""), 0, 4, 7)
	scanner.buffer = append(scanner.buffer, "abcdefg"...)
	scanner.bufferLineStartPos = 3

	// when
	err := scanner.arrangeBuffer(4)

	// then
	assert.Equal(t, err, ErrBufferOverflow)
	assert.Equal(t, scanner.buffer, []byte("abcdefg"))
	assert.Equal(t, scanner.bufferLineStartPos, 3)
}

func TestArrangeBuffer_NotArranged(t *testing.T) {
	// given
	scanner := NewWithSize(strings.NewReader(""), 0, 4, 7)
	scanner.buffer = append(scanner.buffer, "abcde"...)
	scanner.bufferLineStartPos = 3

	// when
	err := scanner.arrangeBuffer(2)

	// then
	assert.Nil(t, err)
	assert.Equal(t, scanner.buffer, []byte("abcde"))
	assert.Equal(t, scanner.bufferLineStartPos, 3)
}

func TestRead(t *testing.T) {
	// given
	scanner := NewWithSize(strings.NewReader("abcdefghijklmnop"), 0, 4, 8)

	// when
	err := scanner.read()

	// then
	assert.Nil(t, err)
	assert.Equal(t, scanner.chunk, []byte("abcd"))
	assert.Equal(t, scanner.buffer, []byte("abcd"))
	assert.Equal(t, scanner.bufferLineStartPos, 0)
	assert.Equal(t, scanner.readerPos, 4)
	assert.False(t, scanner.endOfFile)

	// when
	err = scanner.read()

	// then
	assert.Nil(t, err)
	assert.Equal(t, scanner.chunk, []byte("efgh"))
	assert.Equal(t, scanner.buffer, []byte("abcdefgh"))
	assert.Equal(t, scanner.bufferLineStartPos, 0)
	assert.Equal(t, scanner.readerPos, 8)
	assert.False(t, scanner.endOfFile)
}

func TestRead_EndOfFile(t *testing.T) {
	// given
	scanner := NewWithSize(strings.NewReader("abc"), 0, 4, 7)

	// when
	err := scanner.read()

	// then
	assert.Nil(t, err)
	assert.Equal(t, scanner.chunk, append([]byte("abc"), 0x00))
	assert.Equal(t, scanner.buffer, []byte("abc"))
	assert.Equal(t, scanner.bufferLineStartPos, 0)
	assert.Equal(t, scanner.readerPos, 3)
	assert.True(t, scanner.endOfFile)
}

func TestRead_Arranged(t *testing.T) {
	// given
	scanner := NewWithSize(strings.NewReader("abcdefghijk"), 7, 3, 7)
	scanner.buffer = append(scanner.buffer, "abcdefg"...)
	scanner.bufferLineStartPos = 3

	// when
	err := scanner.read()

	// then
	assert.Nil(t, err)
	assert.Equal(t, scanner.chunk, []byte("hij"))
	assert.Equal(t, scanner.buffer, []byte("defghij"))
	assert.Equal(t, scanner.bufferLineStartPos, 0)
	assert.Equal(t, scanner.readerPos, 10)
	assert.False(t, scanner.endOfFile)
}

func TestRead_BufferOverflow(t *testing.T) {
	// given
	scanner := NewWithSize(strings.NewReader("abcdefghijk"), 7, 4, 7)
	scanner.buffer = append(scanner.buffer, "abcdefg"...)
	scanner.bufferLineStartPos = 3

	// when
	err := scanner.read()

	// then
	assert.Equal(t, err, ErrBufferOverflow)
	assert.Equal(t, scanner.chunk, []byte("hijk"))
	assert.Equal(t, scanner.buffer, []byte("abcdefg"))
	assert.Equal(t, scanner.bufferLineStartPos, 3)
	assert.Equal(t, scanner.readerPos, 7)
	assert.False(t, scanner.endOfFile)
}
