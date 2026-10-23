package view

import (
	"sort"
	"time"

	"github.com/titpetric/task-ui/model"
)

// Recent is an entry of the sidebar's recents: a run of a job, or a run
// of a workflow.
type Recent struct {
	Run  *Run
	Flow *model.WorkflowRun
}

// StartedAt is when the entry's run started.
func (r Recent) StartedAt() time.Time {
	var at *time.Time
	if r.Run != nil {
		at = r.Run.StartedAt
	} else if r.Flow != nil {
		at = r.Flow.StartedAt
	}
	if at == nil {
		return time.Time{}
	}
	return *at
}

// Elapsed is how long the entry's run took, or has taken so far.
func (r Recent) Elapsed() time.Duration {
	if r.Run != nil {
		return r.Run.Elapsed()
	}
	return r.Flow.Elapsed()
}

// Initials returns the initials of the user who started the entry's run.
func (r Recent) Initials() string {
	if r.Run != nil {
		return r.Run.Initials()
	}
	return (&model.User{Name: r.Flow.UserName}).Initials()
}

// Recents merges the runs of jobs and of workflows, newest first, and keeps the first n.
func Recents(runs []Run, flows []model.WorkflowRun, n int) []Recent {
	result := make([]Recent, 0, len(runs)+len(flows))
	for i := range runs {
		result = append(result, Recent{Run: &runs[i]})
	}
	for i := range flows {
		result = append(result, Recent{Flow: &flows[i]})
	}
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].StartedAt().After(result[j].StartedAt())
	})
	if len(result) > n {
		result = result[:n]
	}
	return result
}
