package linescanner

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestLineScanner_New_Forward(t *testing.T) {
	scanner := New(Forward, strings.NewReader(""), 0)
	_, ok := scanner.(*forward)
	assert.True(t, ok)
}

func TestLineScanner_New_Backward(t *testing.T) {
	scanner := New(Backward, strings.NewReader(""), 0)
	_, ok := scanner.(*backward)
	assert.True(t, ok)
}

func TestLineScanner_New_Invalid(t *testing.T) {
	scanner := New(Backward+1, strings.NewReader(""), 0)
	assert.Nil(t, scanner)
}

func TestLineScanner_New_ForwardWithSize(t *testing.T) {
	scanner := NewWithSize(Forward, strings.NewReader(""), 0, 1024, 1024)
	_, ok := scanner.(*forward)
	assert.True(t, ok)
}

func TestLineScanner_New_BackwardWithSize(t *testing.T) {
	scanner := NewWithSize(Backward, strings.NewReader(""), 0, 1024, 1024)
	_, ok := scanner.(*backward)
	assert.True(t, ok)
}

func TestLineScanner_New_InvalidWithSize(t *testing.T) {
	scanner := NewWithSize(Backward+1, strings.NewReader(""), 0, 1024, 1024)
	assert.Nil(t, scanner)
}
