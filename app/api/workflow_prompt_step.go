package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/model"
)

// promptStep turns a requested prompt or summary step into its row: the
// prompt, none for a summary, asked in a folder and a directory inside
// it the user sees, in the root, internal, when it names that, or in no
// workspace, the workspaces root, for one without a folder. It names no
// job or workflow.
func (h *Handlers) promptStep(r *http.Request, step StepRequest) (model.WorkflowStep, error) {
	row := model.WorkflowStep{Kind: step.StepKind(), Folder: model.Canonical(strings.Trim(step.Folder, "/")), Dir: strings.Trim(step.Dir, "/"), Job: model.PromptJob, Prompt: strings.TrimSpace(step.Prompt)}
	if row.Folder == "" {
		row.Dir = ""
	}
	switch {
	case step.Job != "" || step.Workflow != "":
		return row, errors.New("a prompt or summary step names no job or workflow")
	case row.IsSummary() && row.Prompt != "":
		return row, errors.New("a summary step takes no prompt: Claude summarises the steps before it")
	case !row.IsSummary() && row.Prompt == "":
		return row, errors.New("a prompt step needs its prompt")
	case h.opts.Board == nil:
		return row, errors.New("prompts aren't available on this instance")
	}
	if err := h.opts.Board.PromptPlace(r.Context(), access.UserFrom(r.Context()), row.Place()); err != nil {
		return row, fmt.Errorf("can't run the prompt there: %w", err)
	}
	return row, nil
}

// rowSteps returns the steps a workflow run's rows ran, for the checks
// of the workflow's own.
func rowSteps(rows []model.WorkflowRunStep) []model.WorkflowStep {
	steps := make([]model.WorkflowStep, 0, len(rows))
	for _, row := range rows {
		steps = append(steps, model.WorkflowStep{Kind: row.Kind, Folder: row.Folder, Job: row.Job, WorkflowRef: row.WorkflowRef, Prompt: row.Prompt})
	}
	return steps
}
