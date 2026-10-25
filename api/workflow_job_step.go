package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/model"
)

// jobStep turns a requested job step into its row: a job of a folder
// the user sees, with the args it takes; not a shell, nor the bare @host.
func (h *Handlers) jobStep(r *http.Request, step StepRequest) (model.WorkflowStep, error) {
	if step.Folder == "" || step.Job == "" {
		return model.WorkflowStep{}, errors.New("a job step needs a folder and a job")
	}
	if step.Job == model.HostBare {
		return model.WorkflowStep{}, fmt.Errorf("%s / %s: %w", step.Folder, step.Job, folder.ErrHostBare)
	}
	f, job, problem := h.target(r, step.Folder, step.Job)
	if f == nil {
		return model.WorkflowStep{}, h.stepLacks(r, step)
	}
	if problem != "" {
		return model.WorkflowStep{}, fmt.Errorf("%s / %s: %s", step.Folder, step.Job, problem)
	}
	if job.IsShell() {
		return model.WorkflowStep{}, fmt.Errorf("%s / %s: %w", step.Folder, step.Job, folder.ErrShellUnattended)
	}
	if problem := jobArgs(job, step.Args); problem != "" {
		return model.WorkflowStep{}, fmt.Errorf("%s / %s: %s", step.Folder, step.Job, problem)
	}
	return model.WorkflowStep{Kind: model.StepJob, Folder: f.ID, Job: job.ID, Args: model.EncodeList(step.Args)}, nil
}

// stepLacks is the error of a job step whose folder the user has not
// got: the step by its label, the folder relative to its workspace as
// the editor shows it, and which workspace lacks the folder.
func (h *Handlers) stepLacks(r *http.Request, step StepRequest) error {
	ws, title := h.stepSpace(r, step.Folder)
	rel := step.Folder
	if ws != model.Default {
		rel = strings.TrimPrefix(strings.TrimPrefix(step.Folder, ws), "/")
	}
	if rel == "" || rel == model.Default {
		return fmt.Errorf("the root / %s: %s has no jobs at its root", step.Job, title)
	}
	return fmt.Errorf("%s / %s: %s has no folder %s", rel, step.Job, title, rel)
}

// stepSpace returns the workspace of a step's folder, its folder and
// its title: the default one, Default unless titled, for a folder at
// the top of the root.
func (h *Handlers) stepSpace(r *http.Request, folderID string) (ws, title string) {
	ws, title = model.Default, "Default"
	if h.opts.Workspaces == nil {
		return ws, title
	}
	rows, err := h.opts.Workspaces.List(r.Context())
	if err != nil {
		return ws, title
	}
	top, _, _ := strings.Cut(folderID, "/")
	for _, row := range rows {
		switch {
		case row.Folder == top && top != model.Default:
			if row.Title != "" {
				return top, row.Title
			}
			return top, top
		case row.Folder == model.Default && row.Title != "":
			title = row.Title
		}
	}
	return ws, title
}
