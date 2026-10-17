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
