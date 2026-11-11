package hostinfo

// Segment is a host's part of a pool: its total and used, Width its
// share of the pool's total and Fill the share of it used, in percent;
// Color the host's colour, its place among the hosts that report, and
// Title what a hover says.
type Segment struct {
	Host  string
	Color int
	Used  float64
	Total float64
	Width float64
	Fill  float64
	Title string
}
