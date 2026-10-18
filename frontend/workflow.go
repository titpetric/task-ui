package frontend

import (
	"context"
	"net/http"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/model"
)

// workflowView decorates a workflow for the page's user. It is nil when
// the user may not see a step's folder.
func (h *Handlers) workflowView(ctx context.Context, p *view.Page, wf model.Workflow) *view.Workflow {
	steps, err := h.opts.Workflows.Steps(ctx, wf.ID)
	if err != nil {
		return nil
	}
	user := access.UserFrom(ctx)
	v := &view.Workflow{Workflow: wf, Steps: steps, CanRun: true}
	for _, step := range steps {
		if !user.CanSee(step.Folder) {
			return nil
		}
		if !p.CanRunJob(step.Folder, step.Job) {
			v.CanRun = false
		}
	}
	v.Running, _ = h.opts.Flows.Active(wf.ID)
	if runs, err := h.opts.WorkflowRuns.ListByWorkflow(ctx, wf.ID, 1); err == nil && len(runs) > 0 {
		v.LastRun = &runs[0]
	}
	v.CanChange = user.CanAdmin() || (wf.UserID != "" && wf.UserID == user.ID)
	return v
}

// visibleWorkflows returns the workflows the page's user may see.
func (h *Handlers) visibleWorkflows(ctx context.Context, p *view.Page) []view.Workflow {
	result := []view.Workflow{}
	if h.opts.Workflows == nil {
		return result
	}
	workflows, _ := h.opts.Workflows.List(ctx)
	for _, wf := range workflows {
		if v := h.workflowView(ctx, p, wf); v != nil {
			result = append(result, *v)
		}
	}
	return result
}

// workflows lists the workflows and takes a new one. Any user may make
// one, from the jobs they can see.
func (h *Handlers) workflows(w http.ResponseWriter, r *http.Request) {
	p := h.page(r, "Workflows")
	p.Crumbs = view.Crumbs("Workflows")
	p.Workflows = p.SideWorkflows
	h.render(w, http.StatusOK, "workflows", p)
}

// workflow shows one workflow: its steps, to run or change, and its runs.
func (h *Handlers) workflow(w http.ResponseWriter, r *http.Request) {
	wf, err := h.opts.Workflows.Get(r.Context(), param(r, "id"))
	if err != nil {
		h.missing(w, r, "No workflow with that ID.")
		return
	}
	p := h.page(r, wf.Name)
	v := h.workflowView(r.Context(), p, *wf)
	if v == nil {
		h.missing(w, r, "No workflow with that ID.")
		return
	}
	p.Workflow = v
	p.Crumbs = view.Crumbs("Workflows", "/workflows", wf.Name)
	runs, _ := h.opts.WorkflowRuns.ListByWorkflow(r.Context(), wf.ID, 20)
	for _, run := range runs {
		steps, _ := h.opts.WorkflowRuns.Steps(r.Context(), run.ID)
		p.WorkflowRuns = append(p.WorkflowRuns, view.WorkflowRun{WorkflowRun: run, Steps: steps})
	}
	h.render(w, http.StatusOK, "workflow", p)
}
