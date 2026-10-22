package frontend

import (
	"context"
	"log/slog"
	"time"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// activityLimit bounds how many of the latest runs the charts read: more
// than the days of ActivityDays hold on a busy instance, read once per
// page.
const activityLimit = 5000

// activity charts the runs a filter lists that the user may see. The
// duration line is a job's alone: on the runs page it would mix jobs.
func (h *Handlers) activity(ctx context.Context, sees func(*model.Run) bool, filter storage.RunFilter, now time.Time, durations bool) *view.Activity {
	rows, err := h.opts.Runs.List(ctx, filter)
	if err != nil {
		slog.Error("listing runs for the charts", "err", err)
		return nil
	}
	runs := make([]view.ActivityRun, 0, len(rows))
	for _, row := range rows {
		if row.StartedAt == nil || !sees(&row) {
			continue
		}
		runs = append(runs, view.ActivityRun{StartedAt: *row.StartedAt, Elapsed: row.Elapsed(), State: row.State})
	}
	a := view.NewActivity(runs, now, view.ActivityDays)
	if a != nil && !durations {
		a.Durations = nil
	}
	return a
}

// workflowActivity charts a workflow's runs, the duration line with them.
func (h *Handlers) workflowActivity(ctx context.Context, workflowID string, now time.Time) *view.Activity {
	rows, err := h.opts.WorkflowRuns.ListByWorkflow(ctx, workflowID, storage.WorkflowRunFilter{Limit: activityLimit})
	if err != nil {
		slog.Error("listing workflow runs for the charts", "err", err)
		return nil
	}
	runs := make([]view.ActivityRun, 0, len(rows))
	for _, row := range rows {
		if row.StartedAt == nil {
			continue
		}
		runs = append(runs, view.ActivityRun{StartedAt: *row.StartedAt, Elapsed: row.Elapsed(), State: row.State})
	}
	return view.NewActivity(runs, now, view.ActivityDays)
}
