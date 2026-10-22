package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
	"github.com/titpetric/task-ui/workflow"
)

// WorkflowRequest creates or changes a workflow: a name and its jobs, in
// the order they run, and a prompt to ask Claude once they end. It needs
// an enabled step or a prompt.
type WorkflowRequest struct {
	Name  string        `json:"name"`
	Steps []StepRequest `json:"steps"`
	// Prompt is what Claude gets once the workflow's steps end, with
	// their output, as the user who runs it; it takes prompt/edit.
	Prompt string `json:"prompt,omitempty"`
	// PromptFolder is where the prompt runs: a folder, or folder/dir,
	// the user sees; empty runs it in the root, _.
	PromptFolder string `json:"prompt_folder,omitempty"`
	// Sidebar lists the workflow in the sidebar, the default; false
	// keeps it out, for one that is a step of others. Left out on a
	// change, it stays unchanged.
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
	for _, step := range h.expand(r, steps) {
		f, job, _ := h.target(r, step.Folder, step.Job)
		if !h.canRun(r, f, job) {
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

// mayChangeWorkflow reports whether the user of a request may change a
// workflow, with workflow/edit, or delete it, with workflow/delete: its
// owner, or an admin.
func (h *Handlers) mayChangeWorkflow(r *http.Request, wf *model.Workflow, del bool) bool {
	allowed := h.can(r, "workflow/edit")
	if del {
		allowed = h.can(r, "workflow/delete")
	}
	user := access.UserFrom(r.Context())
	return allowed && (user.CanAdmin() || (wf.UserID != "" && user.ID == wf.UserID))
}

// workflowView decorates a workflow for the user asking.
func (h *Handlers) workflowView(r *http.Request, wf model.Workflow, steps []model.WorkflowStep) WorkflowView {
	v := WorkflowView{Workflow: wf, Steps: steps}
	v.Running, _ = h.opts.Flows.Active(wf.ID)
	if runs, err := h.opts.WorkflowRuns.ListByWorkflow(r.Context(), wf.ID, storage.WorkflowRunFilter{Limit: 1}); err == nil && len(runs) > 0 {
		v.LastRun = &runs[0]
	}
	v.CanRun = h.mayRun(r, steps) == ""
	v.CanChange = h.mayChangeWorkflow(r, &wf, false)
	return v
}

// checkSteps turns the requested steps into rows, each a job the user
// may see, or a workflow whose jobs they all see.
func (h *Handlers) checkSteps(r *http.Request, req WorkflowRequest) ([]model.WorkflowStep, error) {
	steps := make([]model.WorkflowStep, 0, len(req.Steps))
	for i, step := range req.Steps {
		disabled := step.Enabled != nil && !*step.Enabled
		if step.Workflow != "" {
			wf, err := h.opts.Workflows.Get(r.Context(), step.Workflow)
			var inner []model.WorkflowStep
			if err == nil {
				inner, err = h.opts.Workflows.Steps(r.Context(), wf.ID)
			}
			if err != nil || !h.visibleSteps(r, inner) {
				return nil, fmt.Errorf("step %d: no workflow %s", i+1, step.Workflow)
			}
			steps = append(steps, model.WorkflowStep{WorkflowRef: wf.ID, Disabled: disabled})
			continue
		}
		f, job, problem := h.target(r, step.Folder, step.Job)
		if problem != "" {
			return nil, fmt.Errorf("step %d, %s / %s: %s", i+1, step.Folder, step.Job, problem)
		}
		if job.IsShell() {
			return nil, fmt.Errorf("step %d, %s / %s: %w", i+1, step.Folder, step.Job, folder.ErrShellUnattended)
		}
		steps = append(steps, model.WorkflowStep{Folder: f.ID, Job: job.ID, Disabled: disabled})
	}
	return steps, nil
}

// checkWorkflow reads a request into the workflow and its steps: the
// steps checkSteps takes, and the prompt, which takes prompt/edit and a
// place the user sees. It answers the error and returns false for a
// request it refuses, and for one with no enabled step and no prompt.
func (h *Handlers) checkWorkflow(w http.ResponseWriter, r *http.Request, req WorkflowRequest, wf *model.Workflow) ([]model.WorkflowStep, bool) {
	steps, err := h.checkSteps(r, req)
	if err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return nil, false
	}
	wf.Name = req.Name
	wf.Prompt, wf.PromptFolder = strings.TrimSpace(req.Prompt), strings.Trim(req.PromptFolder, "/")
	if wf.PromptFolder == model.Default {
		wf.PromptFolder = ""
	}
	if !wf.Runs(steps) {
		fail(w, http.StatusBadRequest, "a workflow needs an enabled step or a prompt")
		return nil, false
	}
	if !wf.HasPrompt() {
		wf.PromptFolder = ""
		return steps, true
	}
	if !h.can(r, "prompt/edit") {
		fail(w, http.StatusForbidden, "you don't have permission to prompt Claude")
		return nil, false
	}
	if h.opts.Board == nil {
		fail(w, http.StatusBadRequest, "prompts aren't available on this instance")
		return nil, false
	}
	if err := h.opts.Board.PromptPlace(r.Context(), access.UserFrom(r.Context()), wf.PromptPlace()); err != nil {
		fail(w, http.StatusBadRequest, "can't run the prompt there: "+err.Error())
		return nil, false
	}
	return steps, true
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
	if !h.require(w, r, "workflow/view") {
		return
	}
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
	if !h.require(w, r, "workflow/edit") {
		return
	}
	var req WorkflowRequest
	if err := readJSON(r, &req); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	user := access.UserFrom(r.Context())
	wf := &model.Workflow{UserID: user.ID, UserName: user.Name, SidebarHidden: req.Sidebar != nil && !*req.Sidebar}
	steps, ok := h.checkWorkflow(w, r, req, wf)
	if !ok {
		return
	}
	if err := h.opts.Workflows.Create(r.Context(), wf, steps); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	steps, _ = h.opts.Workflows.Steps(r.Context(), wf.ID)
	writeJSON(w, http.StatusCreated, h.workflowView(r, *wf, steps))
}

// getWorkflow returns one workflow.
func (h *Handlers) getWorkflow(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "workflow/view") {
		return
	}
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
	if !h.mayChangeWorkflow(r, wf, false) {
		fail(w, http.StatusForbidden, "only its owner or an admin can change this workflow")
		return
	}
	var req WorkflowRequest
	if err := readJSON(r, &req); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	steps, ok := h.checkWorkflow(w, r, req, wf)
	if !ok {
		return
	}
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
	if !h.mayChangeWorkflow(r, wf, true) {
		fail(w, http.StatusForbidden, "only its owner or an admin can change this workflow")
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

// startWorkflow runs a workflow, when the user may run every step, and
// ask its prompt, with prompt/edit; an optional body's input goes to the
// prompt. A workflow already running answers 409 with its run.
func (h *Handlers) startWorkflow(w http.ResponseWriter, r *http.Request) {
	wf, steps, err := h.workflowOf(r)
	if err != nil {
		failErr(w, err)
		return
	}
	var req WorkflowRunRequest
	if err := readJSON(r, &req); err != nil && r.ContentLength > 0 {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	if step := h.mayRun(r, steps); step != "" {
		fail(w, http.StatusForbidden, "you don't have permission to run "+step)
		return
	}
	if wf.HasPrompt() && !h.can(r, "prompt/edit") {
		fail(w, http.StatusForbidden, "the workflow ends with a prompt, and you don't have permission to prompt Claude")
		return
	}
	run, err := h.opts.Flows.Start(r.Context(), wf, access.UserFrom(r.Context()), req.Input)
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
	if !h.require(w, r, "workflow/view") {
		return
	}
	wf, _, err := h.workflowOf(r)
	if err != nil {
		failErr(w, err)
		return
	}
	runs, err := h.opts.WorkflowRuns.ListByWorkflow(r.Context(), wf.ID, storage.WorkflowRunFilter{Limit: queryInt(r, "limit", 20)})
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

// listRecentWorkflowRuns returns the runs of every workflow, newest
// first, those of the workflows the user sees every folder of.
func (h *Handlers) listRecentWorkflowRuns(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "workflow/view") {
		return
	}
	runs, err := h.opts.WorkflowRuns.List(r.Context(), queryInt(r, "limit", 10))
	if err != nil {
		failErr(w, err)
		return
	}
	views := []model.WorkflowRun{}
	for _, run := range runs {
		if h.visibleRun(r, run.ID) {
			views = append(views, run)
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"runs": views})
}

// visibleRun reports whether the user may see a workflow run: every
// folder its steps ran in, the workflows they ran included.
func (h *Handlers) visibleRun(r *http.Request, runID string) bool {
	rows, err := h.opts.WorkflowRuns.Steps(r.Context(), runID)
	if err != nil {
		return false
	}
	steps := make([]model.WorkflowStep, 0, len(rows))
	for _, row := range rows {
		steps = append(steps, model.WorkflowStep{Folder: row.Folder, Job: row.Job, WorkflowRef: row.WorkflowRef})
	}
	return h.visibleSteps(r, steps)
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
		fail(w, http.StatusForbidden, "you don't have permission to run "+step)
		return
	}
	if err := h.opts.Flows.Stop(run.ID); err != nil {
		fail(w, http.StatusConflict, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
