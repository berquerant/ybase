package ybase_test

import (
	"fmt"
	"testing"

	"github.com/berquerant/ybase"
	"github.com/stretchr/testify/assert"
)

func TestPos(t *testing.T) {
	p := ybase.NewPos(1, 0, 0)

	const str = "aあ\nb"
	for i, tc := range []struct {
		line   int
		col    int
		offset int
	}{
		{
			line:   1,
			col:    1,
			offset: 1,
		},
		{
			line:   1,
			col:    2,
			offset: 4,
		},
		{
			line:   2,
			col:    0,
			offset: 5,
		},
		{
			line:   2,
			col:    1,
			offset: 6,
		},
	} {
		r := []rune(str)[i]
		tc := tc
		t.Run(fmt.Sprintf("Add %s", string(r)), func(t *testing.T) {
			p = p.Add(r)
			assert.Equal(t, tc.line, p.Line(), "line")
			assert.Equal(t, tc.col, p.Column(), "col")
			assert.Equal(t, tc.offset, p.Offset(), "offset")

			rsize := len([]byte(string(r)))
			start := p.Offset() - rsize
			end := p.Offset()
			assert.Equal(t, string(r), string(str[start:end]), "rune")
		})
	}
}
