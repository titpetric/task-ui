package view

import (
	"strconv"
)

// ChartUnit is what a chart measures: a fixed top of its scale, or none
// to fit the data, and how it words a value, scaled by the largest one.
type ChartUnit struct {
	Max   float64
	Scale func(largest float64) (divisor float64, suffix string)
}

// The units of the inventory's charts.
var (
	// ChartPercent is a share, 0 to 100.
	ChartPercent = ChartUnit{Max: 100, Scale: fixed("%")}
	// ChartCores is a count of busy cores.
	ChartCores = ChartUnit{Scale: fixed(" cores")}
	// ChartGiB is a size in GB as Beszel counts one, 2^30 bytes.
	ChartGiB = ChartUnit{Scale: steps(1024, " GiB", " TiB")}
	// ChartMB is a size in MB, as Beszel counts container memory.
	ChartMB = ChartUnit{Scale: steps(1024, " MB", " GB")}
	// ChartRate is traffic in bytes a second.
	ChartRate = ChartUnit{Scale: steps(1024, " B/s", " KB/s", " MB/s", " GB/s")}
)

// fixed words every value in one suffix.
func fixed(suffix string) func(float64) (float64, string) {
	return func(float64) (float64, string) { return 1, suffix }
}

// steps words values in the largest suffix the largest value reaches,
// each base times the one before.
func steps(base float64, suffixes ...string) func(float64) (float64, string) {
	return func(largest float64) (float64, string) {
		divisor := 1.0
		for i := range suffixes[:len(suffixes)-1] {
			if largest < divisor*base {
				return divisor, suffixes[i]
			}
			divisor *= base
		}
		return divisor, suffixes[len(suffixes)-1]
	}
}

// word words a value by a divisor and a suffix: one decimal under ten,
// whole above.
func word(v, divisor float64, suffix string) string {
	v /= divisor
	if v < 10 && v != float64(int64(v)) {
		return strconv.FormatFloat(v, 'f', 1, 64) + suffix
	}
	return strconv.FormatFloat(v, 'f', 0, 64) + suffix
}
