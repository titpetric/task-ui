package workflow

import (
	"context"
	"errors"

	"github.com/titpetric/task-ui/model"
)

// prompt runs the workflow's prompt once the steps ended, through the
// Prompter, and returns when Claude replied. The run keeps the steps'
// state, or fails with a prompt that fails after steps that ended ok; a
// stop during the prompt stops the run.
func (r *Runner) prompt(ctx context.Context, a *active, user *model.User) {
	run := a.run
	var err error
	if r.Prompter == nil {
		err = errors.New("prompts aren't available on this instance")
	} else {
		_, err = r.Prompter.AskWorkflow(ctx, user, run, a.workflow, a.input)
	}
	switch {
	case err == nil:
	case ctx.Err() != nil:
		run.State, run.Message = model.StateStopped, "stopped at the prompt"
	case run.State == model.StateOK:
		run.State, run.Message = model.StateFailed, "the prompt: "+err.Error()
	}
}

// promptStep puts a prompt step to Claude through the Prompter and
// waits for the reply: the step ends ok with it when Claude ended ok,
// stopped after a stop, and failed otherwise.
func (r *Runner) promptStep(ctx context.Context, a *active, row *model.WorkflowRunStep, step Step, user *model.User) (state, runID, answer, message string) {
	if r.Prompter == nil {
		return model.StateFailed, "", "", "prompts aren't available on this instance"
	}
	runID, answer, err := r.Prompter.AskStep(ctx, user, a.run, row, step.Place, a.input)
	switch {
	case err == nil:
		return model.StateOK, runID, answer, ""
	case ctx.Err() != nil:
		return model.StateStopped, runID, "", ""
	}
	return model.StateFailed, runID, "", err.Error()
}
