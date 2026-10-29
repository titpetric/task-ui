package workflow

import (
	"context"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/permission"
)

// MayRun returns the first step of the workflow the permissions may not run, "" for none.
// A workflow that runs over a workspace counts its steps in each project, and the steps a project skips for lack of the job not at all.
func (r *Runner) MayRun(ctx context.Context, set permission.Set, wf *model.Workflow, steps []model.WorkflowStep) string {
	expanded, _, projects, err := r.over(ctx, wf, steps)
	if err != nil {
		return err.Error()
	}
	return r.mayRun(ctx, set, expanded, projects, map[string]bool{})
}

func (r *Runner) mayRun(ctx context.Context, set permission.Set, steps []model.WorkflowStep, projects map[string]*folder.Folder, seen map[string]bool) string {
	for _, step := range steps {
		switch {
		case step.IsWorkflow():
			if refused := r.mayRunNested(ctx, set, step, seen); refused != "" {
				return refused
			}
		case step.IsPrompt():
			if !set.Allowed("prompt/edit") {
				return "the prompt in " + model.PlaceLabel(step.Place())
			}
		case step.Disabled:
		default:
			if !r.mayRunJob(set, step, projects) {
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
	return r.mayRun(ctx, set, inner, nil, seen)
}

// mayRunJob reports whether the permissions run a job step: a job the
// catalogue holds that is no shell, with `hosts` for a host command,
// `shell` for a script and the folder's and the job's allow lists
// otherwise. A project of a workspace run over that lacks the job skips
// the step, which takes no permission.
func (r *Runner) mayRunJob(set permission.Set, step model.WorkflowStep, projects map[string]*folder.Folder) bool {
	f := projects[step.Folder]
	if f == nil {
		f = r.catalogue.Find(step.Folder)
	}
	job := r.job(f, step.Job)
	switch {
	case job == nil:
		return projects != nil
	case job.IsShell():
		return false
	case job.IsHost():
		return set.Allowed("hosts")
	case job.IsScript():
		return set.Allowed("shell")
	}
	return set.Run(f.ID, f.Allow, job.Allow)
}
