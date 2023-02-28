# ybase

Utilities to implement a lexer for goyacc.

## Example

``` go
package main

import (
	"bytes"
	"fmt"
	"unicode"

	"github.com/berquerant/ybase"
)

func main() {
	input := "1 + 12 - (34-56)"
	s := ybase.NewLexer(ybase.NewScanner(ybase.NewReader(bytes.NewBufferString(input), nil), func(r ybase.Reader) int {
		r.DiscardWhile(unicode.IsSpace)
		top := r.Peek()
		switch {
		case unicode.IsDigit(top):
			r.NextWhile(unicode.IsDigit)
			return 901
		default:
			switch top {
			case '+':
				_ = r.Next()
				return 911
			case '-':
				_ = r.Next()
				return 912
			case '(':
				_ = r.Next()
				return 921
			case ')':
				_ = r.Next()
				return 922
			}
		}
		return ybase.EOF
	}))
	for s.DoLex(func(tok ybase.Token) { fmt.Printf("%d %s\n", tok.Type(), tok.Value()) }) != ybase.EOF {
	}
	if err := s.Err(); err != nil {
		panic(err)
	}
	// Output:
	// 901 1
	// 911 +
	// 901 12
	// 912 -
	// 921 (
	// 901 34
	// 912 -
	// 901 56
	// 922 )
}
```
