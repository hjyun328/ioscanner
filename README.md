### Get module

```
go get github.com/hjyun328/linescanner
```

### Backward scan
 
```go
package main

import (
	"fmt"
	"github.com/hjyun328/linescanner"
	"io"
	"strings"
)

func main() {
	data := "abcd\nefgh\nijkl"
	scanner := linescanner.NewBackward(strings.NewReader(data), len(data))

	line, err := scanner.Line()
	if err != nil && err != io.EOF {
		panic(err)
	}
	fmt.Println(line, err == io.EOF) // ijkl false

	line, err = scanner.Line()
	if err != nil && err != io.EOF {
		panic(err)
	}
	fmt.Println(line, err == io.EOF) // efgh false

	scanner = linescanner.NewBackward(strings.NewReader(data), scanner.Position())

	line, err = scanner.Line()
	if err != nil && err != io.EOF {
		panic(err)
	}
	fmt.Println(line, err == io.EOF) // abcd true

	scanner = linescanner.NewBackward(strings.NewReader(data), 9)

	line, err = scanner.Line()
	if err != nil && err != io.EOF {
		panic(err)
	}
	fmt.Println(line, err == io.EOF) // efgh false
}
```

### Forward scan

```go
package main

import (
	"fmt"
	"github.com/hjyun328/linescanner"
	"io"
	"strings"
)

func main() {
	data := "abcd\nefgh\nijkl"
	scanner := linescanner.NewForward(strings.NewReader(data), 0)

	line, err := scanner.Line()
	if err != nil && err != io.EOF {
		panic(err)
	}
	fmt.Println(line, err == io.EOF) // abcd false

	scanner = linescanner.NewForward(strings.NewReader(data), scanner.Position())

	line, err = scanner.Line()
	if err != nil && err != io.EOF {
		panic(err)
	}
	fmt.Println(line, err == io.EOF) // efgh false

	line, err = scanner.Line()
	if err != nil && err != io.EOF {
		panic(err)
	}
	fmt.Println(line, err == io.EOF) // ijkl true

	scanner = linescanner.NewForward(strings.NewReader(data), 5)

	line, err = scanner.Line()
	if err != nil && err != io.EOF {
		panic(err)
	}
	fmt.Println(line, err == io.EOF) // efgh false
}
```
