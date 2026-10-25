package workflow

import (
	"context"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/permission"
)

// MayRun returns the first step the permissions may not run, "" for none.
func (r *Runner) MayRun(ctx context.Context, set permission.Set, steps []model.WorkflowStep) string {
	return r.mayRun(ctx, set, steps, map[string]bool{})
}

func (r *Runner) mayRun(ctx context.Context, set permission.Set, steps []model.WorkflowStep, seen map[string]bool) string {
	for _, step := range steps {
		switch {
		case step.IsWorkflow():
			if refused := r.mayRunNested(ctx, set, step, seen); refused != "" {
				return refused
			}
		case step.IsPrompt():
			if !set.Allowed("prompt/edit") {
				return "the prompt in " + step.Place()
			}
		default:
			if !r.mayRunJob(set, step) {
				return step.Folder + " / " + step.Job
			}
		}
	}
	return ""
}

// mayRunNested checks the steps of the workflow a step runs, once per
// workflow, as that workflow's own user when it has one.
func (r *Runner) mayRunNested(ctx context.Context, set permission.Set, step model.WorkflowStep, seen map[string]bool) string {
	if seen[step.WorkflowRef] {
		return ""
	}
	seen[step.WorkflowRef] = true
	inner, err := r.workflows.Steps(ctx, step.WorkflowRef)
	if err != nil {
		return "workflow " + step.WorkflowRef
	}
	if r.Users != nil {
		wf, err := r.workflows.Get(ctx, step.WorkflowRef)
		if err != nil {
			return "workflow " + step.WorkflowRef
		}
		if wf.RunAsUserID != "" {
			user, err := r.Users.Get(ctx, wf.RunAsUserID)
			if err != nil {
				return "workflow " + wf.Name + ", whose user is gone"
			}
			set = r.Users.Permissions(ctx, user)
		}
	}
	return r.mayRun(ctx, set, inner, seen)
}

// mayRunJob reports whether the permissions run a job step: a job the
// catalogue holds that is no shell, with `hosts` for a host command and
// the folder's and the job's allow lists otherwise.
func (r *Runner) mayRunJob(set permission.Set, step model.WorkflowStep) bool {
	f := r.catalogue.Find(step.Folder)
	var job *folder.Job
	if f != nil {
		job = f.Job(step.Job)
	}
	switch {
	case job == nil, job.IsShell():
		return false
	case job.IsHost():
		return set.Allowed("hosts")
	}
	return set.Run(f.ID, f.Allow, job.Allow)
}
