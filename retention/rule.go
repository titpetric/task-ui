// Package retention keeps what task-ui records within the rules of the
// Data retention settings: the recordings of runs, the history of the
// artifact tree and the secret read log, each kept for an age and within
// a size, the oldest going first. task-ui cleanup, the root's @cleanup,
// applies the rules; the settings page measures what each kind holds.
package retention

import (
	"time"
)

// The ages a rule keeps an item for, as the settings page lists them.
const (
	Day   = "24h"
	Week  = "7d"
	Month = "31d"
	Never = "never"
)

// Ages are the ages a rule takes, shortest first.
var Ages = []string{Day, Week, Month, Never}

// ages are how long each age keeps an item.
var ages = map[string]time.Duration{Day: 24 * time.Hour, Week: 7 * 24 * time.Hour, Month: 31 * 24 * time.Hour}

// Rule says how long a kind of data stays and how much of it: Age, one
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

// Valid reports whether the rule names an age of Ages and no negative
// limit.
func (r Rule) Valid() bool {
	_, ok := ages[r.Age]
	return (ok || r.Age == Never) && r.Limit >= 0
}

// Says reads the rule: "keep for 31 days, at most 1.0 GB".
func (r Rule) Says() string {
	keep := map[string]string{Day: "keep for 24 hours", Week: "keep for 7 days", Month: "keep for 31 days"}[r.Age]
	if keep == "" {
		keep = "keep for ever"
	}
	if r.Limit > 0 {
		return keep + ", at most " + FormatSize(r.Limit)
	}
	return keep + ", no size limit"
}
