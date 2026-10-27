package api

import (
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
	"github.com/titpetric/task-ui/workflow"
)

// WorkflowRequest creates or changes a workflow: a name, a description
// and its steps, in the order they run. It needs an enabled step.
type WorkflowRequest struct {
	Name        string        `json:"name"`
	Description string        `json:"description,omitempty"`
	Steps       []StepRequest `json:"steps"`
	// Prompt and PromptFolder are the final prompt of a client that
	// sends one: the workflow takes it as its last prompt step, asked in
	// PromptFolder, a folder or folder/dir, the root when empty.
	Prompt       string `json:"prompt,omitempty"`
	PromptFolder string `json:"prompt_folder,omitempty"`
	// Sidebar lists the workflow in the sidebar, the default; false
	// keeps it out, for one that is a step of others. Left out on a
	// change, it stays unchanged.
	Sidebar *bool `json:"sidebar,omitempty"`
	// User is who every run of the workflow runs as, by name: an admin
	// names anyone, everyone else only themselves. Empty, the workflow
	// runs as whoever starts it; left out, no change on update.
	User *string `json:"user,omitempty"`
}

// WorkflowView is a workflow as the API shows it.
type WorkflowView struct {
	model.Workflow
	// Prompt and PromptFolder hide the workflow's columns of the same
	// name, which no workflow uses: its prompts are steps.
	Prompt       string               `json:"prompt,omitempty"`
	PromptFolder string               `json:"prompt_folder,omitempty"`
	Steps        []model.WorkflowStep `json:"steps"`
	// Running is the ID of the run in progress, if any.
	Running string `json:"running,omitempty"`
	// LastRun is the newest run.
	LastRun *model.WorkflowRun `json:"last_run,omitempty"`
	// RunAs names the user the workflow's runs run as; empty for one
	// that runs as whoever starts it, or as a user nobody holds any more.
	RunAs string `json:"run_as"`
	// CanRun says whether a start by the user asking goes: they may run
	// every step, and so may the user the workflow runs as, when it has
	// one.
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

// mayRun returns the first step the user of a request may not run, or
// "" when the user may run them all, the steps of the workflows the
// steps run too, as the runner's MayRun says, which the scheduler asks
// as well.
func (h *Handlers) mayRun(r *http.Request, steps []model.WorkflowStep) string {
	return h.opts.Flows.MayRun(r.Context(), h.permissions(r), steps)
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
	v.RunAs = h.runAsName(r, &wf)
	if user, err := h.runAsUser(r, &wf); err == nil {
		v.CanRun = h.mayRun(r, steps) == "" && h.mayRunAs(r, user, steps) == ""
	}
	v.CanChange = h.mayChangeWorkflow(r, &wf, false)
	return v
}

// checkSteps turns the requested steps into rows, each a job the user
// may see, a workflow whose jobs they all see, or a prompt or summary
// in a place they see, with a final prompt as the last prompt step.
func (h *Handlers) checkSteps(r *http.Request, req WorkflowRequest) ([]model.WorkflowStep, error) {
	requested := req.Steps
	if prompt := strings.TrimSpace(req.Prompt); prompt != "" {
		folder := req.PromptFolder
		if folder == "" {
			// The old field named the root by leaving its folder out.
			folder = model.Default
		}
		requested = append(slices.Clip(requested), StepRequest{Kind: model.StepPrompt, Folder: folder, Prompt: prompt})
	}
	steps := make([]model.WorkflowStep, 0, len(requested))
	for i, step := range requested {
		row, err := h.checkStep(r, step)
		if err != nil {
			return nil, fmt.Errorf("step %d: %w", i+1, err)
		}
		row.Disabled = step.Enabled != nil && !*step.Enabled
		steps = append(steps, row)
	}
	return steps, nil
}

// checkStep turns one requested step into its row, by its kind.
func (h *Handlers) checkStep(r *http.Request, step StepRequest) (model.WorkflowStep, error) {
	switch step.StepKind() {
	case model.StepPrompt, model.StepSummary:
		return h.promptStep(r, step)
	case model.StepWorkflow:
		wf, err := h.opts.Workflows.Get(r.Context(), step.Workflow)
		var inner []model.WorkflowStep
		if err == nil {
			inner, err = h.opts.Workflows.Steps(r.Context(), wf.ID)
		}
		if err != nil || !h.visibleSteps(r, inner) {
			return model.WorkflowStep{}, fmt.Errorf("no workflow %s", step.Workflow)
		}
		return model.WorkflowStep{Kind: model.StepWorkflow, WorkflowRef: wf.ID}, nil
	case model.StepJob:
		return h.jobStep(r, step)
	}
	return model.WorkflowStep{}, fmt.Errorf("no step kind %q: job, workflow, prompt or summary", step.Kind)
}

// checkWorkflow reads a request into the workflow and its steps, those
// checkSteps takes; a prompt or summary step takes prompt/edit. It
// answers the error and returns false for a request it refuses, and for
// one with no enabled step.
func (h *Handlers) checkWorkflow(w http.ResponseWriter, r *http.Request, req WorkflowRequest, wf *model.Workflow) ([]model.WorkflowStep, bool) {
	asks := strings.TrimSpace(req.Prompt) != ""
	for _, step := range req.Steps {
		asks = asks || step.StepKind() == model.StepPrompt || step.StepKind() == model.StepSummary
	}
	if asks && !h.can(r, "prompt/edit") {
		fail(w, http.StatusForbidden, "you don't have permission to prompt Claude")
		return nil, false
	}
	steps, err := h.checkSteps(r, req)
	if err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return nil, false
	}
	wf.Name, wf.Description = strings.TrimSpace(req.Name), strings.TrimSpace(req.Description)
	if !wf.Runs(steps) {
		fail(w, http.StatusBadRequest, "a workflow needs an enabled step")
		return nil, false
	}
	if req.User != nil {
		runAs, status, problem := h.pickRunAs(r, *req.User)
		if problem != "" {
			fail(w, status, problem)
			return nil, false
		}
		wf.RunAsUserID = ""
		if runAs != nil {
			wf.RunAsUserID = runAs.ID
		}
	}
	return steps, true
}

// workflowOf returns the workflow a route names and its steps, when the
// user may see every step.
func (h *Handlers) workflowOf(r *http.Request) (*model.Workflow, []model.WorkflowStep, error) {
	return h.workflowByID(r, param(r, "id"))
}

// workflowByID returns a workflow by ID and its steps, when the user may
// see every step; storage.ErrNotFound otherwise.
func (h *Handlers) workflowByID(r *http.Request, id string) (*model.Workflow, []model.WorkflowStep, error) {
	wf, err := h.opts.Workflows.Get(r.Context(), id)
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

// createWorkflow adds a workflow. Any user may, from the jobs they see;
// its runs run as the caller unless an admin names another user.
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

// updateWorkflow renames a workflow, replaces its steps and, with a
// user named, changes who its runs run as.
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

// startWorkflow runs a workflow, when the caller may run every step,
// its prompts with prompt/edit, as the user the workflow runs as, who
// must be able to run them too, a 403 naming the step and the user
// otherwise and a 409 for a user nobody holds; an optional body's input
// goes to the prompts. A workflow already running answers 409 with its
// run.
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
	user, err := h.runAsUser(r, wf)
	if err != nil {
		fail(w, http.StatusConflict, "the user this workflow runs as is gone; change it on the workflow")
		return
	}
	if step := h.mayRunAs(r, user, steps); step != "" {
		fail(w, http.StatusForbidden, user.Name+" may not run "+step)
		return
	}
	run, err := h.opts.Flows.Start(r.Context(), wf, user, req.Input)
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
	return h.visibleSteps(r, rowSteps(rows))
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
	steps := rowSteps(rows)
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
