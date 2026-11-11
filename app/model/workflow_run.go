package model

import (
	"strings"
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

// AnswerFrom sets the run's answer, the reply to its input, to the markdown of the last prompt or summary step that answered.
func (r *WorkflowRun) AnswerFrom(steps []WorkflowRunStep) {
	r.Answer = ""
	for i := len(steps) - 1; i >= 0 && r.Answer == ""; i-- {
		if steps[i].IsPrompt() {
			r.Answer = strings.TrimSpace(steps[i].Answer)
		}
	}
}
