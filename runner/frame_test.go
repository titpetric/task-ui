package runner_test

import (
	"strings"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/runner"
)

func TestFrame_Text(t *testing.T) {
	t.Parallel()
	frame := runner.Frame{Cols: 4, Rows: 2, Lines: [][]runner.Cell{
		{{Char: 'h'}, {Char: 'i'}, {}, {}},
		{{}, {}, {}, {}},
	}}
	assert.Equal(t, "hi\n\n", frame.Text())
}

func TestFrame_SVG(t *testing.T) {
	t.Parallel()
	frame := runner.Frame{Cols: 3, Rows: 1, CursorX: 2, Lines: [][]runner.Cell{
		{{Char: '<', FG: runner.DefaultFG, BG: runner.DefaultBG}, {Char: 'x', FG: 1, BG: 4, Bold: true}, {Char: 'y', FG: 2, BG: 2, Reverse: true}},
	}}
	svg := frame.SVG(runner.DarkPalette)
	assert.True(t, strings.HasPrefix(svg, "<svg "))
	assert.True(t, strings.HasSuffix(svg, "</svg>"))
	assert.Contains(t, svg, "&lt;", "text is escaped")
	assert.Contains(t, svg, runner.DarkPalette.ANSI[1])
	assert.Contains(t, svg, `fill="`+runner.DarkPalette.ANSI[4]+`"`, "a run with its own background gets a rect")
	assert.Contains(t, svg, `font-weight="bold"`)
	assert.Contains(t, svg, `opacity="0.6"`, "the cursor")
}

func TestPalette_Color(t *testing.T) {
	t.Parallel()
	p := runner.DarkPalette
	assert.Equal(t, p.FG, p.Color(runner.DefaultFG, true))
	assert.Equal(t, p.BG, p.Color(runner.DefaultBG, false))
	assert.Equal(t, p.ANSI[9], p.Color(9, true))
	assert.Equal(t, "#000000", p.Color(16, true), "the cube starts black")
	assert.Equal(t, "#ffffff", p.Color(231, true), "and ends white")
	assert.Equal(t, "#080808", p.Color(232, true), "grays")
	assert.Equal(t, "#eeeeee", p.Color(255, true))
	assert.Equal(t, p.FG, p.Color(9999, true), "unknown foreground")
	assert.Equal(t, p.BG, p.Color(9999, false), "unknown background")
}

func TestCell(t *testing.T) {
	t.Parallel()
	cell := runner.Cell{Char: 'a', FG: runner.DefaultFG, BG: runner.DefaultBG}
	assert.Equal(t, 'a', cell.Char)
	assert.False(t, cell.Bold)
}

func TestFrame_Trim(t *testing.T) {
	t.Parallel()
	row := func(text string) []runner.Cell {
		cells := make([]runner.Cell, 4)
		for i, r := range text {
			cells[i] = runner.Cell{Char: r, BG: runner.DefaultBG}
		}
		return cells
	}
	frame := runner.Frame{Cols: 4, Rows: 6, CursorY: 1, Lines: [][]runner.Cell{row("one"), row(""), row("two"), row(""), row(""), row("")}}

	short := frame.Trim(2)
	assert.Equal(t, 3, short.Rows)
	assert.Len(t, short.Lines, 3)
	assert.Equal(t, "one\n\ntwo\n", short.Text())

	// At least min rows, and the cursor's row, stay.
	assert.Equal(t, 5, frame.Trim(5).Rows)
	cursor := frame
	cursor.CursorY = 4
	assert.Equal(t, 5, cursor.Trim(1).Rows)
	assert.Equal(t, 6, frame.Trim(10).Rows, "at most the rows there are")

	// A painted empty row counts as content.
	painted := frame
	painted.Lines[5][0].BG = 4
	assert.Equal(t, 6, painted.Trim(1).Rows)
}

// A text frame wraps its text to the width and shows no cursor.
func TestTextFrame(t *testing.T) {
	t.Parallel()
	f := runner.TextFrame(12, 2, "the recording expired today")
	assert.Equal(t, "the\nrecording\nexpired\ntoday", strings.TrimSpace(f.Text()))
	assert.Equal(t, 4, f.Rows)
	assert.NotContains(t, f.SVG(runner.DarkPalette), `opacity="0.6"`)
	assert.Equal(t, 80, runner.TextFrame(0, 1, "x").Cols)
}
