package workflow

import (
	"context"
	"fmt"
	"strings"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/model"
)

// Resolve finds what every enabled step runs, or says which one is missing.
func (r *Runner) Resolve(ctx context.Context, steps []model.WorkflowStep) ([]Step, error) {
	return r.resolve(ctx, steps, make([]int, len(steps)), nil)
}

// resolve finds what each step runs, with its project's index. Given the
// projects of a workflow that runs over a workspace, a step finds its
// folder among them first, and a project without the step's job skips it.
func (r *Runner) resolve(ctx context.Context, steps []model.WorkflowStep, groups []int, projects map[string]*folder.Folder) ([]Step, error) {
	resolved := make([]Step, 0, len(steps))
	for i, step := range steps {
		next, err := r.resolveStep(ctx, step, projects)
		if err != nil {
			return nil, err
		}
		next.Group = groups[i]
		resolved = append(resolved, next)
	}
	return resolved, nil
}

// resolveStep finds what one step runs.
func (r *Runner) resolveStep(ctx context.Context, step model.WorkflowStep, projects map[string]*folder.Folder) (Step, error) {
	switch {
	case step.Disabled:
		return Step{Disabled: true}, nil
	case step.IsPrompt():
		return Step{Asks: true, Summary: step.IsSummary(), Place: step.Place()}, nil
	case step.IsWorkflow():
		wf, err := r.workflows.Get(ctx, step.WorkflowRef)
		if err != nil {
			return Step{}, fmt.Errorf("no workflow %s", step.WorkflowRef)
		}
		return Step{Workflow: wf}, nil
	}
	f := projects[step.Folder]
	if f == nil {
		f = r.catalogue.Find(step.Folder)
	}
	args := model.DecodeList(step.Args)
	job := r.job(f, step.Job)
	switch {
	case job == nil && projects != nil:
		return Step{Disabled: true}, nil
	case job == nil:
		return Step{}, fmt.Errorf("no job %s / %s", step.Folder, step.Job)
	case job.IsShell():
		return Step{}, fmt.Errorf("%s / %s: %w", step.Folder, step.Job, folder.ErrShellUnattended)
	case job.IsHost() && strings.TrimSpace(strings.Join(args, " ")) == "":
		return Step{}, fmt.Errorf("%s: %w", step.Job, ErrHostStep)
	case job.IsScript() && strings.TrimSpace(strings.Join(args, " ")) == "":
		return Step{}, fmt.Errorf("%s: %w", step.Folder, folder.ErrScriptEmpty)
	}
	return Step{Folder: f, Job: job, Args: args}, nil
}

// job returns a folder's job by ID, the script job for @script.
func (r *Runner) job(f *folder.Folder, id string) *folder.Job {
	if id == folder.ScriptID {
		return f.ScriptJob()
	}
	return f.Job(id)
}
