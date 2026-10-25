package workflow

import (
	"context"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/permission"
)

// MayRun returns the first step a user with the permissions may not
// run, "" when the user may run them all, the steps of the workflows the
// steps run too; a step whose folder or job nothing holds counts as one.
func (r *Runner) MayRun(ctx context.Context, set permission.Set, steps []model.WorkflowStep) string {
	flat, err := r.workflows.Expand(ctx, steps)
	if err != nil {
		flat = steps
	}
	for _, step := range flat {
		if step.IsPrompt() {
			if !set.Allowed("prompt/edit") {
				return "the prompt in " + step.Place()
			}
			continue
		}
		f := r.catalogue.Find(step.Folder)
		var job *folder.Job
		if f != nil {
			job = f.Job(step.Job)
		}
		switch {
		case job == nil, job.IsShell():
			return step.Folder + " / " + step.Job
		case job.IsHost():
			if !set.Allowed("hosts") {
				return step.Folder + " / " + step.Job
			}
		case !set.Run(f.ID, f.Allow, job.Allow):
			return step.Folder + " / " + step.Job
		}
	}
	return ""
}
