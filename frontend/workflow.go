package frontend

import (
	"context"
	"net/http"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/model"
)

// workflowNames returns the name of every workflow by ID.
func (h *Handlers) workflowNames(ctx context.Context) map[string]string {
	names := map[string]string{}
	workflows, _ := h.opts.Workflows.List(ctx)
	for _, wf := range workflows {
		names[wf.ID] = wf.Name
	}
	return names
}

// viewSteps turns a workflow's steps into the page's.
func viewSteps(steps []model.WorkflowStep, names map[string]string) []view.Step {
	out := make([]view.Step, 0, len(steps))
	for _, s := range steps {
		out = append(out, view.Step{Folder: s.Folder, Job: s.Job, WorkflowID: s.WorkflowRef, WorkflowName: names[s.WorkflowRef]})
	}
	return out
}

// viewRunSteps turns a workflow run's steps into the page's.
func viewRunSteps(rows []model.WorkflowRunStep, names map[string]string) []view.Step {
	out := make([]view.Step, 0, len(rows))
	for _, s := range rows {
		out = append(out, view.Step{
			Folder: s.Folder, Job: s.Job, WorkflowID: s.WorkflowRef, WorkflowName: names[s.WorkflowRef],
			State: s.State, RunID: s.RunID, WorkflowRunID: s.WorkflowRunRef,
		})
	}
	return out
}

// workflowView decorates a workflow for the page's user. It is nil when
// the user may not see a step's folder, the folders of the workflows
// the steps run among them.
func (h *Handlers) workflowView(ctx context.Context, p *view.Page, wf model.Workflow, names map[string]string) *view.Workflow {
	steps, err := h.opts.Workflows.Steps(ctx, wf.ID)
	if err != nil {
		return nil
	}
	jobs, err := h.opts.Workflows.Expand(ctx, steps)
	if err != nil {
		return nil
	}
	user := access.UserFrom(ctx)
	v := &view.Workflow{Workflow: wf, Steps: viewSteps(steps, names), CanRun: true}
	for _, step := range jobs {
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
	names := map[string]string{}
	for _, wf := range workflows {
		names[wf.ID] = wf.Name
	}
	for _, wf := range workflows {
		if v := h.workflowView(ctx, p, wf, names); v != nil {
			result = append(result, *v)
		}
	}
	return result
}

// stepWorkflows returns the workflows a step of the one edited may run:
// the visible ones, less itself and those that run it, a loop.
func (h *Handlers) stepWorkflows(ctx context.Context, p *view.Page, self string) []view.Workflow {
	result := []view.Workflow{}
	for _, wf := range p.SideWorkflows {
		if loops, _ := h.opts.Workflows.Reaches(ctx, wf.ID, self); loops {
			continue
		}
		result = append(result, wf)
	}
	return result
}

// workflows lists the workflows and takes a new one. Any user may make
// one, from the jobs they can see.
func (h *Handlers) workflows(w http.ResponseWriter, r *http.Request) {
	p := h.page(r, "Workflows")
	p.Crumbs = view.Crumbs("Workflows")
	p.Workflows = p.SideWorkflows
	p.StepWorkflows = p.SideWorkflows
	h.render(w, http.StatusOK, "workflows", p)
}

// workflowNew shows the form for a new workflow. Any user may make one,
// from the jobs and workflows they see.
func (h *Handlers) workflowNew(w http.ResponseWriter, r *http.Request) {
	p := h.page(r, "New workflow")
	p.Crumbs = view.Crumbs("Workflows", "/workflows", "New")
	p.StepWorkflows = p.SideWorkflows
	h.render(w, http.StatusOK, "workflow_new", p)
}

// loadWorkflow fills the page with the workflow a route names, or
// renders the not found page and returns nil.
func (h *Handlers) loadWorkflow(w http.ResponseWriter, r *http.Request) (*view.Page, map[string]string) {
	wf, err := h.opts.Workflows.Get(r.Context(), param(r, "id"))
	if err != nil {
		h.missing(w, r, "No workflow with that ID.")
		return nil, nil
	}
	p := h.page(r, wf.Name)
	names := h.workflowNames(r.Context())
	v := h.workflowView(r.Context(), p, *wf, names)
	if v == nil {
		h.missing(w, r, "No workflow with that ID.")
		return nil, nil
	}
	p.Workflow = v
	p.Crumbs = view.Crumbs("Workflows", "/workflows", wf.Name)
	return p, names
}

// workflow shows one workflow: its pipeline with the latest run's
// states, and its runs. The page follows the workflow's events.
func (h *Handlers) workflow(w http.ResponseWriter, r *http.Request) {
	p, names := h.loadWorkflow(w, r)
	if p == nil {
		return
	}
	runs, _ := h.opts.WorkflowRuns.ListByWorkflow(r.Context(), p.Workflow.ID, 20)
	for _, run := range runs {
		rows, _ := h.opts.WorkflowRuns.Steps(r.Context(), run.ID)
		p.WorkflowRuns = append(p.WorkflowRuns, view.WorkflowRun{WorkflowRun: run, Steps: viewRunSteps(rows, names)})
	}
	var latest []view.Step
	if len(p.WorkflowRuns) > 0 {
		latest = p.WorkflowRuns[0].Steps
	}
	p.Pipeline = view.Pipeline(p.Workflow.Steps, latest)
	h.frames(r.Context(), p.Pipeline, names, true)
	h.render(w, http.StatusOK, "workflow", p)
}

// frames puts on each step of a run's chain the screen its run ended
// on, and under a step that ran a workflow the chain of that run, one
// level down.
func (h *Handlers) frames(ctx context.Context, links []view.Link, names map[string]string, deeper bool) {
	for i := range links {
		if links[i].RunID != "" {
			links[i].FrameSVG, links[i].Lines, links[i].Complete = h.frameOf(ctx, links[i].RunID)
		}
		if deeper && links[i].WorkflowRunID != "" {
			rows, err := h.opts.WorkflowRuns.Steps(ctx, links[i].WorkflowRunID)
			if err != nil {
				continue
			}
			links[i].Sub = view.Chain(viewRunSteps(rows, names))
			h.frames(ctx, links[i].Sub, names, false)
		}
	}
}

// workflowEdit shows the form that changes a workflow, for its owner or
// an admin.
func (h *Handlers) workflowEdit(w http.ResponseWriter, r *http.Request) {
	p, _ := h.loadWorkflow(w, r)
	if p == nil {
		return
	}
	if !p.Workflow.CanChange {
		p.Title = "Not yours"
		p.Error = "Its owner or an admin changes a workflow."
		h.render(w, http.StatusForbidden, "notfound", p)
		return
	}
	p.Title = "Edit " + p.Workflow.Name
	p.Crumbs = view.Crumbs("Workflows", "/workflows", p.Workflow.Name, "/workflows/"+p.Workflow.ID, "Edit")
	p.StepWorkflows = h.stepWorkflows(r.Context(), p, p.Workflow.ID)
	h.render(w, http.StatusOK, "workflow_edit", p)
}
