package view

// ChartLegend is a series under its chart: its name and colour, and the
// least, the mean and the most of its points, worded in the chart's
// unit.
type ChartLegend struct {
	Name  string
	Color int
	Min   string
	Avg   string
	Max   string
}
