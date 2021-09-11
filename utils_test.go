package linescanner

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMinInt(t *testing.T) {
	// case 1
	x, y := 2, 1
	min := minInt(x, y)
	assert.Equal(t, min, 1)

	// case 2
	min = minInt(y, x)
	assert.Equal(t, min, 1)

	// case 3
	x, y = 1, 1
	assert.Equal(t, min, 1)
}

func TestRemoveCarrageReturn(t *testing.T) {
	// given
	line := []byte("abcd\r")

	// when
	lineStr := removeCarrageReturn(line)

	// then
	assert.Equal(t, lineStr, "abcd")
}

func TestRemoveCarrageReturn_NoCarrageReturn(t *testing.T) {
	// given
	line := []byte("abcd")

	// when
	lineStr := removeCarrageReturn(line)

	// then
	assert.Equal(t, lineStr, "abcd")
}

func TestRemoveCarrageReturn_EmptyLine(t *testing.T) {
	// given
	line := []byte("")

	// when
	lineStr := removeCarrageReturn(line)

	// then
	assert.Empty(t, lineStr)
}
