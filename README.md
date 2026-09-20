# gitignorematch

A Go library for matching paths against `.gitignore` patterns. Requires Go 1.18 or later.

## Install

```sh
go get github.com/skw398/gitignorematch
```

## Usage

```go
package main

import (
	"fmt"

	"github.com/skw398/gitignorematch"
)

func main() {
	matcher, err := gitignorematch.Parse("*.log\n!important.log\n")
	if err != nil {
		panic(err)
	}

	fmt.Println(matcher.Match("app.log", false))       // Ignore
	fmt.Println(matcher.Match("important.log", false)) // Include
	fmt.Println(matcher.Match("main.go", false))       // NoMatch
}
```

Use slash-separated paths relative to the directory containing the rules.
Pass `true` as the second argument when matching a directory. A file inside an
ignored directory stays ignored, even if a negated rule matches it.

`ParseFile` reads a `.gitignore` file; `ParseReader` accepts an `io.Reader`.
`Explain` returns the decision, matching pattern, and one-based line number.
For ASCII case-insensitive matching, use `ParseWithOptions` with
`Options{CaseInsensitive: true}`.

The caller handles directory traversal and loading rules from nested `.gitignore` files.

## License

[MIT](LICENSE)
