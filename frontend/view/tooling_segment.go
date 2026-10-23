package view

import (
	"html/template"
)

// ToolingSegment is one command's share of a day's bar.
type ToolingSegment struct {
	// Name is the command key, or the tool, or "other" for the rest.
	Name string

	// Color is its colour class: s0 to s5, or s-other.
	Color string

	// Title is what a hover over it says: the day, the name, the time
	// and the runs.
	Title string

	// Style is its height, in percent of the day.
	Style template.CSS
}
