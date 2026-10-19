package model

import (
	"time"
)

// Elapsed is how long the workflow run took, or has taken so far.
func (r *WorkflowRun) Elapsed() time.Duration {
	if r == nil || r.StartedAt == nil {
		return 0
	}
	if r.FinishedAt != nil {
		return r.FinishedAt.Sub(*r.StartedAt)
	}
	return time.Since(*r.StartedAt)
}
