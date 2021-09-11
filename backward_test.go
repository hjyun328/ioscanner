package linescanner

import (
	"fmt"
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

	bytes := make([]byte, 1, 100)
	r := strings.NewReader("abc")
	n, _ := r.Read(bytes)
	fmt.Println(n, bytes)

}
