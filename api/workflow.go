package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
	"github.com/titpetric/task-ui/workflow"
)

// StepRequest names one step of a workflow: a job of a folder, or a
// workflow to run in its place.
type StepRequest struct {
	Folder string `json:"folder,omitempty"`
	Job    string `json:"job,omitempty"`
	// Workflow is the ID of a workflow the step runs instead of a job.
	Workflow string `json:"workflow,omitempty"`
}

// WorkflowRequest creates or changes a workflow: a name and its jobs, in
// the order they run.
type WorkflowRequest struct {
	Name  string        `json:"name"`
	Steps []StepRequest `json:"steps"`
	// Sidebar lists the workflow in the sidebar, the default; false
	// keeps it out, for one that is a step of others. Left out on a
	// change, it stays as it is.
	Sidebar *bool `json:"sidebar,omitempty"`
}

// WorkflowView is a workflow as the API shows it.
type WorkflowView struct {
	model.Workflow
	Steps []model.WorkflowStep `json:"steps"`
	// Running is the ID of the run in progress, if any.
	Running string `json:"running,omitempty"`
	// LastRun is the newest run.
	LastRun *model.WorkflowRun `json:"last_run,omitempty"`
	// CanRun says whether the user asking may run every step.
	CanRun bool `json:"can_run"`
	// CanChange says whether the user asking may change or delete it.
	CanChange bool `json:"can_change"`
}

// WorkflowRunView is a workflow run with its steps.
type WorkflowRunView struct {
	model.WorkflowRun
	Steps []model.WorkflowRunStep `json:"steps"`
}

// expand replaces the steps that run a workflow with that workflow's
// jobs, through any depth, so a check sees every job the chain runs.
func (h *Handlers) expand(r *http.Request, steps []model.WorkflowStep) []model.WorkflowStep {
	flat, err := h.opts.Workflows.Expand(r.Context(), steps)
	if err != nil {
		return steps
	}
	return flat
}

// mayRun returns the first job the user may not run, or "" when the
// user may run them all, the jobs of the workflows the steps run too.
func (h *Handlers) mayRun(r *http.Request, steps []model.WorkflowStep) string {
	user := access.UserFrom(r.Context())
	for _, step := range h.expand(r, steps) {
		f, job, _ := h.target(r, step.Folder, step.Job)
		if !canRun(user, f, job) {
			return step.Folder + " / " + step.Job
		}
	}
	return ""
}

// visibleSteps reports whether the user may see every step's folder,
// the folders of the workflows the steps run too.
func (h *Handlers) visibleSteps(r *http.Request, steps []model.WorkflowStep) bool {
	user := access.UserFrom(r.Context())
	for _, step := range h.expand(r, steps) {
		if !user.CanSee(step.Folder) {
			return false
		}
	}
	return true
}

// mayChangeWorkflow reports whether the user may change a workflow: its
// owner or an admin.
func mayChangeWorkflow(user *model.User, wf *model.Workflow) bool {
	return user.CanAdmin() || (wf.UserID != "" && user.ID == wf.UserID)
}

// workflowView decorates a workflow for the user asking.
func (h *Handlers) workflowView(r *http.Request, wf model.Workflow, steps []model.WorkflowStep) WorkflowView {
	v := WorkflowView{Workflow: wf, Steps: steps}
	v.Running, _ = h.opts.Flows.Active(wf.ID)
	if runs, err := h.opts.WorkflowRuns.ListByWorkflow(r.Context(), wf.ID, 1); err == nil && len(runs) > 0 {
		v.LastRun = &runs[0]
	}
	v.CanRun = h.mayRun(r, steps) == ""
	v.CanChange = mayChangeWorkflow(access.UserFrom(r.Context()), &wf)
	return v
}

// checkSteps turns the requested steps into rows, each a job the user
// may see, or a workflow whose jobs they all see.
func (h *Handlers) checkSteps(r *http.Request, req WorkflowRequest) ([]model.WorkflowStep, error) {
	if len(req.Steps) == 0 {
		return nil, errors.New("a workflow needs at least one step")
	}
	steps := make([]model.WorkflowStep, 0, len(req.Steps))
	for i, step := range req.Steps {
		if step.Workflow != "" {
			wf, err := h.opts.Workflows.Get(r.Context(), step.Workflow)
			var inner []model.WorkflowStep
			if err == nil {
				inner, err = h.opts.Workflows.Steps(r.Context(), wf.ID)
			}
			if err != nil || !h.visibleSteps(r, inner) {
				return nil, fmt.Errorf("step %d: no workflow %s", i+1, step.Workflow)
			}
			steps = append(steps, model.WorkflowStep{WorkflowRef: wf.ID})
			continue
		}
		f, job, problem := h.target(r, step.Folder, step.Job)
		if problem != "" {
			return nil, fmt.Errorf("step %d, %s / %s: %s", i+1, step.Folder, step.Job, problem)
		}
		steps = append(steps, model.WorkflowStep{Folder: f.ID, Job: job.ID})
	}
	return steps, nil
}

// workflowOf returns the workflow a route names and its steps, when the
// user may see every step.
func (h *Handlers) workflowOf(r *http.Request) (*model.Workflow, []model.WorkflowStep, error) {
	wf, err := h.opts.Workflows.Get(r.Context(), param(r, "id"))
	if err != nil {
		return nil, nil, err
	}
	steps, err := h.opts.Workflows.Steps(r.Context(), wf.ID)
	if err != nil {
		return nil, nil, err
	}
	if !h.visibleSteps(r, steps) {
		return nil, nil, storage.ErrNotFound
	}
	return wf, steps, nil
}

// listWorkflows returns the workflows whose steps the user may see.
func (h *Handlers) listWorkflows(w http.ResponseWriter, r *http.Request) {
	workflows, err := h.opts.Workflows.List(r.Context())
	if err != nil {
		failErr(w, err)
		return
	}
	views := []WorkflowView{}
	for _, wf := range workflows {
		steps, err := h.opts.Workflows.Steps(r.Context(), wf.ID)
		if err != nil {
			failErr(w, err)
			return
		}
		if h.visibleSteps(r, steps) {
			views = append(views, h.workflowView(r, wf, steps))
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"workflows": views})
}

// createWorkflow adds a workflow. Any user may, from the jobs they see.
func (h *Handlers) createWorkflow(w http.ResponseWriter, r *http.Request) {
	var req WorkflowRequest
	if err := readJSON(r, &req); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	steps, err := h.checkSteps(r, req)
	if err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	user := access.UserFrom(r.Context())
	wf := &model.Workflow{Name: req.Name, UserID: user.ID, UserName: user.Name, SidebarHidden: req.Sidebar != nil && !*req.Sidebar}
	if err := h.opts.Workflows.Create(r.Context(), wf, steps); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	steps, _ = h.opts.Workflows.Steps(r.Context(), wf.ID)
	writeJSON(w, http.StatusCreated, h.workflowView(r, *wf, steps))
}

// getWorkflow returns one workflow.
func (h *Handlers) getWorkflow(w http.ResponseWriter, r *http.Request) {
	wf, steps, err := h.workflowOf(r)
	if err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, h.workflowView(r, *wf, steps))
}

// updateWorkflow renames a workflow and replaces its steps.
func (h *Handlers) updateWorkflow(w http.ResponseWriter, r *http.Request) {
	wf, _, err := h.workflowOf(r)
	if err != nil {
		failErr(w, err)
		return
	}
	if !mayChangeWorkflow(access.UserFrom(r.Context()), wf) {
		fail(w, http.StatusForbidden, "not your workflow")
		return
	}
	var req WorkflowRequest
	if err := readJSON(r, &req); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	steps, err := h.checkSteps(r, req)
	if err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	wf.Name = req.Name
	if req.Sidebar != nil {
		wf.SidebarHidden = !*req.Sidebar
	}
	if err := h.opts.Workflows.Update(r.Context(), wf, steps); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	steps, _ = h.opts.Workflows.Steps(r.Context(), wf.ID)
	writeJSON(w, http.StatusOK, h.workflowView(r, *wf, steps))
}

// deleteWorkflow removes a workflow; its runs stay in the history. One
// another workflow runs stays, with 409.
func (h *Handlers) deleteWorkflow(w http.ResponseWriter, r *http.Request) {
	wf, _, err := h.workflowOf(r)
	if err != nil {
		failErr(w, err)
		return
	}
	if !mayChangeWorkflow(access.UserFrom(r.Context()), wf) {
		fail(w, http.StatusForbidden, "not your workflow")
		return
	}
	err = h.opts.Workflows.Delete(r.Context(), wf.ID)
	if errors.Is(err, storage.ErrInUse) {
		fail(w, http.StatusConflict, err.Error())
		return
	}
	if err != nil {
		failErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// startWorkflow runs a workflow, when the user may run every step. A
// workflow already running answers 409 with its run.
func (h *Handlers) startWorkflow(w http.ResponseWriter, r *http.Request) {
	wf, steps, err := h.workflowOf(r)
	if err != nil {
		failErr(w, err)
		return
	}
	if step := h.mayRun(r, steps); step != "" {
		fail(w, http.StatusForbidden, "your role may not run "+step)
		return
	}
	run, err := h.opts.Flows.Start(r.Context(), wf, access.UserFrom(r.Context()))
	if errors.Is(err, workflow.ErrBusy) {
		running, _ := h.opts.Flows.Active(wf.ID)
		writeJSON(w, http.StatusConflict, map[string]any{"error": err.Error(), "run_id": running})
		return
	}
	if err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, run)
}

// listWorkflowRuns returns a workflow's runs, newest first, with their
// steps.
func (h *Handlers) listWorkflowRuns(w http.ResponseWriter, r *http.Request) {
	wf, _, err := h.workflowOf(r)
	if err != nil {
		failErr(w, err)
		return
	}
	runs, err := h.opts.WorkflowRuns.ListByWorkflow(r.Context(), wf.ID, queryInt(r, "limit", 20))
	if err != nil {
		failErr(w, err)
		return
	}
	views := []WorkflowRunView{}
	for _, run := range runs {
		steps, err := h.opts.WorkflowRuns.Steps(r.Context(), run.ID)
		if err != nil {
			failErr(w, err)
			return
		}
		views = append(views, WorkflowRunView{WorkflowRun: run, Steps: steps})
	}
	writeJSON(w, http.StatusOK, map[string]any{"runs": views})
}

// stopWorkflowRun ends a workflow run, for a user who may run its steps.
func (h *Handlers) stopWorkflowRun(w http.ResponseWriter, r *http.Request) {
	run, err := h.opts.WorkflowRuns.Get(r.Context(), param(r, "id"))
	if err != nil {
		failErr(w, err)
		return
	}
	rows, err := h.opts.WorkflowRuns.Steps(r.Context(), run.ID)
	if err != nil {
		failErr(w, err)
		return
	}
	steps := make([]model.WorkflowStep, 0, len(rows))
	for _, row := range rows {
		steps = append(steps, model.WorkflowStep{Folder: row.Folder, Job: row.Job, WorkflowRef: row.WorkflowRef})
	}
	if !h.visibleSteps(r, steps) {
		failErr(w, storage.ErrNotFound)
		return
	}
	if step := h.mayRun(r, steps); step != "" {
		fail(w, http.StatusForbidden, "your role may not run "+step)
		return
	}
	if err := h.opts.Flows.Stop(run.ID); err != nil {
		fail(w, http.StatusConflict, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
