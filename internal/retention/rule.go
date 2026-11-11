package retention

import (
	"strconv"
	"time"
)

// The ages a rule keeps an item for, as the settings page lists them.
const (
	Day     = "24h"
	Week    = "7d"
	Month   = "31d"
	Quarter = "90d"
	Year    = "365d"
	Never   = "never"
)

// Ages are the ages a rule takes, shortest first.
var Ages = []string{Day, Week, Month, Quarter, Year, Never}

// ages are how long each age keeps an item.
var ages = map[string]time.Duration{Day: 24 * time.Hour, Week: 7 * 24 * time.Hour, Month: 31 * 24 * time.Hour, Quarter: 90 * 24 * time.Hour, Year: 365 * 24 * time.Hour}

// labels name the ages as the page and Says read them.
var labels = map[string]string{Day: "24 hours", Week: "7 days", Month: "31 days", Quarter: "90 days", Year: "365 days", Never: "Never"}

// Label names an age as the page lists it: "31 days", "Never".
func Label(age string) string {
	return labels[age]
}

// Limits are the size limits the page lists, in bytes, besides no limit.
var Limits = []int64{100 << 20, 500 << 20, 1 << 30, 5 << 30, 10 << 30, 50 << 30}

// SizeLabel names a size limit as the page lists it, whole units without
// a decimal, "100 MB", and FormatSize for any other.
func SizeLabel(n int64) string {
	for _, unit := range []struct {
		bytes  int64
		suffix string
	}{{1 << 40, "TB"}, {1 << 30, "GB"}, {1 << 20, "MB"}, {1 << 10, "KB"}} {
		if n >= unit.bytes && n%unit.bytes == 0 {
			return strconv.FormatInt(n/unit.bytes, 10) + " " + unit.suffix
		}
		if n >= unit.bytes {
			return strconv.FormatFloat(float64(n)/float64(unit.bytes), 'f', 1, 64) + " " + unit.suffix
		}
	}
	return FormatSize(n)
}

// Rule says how long an area's data stays and how much of it: Age, one
// of Ages, and Limit, the bytes it may take, none when zero. Past
// either, the oldest items go first until both hold.
type Rule struct {
	Age   string `json:"age"`
	Limit int64  `json:"limit"`
}

// Keep returns how long the rule keeps an item, zero for ever.
func (r Rule) Keep() time.Duration {
	return ages[r.Age]
}

// Valid reports whether the rule names an age of Ages and no negative limit.
func (r Rule) Valid() bool {
	_, ok := ages[r.Age]
	return (ok || r.Age == Never) && r.Limit >= 0
}

// Keeps reads the rule's age: "keep for 31 days", "keep for ever".
func (r Rule) Keeps() string {
	if r.Age == Never || labels[r.Age] == "" {
		return "keep for ever"
	}
	return "keep for " + labels[r.Age]
}

// Says reads the rule: "keep for 31 days, at most 1.0 GB".
func (r Rule) Says() string {
	if r.Limit > 0 {
		return r.Keeps() + ", at most " + FormatSize(r.Limit)
	}
	return r.Keeps() + ", no size limit"
}
