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
