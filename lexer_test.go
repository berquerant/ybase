package ybase_test

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
	"unicode"

	"github.com/berquerant/ybase"
	"github.com/stretchr/testify/assert"
)

func newReader(input string) ybase.Reader {
	return ybase.NewReader(bytes.NewBufferString(input), slog.Info)
}

func TestReader(t *testing.T) {
	input := "abcd---xxx"
	reader := newReader(input)

	assertResult := func(err error, want, got rune, buf string) func(*testing.T) {
		return func(t *testing.T) {
			assert.Equal(t, err, reader.Err(), "err")
			assert.Equal(t, want, got, "tok")
			assert.Equal(t, buf, reader.Buffer(), "buf")
		}
	}

	t.Run("1st peek", assertResult(nil, 'a', reader.Peek(), ""))
	t.Run("1st next", assertResult(nil, 'a', reader.Next(), "a"))
	t.Run("2nd next", assertResult(nil, 'b', reader.Next(), "ab"))
	t.Run("1st discard", assertResult(nil, 'c', reader.Discard(), "ab"))
	t.Run("2nd peek", assertResult(nil, 'd', reader.Peek(), "ab"))
	t.Run("3th next", assertResult(nil, 'd', reader.Next(), "abd"))
	t.Run("1st reset", func(t *testing.T) {
		reader.ResetBuffer()
		assert.Nil(t, reader.Err())
		assert.Equal(t, "", reader.Buffer())
	})
	t.Run("3rd peek", assertResult(nil, '-', reader.Peek(), ""))
	t.Run("1st next while", func(t *testing.T) {
		reader.NextWhile(func(r rune) bool { return r == '-' })
		assert.Nil(t, reader.Err())
		assert.Equal(t, "---", reader.Buffer())
	})
	t.Run("4th peek", assertResult(nil, 'x', reader.Peek(), "---"))
	t.Run("1st discard while", func(t *testing.T) {
		reader.DiscardWhile(func(r rune) bool { return r == 'x' })
		assert.Nil(t, reader.Err())
		assert.Equal(t, "---", reader.Buffer())
	})
	t.Run("final next", assertResult(nil, ybase.EOF, reader.Next(), "---"))
}

func newTokens(v ...any) []ybase.Token {
	var toks []ybase.Token
	for i := 0; i < len(v); i++ {
		t := v[i].(int)
		i++
		val := v[i].(string)
		toks = append(toks, ybase.NewToken(t, val, nil, nil))
	}
	return toks
}

func TestLexer(t *testing.T) {
	for _, tc := range []struct {
		title string
		input string
		scan  ybase.ScanFunc
		want  []ybase.Token
	}{
		{
			title: "bit operations",
			input: "1001 | 1111 & 1011",
			scan: func(r ybase.Reader) int {
				r.DiscardWhile(unicode.IsSpace)
				switch r.Peek() {
				case '0', '1':
					r.NextWhile(func(x rune) bool {
						return strings.ContainsRune("01", x)
					})
					return 10
				case '|':
					_ = r.Next()
					return 101
				case '&':
					_ = r.Next()
					return 102
				default:
					return ybase.EOF
				}
			},
			want: newTokens(
				10, "1001",
				101, "|",
				10, "1111",
				102, "&",
				10, "1011",
			),
		},
		{
			title: "identifiers delimited by space",
			input: "to be or not to be",
			scan: func(r ybase.Reader) int {
				r.DiscardWhile(unicode.IsSpace)
				r.NextWhile(unicode.IsLetter)
				if r.Buffer() == "" {
					return ybase.EOF
				}
				return 1
			},
			want: newTokens(
				1, "to",
				1, "be",
				1, "or",
				1, "not",
				1, "to",
				1, "be",
			),
		},
		{
			title: "identifiers and digits delimited by space",
			input: "2 be or not 22 be99",
			scan: func(r ybase.Reader) int {
				r.DiscardWhile(unicode.IsSpace)

				top := r.Peek()
				switch {
				case unicode.IsDigit(top):
					r.NextWhile(unicode.IsDigit)
					return 10
				case unicode.IsLetter(top):
					r.NextWhile(unicode.IsLetter)
					return 1
				default:
					return ybase.EOF
				}
			},
			want: newTokens(
				10, "2",
				1, "be",
				1, "or",
				1, "not",
				10, "22",
				1, "be",
				10, "99",
			),
		},
	} {
		tc := tc
		t.Run(tc.title, func(t *testing.T) {
			s := ybase.NewLexer(ybase.NewScanner(ybase.NewReader(bytes.NewBufferString(tc.input), slog.Info), tc.scan))
			got := []ybase.Token{}

			for s.DoLex(func(tok ybase.Token) { got = append(got, tok) }) != ybase.EOF {
			}
			if err := s.Err(); err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, len(tc.want), len(got))
			for i, w := range tc.want {
				g := got[i]
				assert.Equal(t, w.Type(), g.Type(), i)
				assert.Equal(t, w.Value(), g.Value(), i)
			}
		})
	}
}
