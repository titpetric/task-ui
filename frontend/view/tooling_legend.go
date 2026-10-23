package view

import (
	"time"
)

// ToolingLegend is one entry under a chart: a top command of the
// window, or the rest together as other, with its colour, its time and
// its runs over the window.
type ToolingLegend struct {
	Name  string
	Color string
	Runs  int64
	Total time.Duration
}
