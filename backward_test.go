package linescanner

import (
	"strings"
	"testing"
)

func Test(t *testing.T) {
	// given
	text := "defg\nhijk\nlmno"
	reader := strings.NewReader(text)
	backward := newBackwardWithSize(reader, len(text), 4, 1024)

	// when
	backward.Line()
	backward.Line()
	backward.Line()
	backward.Line()
}
