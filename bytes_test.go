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

	t.Run("Context", func(t *testing.T) {
		for _, tc := range []struct {
			title string
			linum int
			count int
			want  *ybase.Context
			notOK bool
		}{
			{
				title: "line 7 count 1",
				linum: 7,
				count: 1,
				want: &ybase.Context{
					Target: &ybase.ContextLine{
						Linum: 7,
						Line:  []byte(`  text2: text`),
					},
					Lines: []*ybase.ContextLine{
						{
							Linum: 6,
							Line:  []byte(`  text1: テキスト`),
						},
						{
							Linum: 7,
							Line:  []byte(`  text2: text`),
						},
					},
				},
			},
			{
				title: "line 3 count 2",
				linum: 3,
				count: 2,
				want: &ybase.Context{
					Target: &ybase.ContextLine{
						Linum: 3,
						Line:  []byte(`metadata:`),
					},
					Lines: []*ybase.ContextLine{
						{
							Linum: 1,
							Line:  []byte(`apiVersion: v1`),
						},
						{
							Linum: 2,
							Line:  []byte(`kind: Text`),
						},
						{
							Linum: 3,
							Line:  []byte(`metadata:`),
						},
						{
							Linum: 4,
							Line:  []byte(`  name: sometext`),
						},
						{
							Linum: 5,
							Line:  []byte(`spec:`),
						},
					},
				},
			},
			{
				title: "line 1 count 1",
				linum: 1,
				count: 1,
				want: &ybase.Context{
					Target: &ybase.ContextLine{
						Linum: 1,
						Line:  []byte(`apiVersion: v1`),
					},
					Lines: []*ybase.ContextLine{
						{
							Linum: 1,
							Line:  []byte(`apiVersion: v1`),
						},
						{
							Linum: 2,
							Line:  []byte(`kind: Text`),
						},
					},
				},
			},
			{
				title: "line 1 only",
				linum: 1,
				count: 0,
				want: &ybase.Context{
					Target: &ybase.ContextLine{
						Linum: 1,
						Line:  []byte(`apiVersion: v1`),
					},
					Lines: []*ybase.ContextLine{
						{
							Linum: 1,
							Line:  []byte(`apiVersion: v1`),
						},
					},
				},
			},
			{
				title: "negative count",
				linum: 1,
				count: -1,
				notOK: true,
			},
			{
				title: "out of bounds linum",
				linum: 8,
				count: 1,
				notOK: true,
			},
			{
				title: "not natural linum",
				linum: 0,
				count: 1,
				notOK: true,
			},
		} {
			t.Run(tc.title, func(t *testing.T) {
				got, ok := b.Context(tc.linum, tc.count)
				assert.Equal(t, !tc.notOK, ok)
				if tc.notOK {
					return
				}
				assert.Equal(t, tc.want, got)
			})
		}
	})

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
