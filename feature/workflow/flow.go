package workflow

import (
	"context"
	"errors"

	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/permission"
)

// ErrRunAsGone refuses a run of a workflow whose run-as user nobody holds any more.
var ErrRunAsGone = errors.New("the user this workflow runs as is gone; change it on the workflow")

// Flow is a workflow with its steps, bound to the runner: who its runs
// run as, which step a user may not run, and its start. A handler loads
// one per request and asks it.
type Flow struct {
	workflow *model.Workflow
	steps    []model.WorkflowStep
	runner   *Runner
}

// Flow binds a workflow and its steps to the runner.
func (r *Runner) Flow(wf *model.Workflow, steps []model.WorkflowStep) *Flow {
	return &Flow{workflow: wf, steps: steps, runner: r}
}

// FlowByID returns the workflow by ID with its steps, bound to the runner.
func (r *Runner) FlowByID(ctx context.Context, id string) (*Flow, error) {
	wf, err := r.workflows.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	steps, err := r.workflows.Steps(ctx, wf.ID)
	if err != nil {
		return nil, err
	}
	return r.Flow(wf, steps), nil
}

// Workflow returns the workflow's row.
func (f *Flow) Workflow() *model.Workflow {
	return f.workflow
}

// Steps returns the workflow's steps in order.
func (f *Flow) Steps() []model.WorkflowStep {
	return f.steps
}

// RunAs returns the user the workflow's runs run as: the one set on it, else starter.
// ErrRunAsGone for a set user nobody holds; without Users, the set user by ID and the workflow's owner name.
func (f *Flow) RunAs(ctx context.Context, starter *model.User) (*model.User, error) {
	switch {
	case f.workflow.RunAsUserID == "":
		return starter, nil
	case f.runner.Users == nil:
		return &model.User{ID: f.workflow.RunAsUserID, Name: f.workflow.UserName}, nil
	}
	user, err := f.runner.Users.Get(ctx, f.workflow.RunAsUserID)
	if err != nil {
		return nil, ErrRunAsGone
	}
	return user, nil
}

// RunAsName names the user set on the workflow, "" for whoever starts it and for a user nobody holds.
func (f *Flow) RunAsName(ctx context.Context) string {
	if f.workflow.RunAsUserID == "" {
		return ""
	}
	user, err := f.RunAs(ctx, nil)
	if err != nil {
		return ""
	}
	return user.Name
}

// MayRun returns the first step the permissions may not run, "" for none, as Runner.MayRun says.
func (f *Flow) MayRun(ctx context.Context, set permission.Set) string {
	return f.runner.MayRun(ctx, set, f.workflow, f.steps)
}

// Check returns the user a start by starter runs as, with set starter's permissions, and the first step either may not run.
// A zero Refusal lets the start go; ErrRunAsGone refuses it.
func (f *Flow) Check(ctx context.Context, starter *model.User, set permission.Set) (*model.User, Refusal, error) {
	if step := f.MayRun(ctx, set); step != "" {
		return nil, Refusal{Step: step}, nil
	}
	runAs, err := f.RunAs(ctx, starter)
	if err != nil {
		return nil, Refusal{}, err
	}
	if f.runner.Users != nil && runAs.ID != starter.ID {
		if step := f.MayRun(ctx, f.runner.Users.Permissions(ctx, runAs)); step != "" {
			return runAs, Refusal{Step: step, User: runAs.Name}, nil
		}
	}
	return runAs, Refusal{}, nil
}

// Start begins a run as the user Check returns, with the run's input.
func (f *Flow) Start(ctx context.Context, runAs *model.User, input string) (*model.WorkflowRun, error) {
	return f.runner.Start(ctx, f.workflow, runAs, input)
}
