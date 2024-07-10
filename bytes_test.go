package ybase_test

import (
	"testing"

	"github.com/berquerant/ybase"
	"github.com/stretchr/testify/assert"
)

func TestBytes(t *testing.T) {
	const document = `apiVersion: v1
kind: Text
metadata:
  name: sometext
spec:
  text1: テキスト
  text2: text`
	b := ybase.Bytes(document)

	t.Run("LineColumn", func(t *testing.T) {
		for _, tc := range []struct {
			title  string
			offset int
		}{
			{
				title:  "negative",
				offset: -1,
			},
			{
				title:  "out of range",
				offset: 300,
			},
		} {
			t.Run(tc.title, func(t *testing.T) {
				_, _, ok := b.LineColumn(tc.offset)
				assert.False(t, ok)
			})
		}
	})

	t.Run("Offset", func(t *testing.T) {
		for _, tc := range []struct {
			title  string
			line   int
			column int
		}{
			{
				title:  "zero line",
				line:   0,
				column: 1,
			},
			{
				title:  "zero column",
				line:   1,
				column: 0,
			},
			{
				title:  "line out of range",
				line:   10,
				column: 1,
			},
			{
				title:  "column out of rage",
				line:   1,
				column: 100,
			},
		} {
			t.Run(tc.title, func(t *testing.T) {
				_, ok := b.Offset(tc.line, tc.column)
				assert.False(t, ok)
			})
		}
	})

	assertRound := func(t *testing.T, line, column int) {
		offset, ok := b.Offset(line, column)
		assert.True(t, ok, "offset(%d, %d)", line, column)
		gotLine, gotColumn, ok := b.LineColumn(offset)
		assert.True(t, ok, "linecolumn(%d)", offset)
		t.Logf("round: %d(%d, %d) = %q", offset, line, column, document[offset])
		assert.Equal(t, line, gotLine, "line")
		assert.Equal(t, column, gotColumn, "column")
	}

	t.Run("RoundTrip", func(t *testing.T) {
		for _, tc := range []struct {
			title  string
			line   int
			column int
		}{
			{
				title:  "first char",
				line:   1,
				column: 1,
			},
			{
				title:  "first line",
				line:   1,
				column: 11,
			},
			{
				title:  "second line",
				line:   2,
				column: 1,
			},
			{
				title:  "multibyte char",
				line:   6,
				column: 11,
			},
			{
				title:  "final line",
				line:   7,
				column: 3,
			},
		} {
			t.Run(tc.title, func(t *testing.T) {
				assertRound(t, tc.line, tc.column)
			})
		}
	})
}
