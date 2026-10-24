package model

import (
	"time"
)

// Due reports whether the schedule should start a run at now: the schedule
// is on and its next run is not in the future.
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

// ArgList returns the arguments the schedule's runs pass the job, the command of @host:<host>.
func (s *Schedule) ArgList() []string {
	return DecodeList(s.Args)
}

// Triages reports whether a failed run of the schedule gets Claude's triage.
func (s *Schedule) Triages() bool {
	return s != nil && s.Triage != 0
}
