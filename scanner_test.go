package linescanner

import (
	"github.com/stretchr/testify/assert"
	"io"
	"strings"
	"testing"
)

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
