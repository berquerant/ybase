package ybase

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
)

const EOF = -1

// Debugf outputs debug logs.
type DebugFunc func(format string, v ...any)

func NilDebugFunc(format string, v ...any) {}

// Reader represents a reader object for lex.
type Reader interface {
	// ResetBuffer clears the buffer.
	ResetBuffer()
	// Buffer returns the read runes.
	Buffer() string
	// Next gets the next rune and advances the pos.
	Next() rune
	// Peek gets the next rune but keeps the pos.
	Peek() rune
	// Discard ignores the next rune.
	Discard() rune
	// Err returns an error during the reading.
	Err() error
	// Debugf outputs debug logs.
	Debugf(format string, v ...any)
	// Errorf outputs logs and set an error.
	Errorf(format string, v ...any)
	// DiscardWhile calls Discard() while pred(Peek()).
	DiscardWhile(pred func(rune) bool)
	// NextWhile calls Next() while pred(Peek()).
	NextWhile(pred func(rune) bool)
}

type reader struct {
	pos       Pos
	rdr       *bufio.Reader
	buf       bytes.Buffer
	err       error
	debugFunc DebugFunc
}

func NewReaderWithInitPos(rdr io.Reader, debugFunc DebugFunc, initPos Pos) Reader {
	if debugFunc == nil {
		debugFunc = NilDebugFunc
	}
	return &reader{
		pos:       initPos,
		rdr:       bufio.NewReader(rdr),
		debugFunc: debugFunc,
	}
}

func NewReader(rdr io.Reader, debugFunc DebugFunc) Reader {
	return NewReaderWithInitPos(rdr, debugFunc, NewPos(1, 0, 0))
}

func (r *reader) ResetBuffer()      { r.buf.Reset() }
func (r *reader) Buffer() string    { return r.buf.String() }
func (r *reader) Err() error        { return r.err }
func (r *reader) logHeader() string { return fmt.Sprintf("[ybase][%s][%s]", r.pos, r.buf.String()) }
func (r *reader) Debugf(format string, v ...any) {
	r.debugFunc("%s %s", r.logHeader(), fmt.Sprintf(format, v...))
}
func (r *reader) Errorf(format string, v ...any) {
	r.err = fmt.Errorf("%s %w", r.logHeader(), fmt.Errorf(format, v...))
	r.Debugf("[ybase] error %v", r.err)
}

func (r *reader) DiscardWhile(pred func(rune) bool) {
	for x := r.Peek(); pred(x); x = r.Peek() {
		_ = r.Discard()
	}
}

func (r *reader) NextWhile(pred func(rune) bool) {
	for x := r.Peek(); pred(x); x = r.Peek() {
		r.next()
	}
}

func (r *reader) Discard() rune {
	g, _, err := r.rdr.ReadRune()
	r.Debugf("[Discard] %q %v", g, err)
	if err != nil {
		if !errors.Is(err, io.EOF) {
			r.Errorf("[Discard] from reader %w", err)
		}
		return EOF
	}
	r.pos = r.pos.Add(g)
	return g
}

func (r *reader) Peek() rune {
	g, _, err := r.rdr.ReadRune()
	r.Debugf("[Peek] %q %v", g, err)
	if err != nil {
		if !errors.Is(err, io.EOF) {
			r.Errorf("[Peek] from reader %w", err)
		}
		return EOF
	}
	if err := r.rdr.UnreadRune(); err != nil {
		r.Errorf("[Peek] failed to unread %w", err)
		return EOF
	}
	return g
}

func (r *reader) Next() rune {
	g, _, err := r.rdr.ReadRune()
	r.Debugf("[Next] %q %v", g, err)
	if err != nil {
		if !errors.Is(err, io.EOF) {
			r.Errorf("[Next] from reader %w", err)
		}
		return EOF
	}

	r.pos = r.pos.Add(g)
	if _, err := r.buf.WriteRune(g); err != nil {
		r.Errorf("[Next] failed to write buffer %w", err)
		return EOF
	}
	return g
}

func (r *reader) next() { _ = r.Next() }

// ScanFunc scans source and calculate token.
type ScanFunc func(Reader) int

type Scanner interface {
	Reader
	Scan() int
	// Error consumes an error from yyLexer.
	Error(msg string)
}

type scanner struct {
	Reader
	scanFunc ScanFunc
}

func NewScanner(rdr Reader, scanFunc ScanFunc) Scanner {
	return &scanner{
		Reader:   rdr,
		scanFunc: scanFunc,
	}
}

func (s *scanner) Scan() int        { return s.scanFunc(s.Reader) }
func (s *scanner) Error(msg string) { s.Errorf(msg) }

// Lexer is an utility to implement yyLexer.
//
// Recommendation:
// - Set level to yyDebug (YYDEBUG in yacc).
// - Set yyErrorVerbose to true (YYERROR_VERBOSE in yacc)
//
// Implements yyLexer by Error(string) and Lex(*yySymType) int, e.g.
//
//   type ActualLexer struct {
//     Lexer
//   }
//
//   func (a *ActualLexer) Lex(lval *yySymType) int {
//     return a.DoLex(func(tok Token) {
//       lval.token = tok  // declares in %union
//     })
//   }
type Lexer interface {
	Scanner
	// DoLex runs the lexical analysis.
	// Returns EOF if EOF or an error occurs.
	DoLex(callback func(Token)) int
}

type lexer struct {
	Scanner
}

func NewLexer(scanner Scanner) Lexer {
	return &lexer{
		Scanner: scanner,
	}
}

func (l *lexer) DoLex(callback func(Token)) int {
	if l.Err() != nil {
		return EOF
	}
	t := l.Scan()
	if t == EOF || l.Err() != nil {
		return EOF
	}
	v := l.Buffer()
	tok := NewToken(t, v)
	callback(tok)
	l.Debugf("[Lex] %s", tok)
	l.ResetBuffer()
	return tok.Type()
}
