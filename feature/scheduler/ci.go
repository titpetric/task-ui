package scheduler

import (
	"context"

	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/storage"
)

// The name and expression a source-declared daily CI target is scheduled as.
const (
	CISource     = "ci:daily"
	CIExpression = "@daily"
)

// CISchedules is the part of schedule storage reconciliation writes through.
type CISchedules interface {
	List(ctx context.Context) ([]model.Schedule, error)
	Create(ctx context.Context, schedule *model.Schedule) error
	Update(ctx context.Context, schedule *model.Schedule) error
}

// CISearch is the part of the search index reconciliation reads targets from.
type CISearch interface {
	Targets(ctx context.Context, name string) ([]storage.SearchDocument, error)
}

// CIReconciler registers a disabled daily schedule for every indexed ci:daily
// target, for an admin to approve once, and disables the rows whose target left.
type CIReconciler struct {
	Schedules CISchedules
	Search    CISearch
}

// Run adds the schedules the index asks for and answers how many it added.
func (r *CIReconciler) Run(ctx context.Context) (int, error) {
	if r == nil || r.Schedules == nil || r.Search == nil {
		return 0, nil
	}
	schedules, err := r.Schedules.List(ctx)
	if err != nil {
		return 0, err
	}
	existing := make(map[string]bool, len(schedules))
	for _, schedule := range schedules {
		existing[ciKey(schedule.Folder, schedule.Job)] = true
	}
	targets, err := r.Search.Targets(ctx, CISource)
	if err != nil {
		return 0, err
	}
	wanted := make(map[string]bool, len(targets))
	added := 0
	for _, target := range targets {
		if target.Dir != "" || target.FolderID == "" || target.Target == "" {
			continue
		}
		key := ciKey(target.FolderID, target.Target)
		wanted[key] = true
		if existing[key] {
			continue
		}
		schedule := &model.Schedule{Folder: target.FolderID, Job: target.Target, Expression: CIExpression, Enabled: 0, Source: CISource}
		if err := r.Schedules.Create(ctx, schedule); err != nil {
			return added, err
		}
		existing[key] = true
		added++
	}
	for i := range schedules {
		schedule := &schedules[i]
		if schedule.Source != CISource || schedule.Enabled == 0 || wanted[ciKey(schedule.Folder, schedule.Job)] {
			continue
		}
		schedule.Enabled = 0
		if err := r.Schedules.Update(ctx, schedule); err != nil {
			return added, err
		}
	}
	return added, nil
}

// ciKey names a folder and job once.
func ciKey(folder, job string) string {
	return folder + "\x00" + job
}
