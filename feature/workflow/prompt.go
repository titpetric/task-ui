package workflow

import (
	"context"

	"github.com/titpetric/task-ui/model"
)

// promptStep puts a prompt or summary step to Claude through the
// Prompter and waits for the reply: the step ends ok with it when Claude
// ended ok, stopped after a stop, and failed otherwise.
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
