package ybase

import (
	"encoding/json"
	"fmt"
)

type (
	Token interface {
		Type() int
		Value() string
		Start() Pos
		End() Pos
	}

	token struct {
		t     int
		v     string
		start Pos
		end   Pos
	}
)

func NewToken(t int, v string, start, end Pos) Token {
	return &token{
		t:     t,
		v:     v,
		start: start,
		end:   end,
	}
}

func (s token) Type() int      { return s.t }
func (s token) Value() string  { return s.v }
func (s token) Start() Pos     { return s.start }
func (s token) End() Pos       { return s.end }
func (s token) String() string { return fmt.Sprintf("%d,%s", s.t, s.v) }
func (s token) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"type":  s.t,
		"value": s.v,
		"start": s.start,
		"end":   s.end,
	})
}
