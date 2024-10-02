package ybase

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"
)

const EOF = -1

var ErrYbase = errors.New("Ybase")

// DebugFunc outputs debug logs.
// Assuming a function like slog.Debug.
type DebugFunc func(msg string, v ...any)

func NilDebugFunc(msg string, v ...any) {}

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
	Debugf(msg string, v ...any)
	// Errorf outputs logs and set an error.
	Errorf(err error, msg string, v ...any)
	// DiscardWhile calls Discard() while pred(Peek()).
	DiscardWhile(pred func(rune) bool)
	// NextWhile calls Next() while pred(Peek()).
	NextWhile(pred func(rune) bool)
	// Pos returns the current position.
	Pos() Pos
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

func (r reader) Pos() Pos       { return r.pos }
func (r *reader) ResetBuffer()  { r.buf.Reset() }
func (r reader) Buffer() string { return r.buf.String() }
func (r reader) Err() error     { return r.err }
func (r reader) logAttrs() []any {
	return []any{
		slog.Int("line", r.pos.Line()),
		slog.Int("column", r.pos.Column()),
		slog.Int("offset", r.pos.Offset()),
		slog.String("buf", r.buf.String()),
	}
}
func (r reader) Debugf(msg string, v ...any) {
	attrs := r.logAttrs()
	attrs = append(attrs, v...)
	r.debugFunc("ybase: "+msg, attrs...)
}
func (r reader) Errorf(err error, msg string, v ...any) {
	r.err = errors.Join(ErrYbase, fmt.Errorf("%w: %s", err, msg))
	attrs := r.logAttrs()
	attrs = append(attrs, v...)
	attrs = append(attrs, slog.Any("err", r.err))
	r.debugFunc("ybase: "+msg, attrs...)
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
	r.Debugf("Discard", slog.Any("rune", g), slog.Any("err", err))
	if err != nil {
		if !errors.Is(err, io.EOF) {
			r.Errorf(err, "Discard from reader")
		}
		return EOF
	}
	r.pos = r.pos.Add(g)
	return g
}

func (r *reader) Peek() rune {
	g, _, err := r.rdr.ReadRune()
	r.Debugf("Peek", slog.Any("rune", g), slog.Any("err", err))
	if err != nil {
		if !errors.Is(err, io.EOF) {
			r.Errorf(err, "Peek from reader")
		}
		return EOF
	}
	if err := r.rdr.UnreadRune(); err != nil {
		r.Errorf(err, "Peek failed to unread")
		return EOF
	}
	return g
}

func (r *reader) Next() rune {
	g, _, err := r.rdr.ReadRune()
	r.Debugf("Next", slog.Any("rune", g), slog.Any("err", err))
	if err != nil {
		if !errors.Is(err, io.EOF) {
			r.Errorf(err, "Next from reader")
		}
		return EOF
	}

	r.pos = r.pos.Add(g)
	if _, err := r.buf.WriteRune(g); err != nil {
		r.Errorf(err, "Next failed to write buffer")
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

func (s *scanner) Scan() int { return s.scanFunc(s.Reader) }
func (s scanner) Error(msg string) {
	s.Errorf(fmt.Errorf("%w: %s", ErrYbase, msg), msg)
}

// Lexer is an utility to implement yyLexer.
//
// Recommendation:
// - Set level to yyDebug (YYDEBUG in yacc).
// - Set yyErrorVerbose to true (YYERROR_VERBOSE in yacc)
//
// Implements yyLexer by Error(string) and Lex(*yySymType) int, e.g.
//
//	type ActualLexer struct {
//	  Lexer
//	}
//
//	func (a *ActualLexer) Lex(lval *yySymType) int {
//	  return a.DoLex(func(tok Token) {
//	    lval.token = tok  // declares in %union
//	  })
//	}
type Lexer interface {
	Scanner
	// DoLex runs the lexical analysis.
	// Returns EOF if EOF or an error occurs.
	DoLex(callback func(Token)) int
}

type lexer struct {
	Scanner
	pos Pos
}

func NewLexer(scanner Scanner) Lexer {
	return &lexer{
		Scanner: scanner,
		pos:     scanner.Pos(),
	}
}

func (l *lexer) DoLex(callback func(Token)) int {
	if l.Err() != nil {
		return EOF
	}
	start := l.pos
	t := l.Scan()
	if t == EOF || l.Err() != nil {
		return EOF
	}
	end := l.Pos()
	l.pos = end
	v := l.Buffer()
	tok := NewToken(t, v, start, end)
	callback(tok)
	l.Debugf("Lex",
		slog.Int("type", tok.Type()),
		slog.String("value", tok.Value()),
		slog.Int("start.line", tok.Start().Line()),
		slog.Int("start.column", tok.Start().Column()),
		slog.Int("start.offset", tok.Start().Offset()),
		slog.Int("end.line", tok.End().Line()),
		slog.Int("end.column", tok.End().Column()),
		slog.Int("end.offset", tok.End().Offset()),
	)
	l.ResetBuffer()
	return tok.Type()
}
