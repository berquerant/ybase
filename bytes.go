package ybase

import "bytes"

type Bytes []byte

// LineColumn calculates line number and column from offset.
func (b Bytes) LineColumn(offset int) (int, int, bool) {
	if offset < 0 || offset >= len(b) {
		return 0, 0, false
	}

	target := b[:offset]
	xs := bytes.Split(target, []byte("\n"))
	line := len(xs)
	last := string(xs[len(xs)-1])
	column := len(last) + 1
	return line, column, true
}

// Offset calculates offset from line number and column.
func (b Bytes) Offset(line, column int) (int, bool) {
	if line < 1 || column < 1 {
		return 0, false
	}
	xs := bytes.Split(b, []byte("\n"))
	if line-1 >= len(xs) {
		return 0, false
	}
	lineRow := string(xs[line-1])
	if column-1 >= len(lineRow) {
		return 0, false
	}

	var offset int
	for _, x := range xs[:line-1] {
		offset += len(x)
	}
	offset += len(xs[:line-1]) - 1
	offset += len([]byte(lineRow[:column]))
	return offset, true
}
