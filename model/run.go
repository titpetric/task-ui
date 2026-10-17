package model

import (
	"time"
)

// Key is the identity the run lock uses: one run per job per folder.
func (r *Run) Key() string {
	return r.Folder + "/" + r.Job
}

// Finished reports whether the run ended.
func (r *Run) Finished() bool {
	return r != nil && Finished(r.State)
}

// Elapsed is how long the run took, or has been going.
func (r *Run) Elapsed() time.Duration {
	if r == nil || r.StartedAt == nil {
		return 0
	}
	if r.FinishedAt != nil {
		return r.FinishedAt.Sub(*r.StartedAt)
	}
	return time.Since(*r.StartedAt)
}
