package hostinfo

import (
	"fmt"
)

// Pool is a resource of the cluster summed over the hosts that report:
// how much there is and the part in use, in Unit, and a segment a host,
// as wide as its share of the whole. Figure is the sum as the page says
// it, "41.2 / 64 GiB".
type Pool struct {
	Name     string
	Unit     string
	Used     float64
	Total    float64
	Figure   string
	Segments []Segment
}

// Percent is the share of the pool in use; 0 for a pool without a total.
func (p Pool) Percent() float64 {
	if p.Total <= 0 {
		return 0
	}
	return p.Used * 100 / p.Total
}

// finish sizes the segments by their share of the total and words the
// figure.
func (p *Pool) finish() {
	for i := range p.Segments {
		s := &p.Segments[i]
		if p.Total > 0 {
			s.Width = s.Total * 100 / p.Total
		}
		if s.Total > 0 {
			s.Fill = min(100, s.Used*100/s.Total)
		}
		s.Title = fmt.Sprintf("%s: %s of %s", s.Host, amount(s.Used, p.Unit), amount(s.Total, p.Unit))
	}
	p.Figure = amount(p.Used, "") + " / " + amount(p.Total, p.Unit)
	if p.Unit == "GiB" {
		p.Figure = sizeFigure(p.Used, p.Total)
	}
}

// amount words a quantity in a unit: whole cores, GiB to one decimal or
// TiB past a thousand.
func amount(v float64, unit string) string {
	switch unit {
	case "GiB":
		if v >= 1000 {
			return fmt.Sprintf("%.1f TiB", v/1024)
		}
		return fmt.Sprintf("%.1f GiB", v)
	case "":
		return trim(v)
	default:
		return trim(v) + " " + unit
	}
}

// sizeFigure words a used and a total size in the total's unit.
func sizeFigure(used, total float64) string {
	if total >= 1000 {
		return fmt.Sprintf("%.1f / %.1f TiB", used/1024, total/1024)
	}
	return fmt.Sprintf("%.1f / %.1f GiB", used, total)
}

// trim words a number to one decimal, none for a whole one.
func trim(v float64) string {
	if v == float64(int64(v)) {
		return fmt.Sprintf("%d", int64(v))
	}
	return fmt.Sprintf("%.1f", v)
}
