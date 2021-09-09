package linescanner

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func Test_SingleLineCount(t *testing.T) {
	// given
	reader := strings.NewReader("ab\ncdefg")
	scanner := NewWithSize(reader, 0, 4, 5)

	// when
	line, err := scanner.Line(1)

	// then
	assert.Nil(t, err)
	assert.Equal(t, len(line), 1)
	assert.Equal(t, line[0], "ab")
	assert.Equal(t, scanner.Position(), 3)
	assert.Equal(t, scanner.Position(), scanner.readerLineStartPos)
	assert.Equal(t, scanner.bufferLineStartPos, 3)
	assert.Equal(t, scanner.readerPos, 4)
	assert.False(t, scanner.eof)
	assert.False(t, scanner.eob)
}

func Test_EndOfFileBufferRemains(t *testing.T) {
	// given
	reader := strings.NewReader("ab\ncdefg\nhijk")
	scanner := NewWithSize(reader, 0, 1024, 1024)

	// when
	line, err := scanner.Line(2)

	// then
	assert.Nil(t, err)
	assert.Equal(t, len(line), 2)
	assert.Equal(t, line[0], "ab")
	assert.Equal(t, line[1], "cdefg")
	assert.Equal(t, scanner.Position(), 9)
	assert.Equal(t, scanner.Position(), scanner.readerLineStartPos)
	assert.Equal(t, scanner.bufferLineStartPos, 9)
	assert.Equal(t, scanner.readerPos, 13)
	assert.True(t, scanner.eof)
	assert.False(t, scanner.eob)
}

func Test_BufferOverflow(t *testing.T) {
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
	assert.False(t, scanner.eof)
	assert.False(t, scanner.eob)
}

func Test_ExceedLineCount(t *testing.T) {
	// given
	reader := strings.NewReader("ab\ncdefg")
	scanner := NewWithSize(reader, 0, 4, 5)

	// when
	line, err := scanner.Line(100)

	// then
	assert.NotNil(t, err)
	assert.Equal(t, len(line), 2)
	assert.Equal(t, line[0], "ab")
	assert.Equal(t, line[1], "cdefg")
	assert.Equal(t, scanner.Position(), 8)
	assert.Equal(t, scanner.Position(), scanner.readerLineStartPos)
	assert.Equal(t, scanner.bufferLineStartPos, 5)
	assert.Equal(t, scanner.readerPos, 8)
	assert.True(t, scanner.eof)
	assert.True(t, scanner.eob)
}

func Test_CarrageReturn(t *testing.T) {
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
	assert.Equal(t, scanner.Position(), 9)
	assert.Equal(t, scanner.Position(), scanner.readerLineStartPos)
	assert.Equal(t, scanner.bufferLineStartPos, 5)
	assert.Equal(t, scanner.readerPos, 9)
	assert.True(t, scanner.eof)
	assert.True(t, scanner.eob)
}

func Test_Empty(t *testing.T) {
	// given
	reader := strings.NewReader("")
	scanner := NewWithSize(reader, 0, 4, 4)

	// when
	line, err := scanner.Line(100)

	// then
	assert.NotNil(t, err)
	assert.Equal(t, len(line), 2)
	assert.Equal(t, line[0], "ab")
	assert.Equal(t, line[1], "cdefg")
	assert.Equal(t, scanner.Position(), 9)
	assert.Equal(t, scanner.Position(), scanner.readerLineStartPos)
	assert.Equal(t, scanner.bufferLineStartPos, 5)
	assert.Equal(t, scanner.readerPos, 9)
	assert.True(t, scanner.eof)
	assert.True(t, scanner.eob)
}
