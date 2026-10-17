package model

import (
	"time"
)

// Due reports whether the schedule should start a run at now: it is
// enabled and its next run is not in the future.
func (s *Schedule) Due(now time.Time) bool {
	if s == nil || s.Enabled == 0 || s.NextRunAt == nil {
		return false
	}
	return !s.NextRunAt.After(now)
}

// IsEnabled reports whether the schedule is on.
func (s *Schedule) IsEnabled() bool {
	return s != nil && s.Enabled != 0
}
