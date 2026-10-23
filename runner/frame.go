package runner

import (
	"fmt"
	"html"
	"strings"
)

// Frame is what a terminal shows at one moment: the cells of the screen
// with their colours, and where the cursor is. The terminal emulator
// every session feeds produces it, so it matches for a program
// that redraws the screen as much as for one that prints lines.
type Frame struct {
	Cols, Rows int
	Lines      [][]Cell
	CursorX    int
	CursorY    int
	Title      string
}

// Text renders the frame as plain lines, trailing spaces trimmed.
func (f Frame) Text() string {
	var b strings.Builder
	for _, line := range f.Lines {
		var row strings.Builder
		for _, cell := range line {
			if cell.Char == 0 {
				row.WriteRune(' ')
			} else {
				row.WriteRune(cell.Char)
			}
		}
		b.WriteString(strings.TrimRight(row.String(), " "))
		b.WriteByte('\n')
	}
	return b.String()
}

// Trim returns the frame without its trailing empty rows, keeping at least min rows and the cursor's row.
// A tile shows this much of a terminal that has only begun.
func (f Frame) Trim(min int) Frame {
	rows := f.CursorY + 1
	for y := len(f.Lines) - 1; y >= rows; y-- {
		if !blank(f.Lines[y]) {
			rows = y + 1
			break
		}
	}
	if rows < min {
		rows = min
	}
	if rows >= f.Rows {
		return f
	}
	trimmed := f
	trimmed.Rows = rows
	trimmed.Lines = f.Lines[:rows]
	return trimmed
}

// blank reports whether a row shows nothing.
func blank(line []Cell) bool {
	for _, cell := range line {
		if cell.Char != 0 && cell.Char != ' ' {
			return false
		}
		if cell.Reverse || (cell.BG != DefaultBG && cell.BG != 0) {
			return false
		}
	}
	return true
}

// SVG renders the frame as an image that scales: a text row per line, a tspan and a background rect per colour run.
func (f Frame) SVG(p Palette) string {
	const cw, ch = 8.4, 17.0
	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %.1f %.1f" font-family="JetBrains Mono, ui-monospace, SFMono-Regular, Menlo, Consolas, monospace" font-size="14">`,
		float64(f.Cols)*cw, float64(f.Rows)*ch)
	fmt.Fprintf(&b, `<rect width="100%%" height="100%%" fill="%s"/>`, p.BG)
	for y, line := range f.Lines {
		f.svgLine(&b, p, y, line, cw, ch)
	}
	if f.CursorY < f.Rows && f.CursorX < f.Cols {
		fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s" opacity="0.6"/>`,
			float64(f.CursorX)*cw, float64(f.CursorY)*ch, cw, ch, p.ANSI[3])
	}
	b.WriteString(`</svg>`)
	return b.String()
}

// svgLine renders one row: runs of cells sharing colours and weight.
func (f Frame) svgLine(b *strings.Builder, p Palette, y int, line []Cell, cw, ch float64) {
	fmt.Fprintf(b, `<text y="%.1f" xml:space="preserve">`, float64(y)*ch+ch-4)
	start := 0
	for start < len(line) {
		end := start + 1
		for end < len(line) && sameRun(line[start], line[end]) {
			end++
		}
		cell := line[start]
		fg, bg := p.Color(cell.FG, true), p.Color(cell.BG, false)
		if cell.Reverse {
			fg, bg = bg, fg
		}
		text := runText(line[start:end])
		if bg != p.BG {
			fmt.Fprintf(b, `</text><rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s"/><text y="%.1f" xml:space="preserve">`,
				float64(start)*cw, float64(y)*ch, float64(end-start)*cw, ch, bg, float64(y)*ch+ch-4)
		}
		weight := ""
		if cell.Bold {
			weight = ` font-weight="bold"`
		}
		fmt.Fprintf(b, `<tspan x="%.1f" fill="%s"%s>%s</tspan>`, float64(start)*cw, fg, weight, html.EscapeString(text))
		start = end
	}
	b.WriteString(`</text>`)
}

func sameRun(a, b Cell) bool {
	return a.FG == b.FG && a.BG == b.BG && a.Bold == b.Bold && a.Reverse == b.Reverse
}

func runText(cells []Cell) string {
	var b strings.Builder
	for _, cell := range cells {
		if cell.Char == 0 {
			b.WriteRune(' ')
		} else {
			b.WriteRune(cell.Char)
		}
	}
	return b.String()
}

// TextFrame shows a text in a frame of a size, wrapped to the width, no cursor, for a run with no recording.
func TextFrame(cols, rows int, text string) Frame {
	if cols <= 0 {
		cols = 80
	}
	f := Frame{Cols: cols, Rows: rows, CursorX: cols, CursorY: rows}
	for _, paragraph := range strings.Split(text, "\n") {
		line := []Cell{}
		for _, word := range strings.Fields(paragraph) {
			if len(line) > 0 && len(line)+1+len([]rune(word)) > cols {
				f.Lines = append(f.Lines, line)
				line = []Cell{}
			}
			if len(line) > 0 {
				line = append(line, Cell{Char: ' ', FG: DefaultFG, BG: DefaultBG})
			}
			for _, r := range word {
				line = append(line, Cell{Char: r, FG: DefaultFG, BG: DefaultBG})
			}
		}
		f.Lines = append(f.Lines, line)
	}
	if f.Rows < len(f.Lines) {
		f.Rows = len(f.Lines)
		f.CursorY = f.Rows
	}
	return f
}
