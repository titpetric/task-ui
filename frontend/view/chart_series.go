package view

// ChartSeries is a line of a chart: its name, its colour, one of the
// six series colours, and its points, oldest first.
type ChartSeries struct {
	Name   string
	Color  int
	Points []ChartPoint
}
