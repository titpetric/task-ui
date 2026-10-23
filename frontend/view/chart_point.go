package view

import (
	"time"
)

// ChartPoint is a measure of a chart at a time; a NaN value is a gap.
type ChartPoint struct {
	At time.Time
	V  float64
}
