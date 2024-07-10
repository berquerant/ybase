package ybase

import "fmt"

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
