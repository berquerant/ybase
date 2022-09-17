package ybase

import (
	"encoding/json"
	"fmt"
)

type (
	Token interface {
		Type() int
		Value() string
	}

	token struct {
		t int
		v string
	}
)

func NewToken(t int, v string) Token {
	return &token{
		t: t,
		v: v,
	}
}

func (s *token) Type() int      { return s.t }
func (s *token) Value() string  { return s.v }
func (s *token) String() string { return fmt.Sprintf("%d,%s", s.t, s.v) }
func (s *token) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"type":  s.t,
		"value": s.v,
	})
}

type (
	Pos interface {
		Line() int
		Column() int
		Offset() int
		Add(r rune) Pos
	}

	pos struct {
		line, col, offset int
	}
)

func NewPos(line, col, offset int) Pos {
	return &pos{
		line:   line,
		col:    col,
		offset: offset,
	}
}

func (s *pos) Line() int      { return s.line }
func (s *pos) Column() int    { return s.col }
func (s *pos) Offset() int    { return s.offset }
func (s *pos) String() string { return fmt.Sprintf("%d,%d,%d", s.line, s.col, s.offset) }
func (s *pos) Add(r rune) Pos {
	size := len([]byte(string(r)))
	if r == '\n' {
		return &pos{
			line:   s.line + 1,
			col:    0,
			offset: s.offset + size,
		}
	}
	return &pos{
		line:   s.line,
		col:    s.col + 1,
		offset: s.offset + size,
	}
}
