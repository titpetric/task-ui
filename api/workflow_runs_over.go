package api

import (
	"fmt"
	"net/http"
	"slices"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/model"
)

// runsOver checks the workspace a workflow runs over: one the user sees,
// with steps that run in its projects, which no workflow step does.
func (h *Handlers) runsOver(r *http.Request, workspace string, steps []model.WorkflowStep) error {
	if !h.isWorkspace(r, workspace) || !access.UserFrom(r.Context()).CanSee(workspace) {
		return fmt.Errorf("no workspace %s to run over", workspace)
	}
	if slices.ContainsFunc(steps, func(s model.WorkflowStep) bool { return s.IsWorkflow() }) {
		return fmt.Errorf("a workflow that runs over %s runs its steps in each project, and a workflow step has no project", workspace)
	}
	return nil
}

// isWorkspace reports whether a folder ID names a workspace: the default one, or a workspace row's.
func (h *Handlers) isWorkspace(r *http.Request, id string) bool {
	if id == model.Default {
		return true
	}
	if h.opts.Workspaces == nil {
		return false
	}
	rows, err := h.opts.Workspaces.List(r.Context())
	if err != nil {
		return false
	}
	return slices.ContainsFunc(rows, func(row model.Workspace) bool { return row.Folder == id })
}
