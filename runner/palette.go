package runner

import (
	"fmt"
)

// Palette maps terminal colours to CSS colours for the SVG.
type Palette struct {
	FG, BG string
	ANSI   [16]string
}

// DarkPalette is the console's own colours.
var DarkPalette = Palette{
	FG: "#d5dee7", BG: "#090c10",
	ANSI: [16]string{"#10161d", "#e06c6c", "#57b894", "#e0a44f", "#6ea8dc", "#c38fd9", "#5fc1c9", "#d5dee7",
		"#5a6b79", "#f08a8a", "#78d1ae", "#f0c070", "#93c1ea", "#d7aee8", "#86d8de", "#ffffff"},
}

// Color returns the CSS colour of a terminal colour.
func (p Palette) Color(c uint32, foreground bool) string {
	switch {
	case c == DefaultFG:
		return p.FG
	case c == DefaultBG:
		return p.BG
	case c < 16:
		return p.ANSI[c]
	case c < 232:
		c -= 16
		levels := [6]int{0, 95, 135, 175, 215, 255}
		return fmt.Sprintf("#%02x%02x%02x", levels[c/36], levels[c/6%6], levels[c%6])
	case c < 256:
		gray := 8 + int(c-232)*10
		return fmt.Sprintf("#%02x%02x%02x", gray, gray, gray)
	}
	if foreground {
		return p.FG
	}
	return p.BG
}
