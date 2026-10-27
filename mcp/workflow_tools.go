package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/model"
)

// workflowTools read, make and change workflows as the token's user,
// through /api/workflows, so an agent builds a workflow and tests it
// with run_workflow, not only runs one (docs/api.md). There is no
// delete.
func workflowTools() []Tool {
	str := func(text string) Property { return Property{Type: "string", Description: text} }
	step := Schema{Type: "object", Properties: map[string]Property{
		"kind": {Type: "string", Enum: []string{model.StepJob, model.StepWorkflow, model.StepPrompt, model.StepSummary}, Description: "What the step does: job runs a job of a folder, workflow runs another workflow, " +
			"prompt asks Claude its prompt, summary asks Claude for task-ui's summary of the steps before it, with no prompt. Left out, prompt with a prompt, workflow with a workflow, else job."},
		"folder":   str("The folder of a job step, or the folder a prompt or summary step asks in, _ for the root."),
		"job":      str("The job of a job step, as list_jobs names it."),
		"workflow": str("The ID or name of a workflow a workflow step runs."),
		"prompt":   str("What a prompt step asks Claude, with the output of the steps before it; its answer is its output."),
		"dir":      str("A directory inside the folder a prompt or summary step asks in."),
		"args":     {Type: "array", Description: "Args for a job that takes them.", Items: &Schema{Type: "string"}},
		"enabled":  {Type: "boolean", Description: "False keeps the step and a run skips it; true unless given."},
	}}
	fields := func() map[string]Property {
		return map[string]Property{
			"name":        str("The workflow's name."),
			"description": str("What the workflow is for, a line or two."),
			"steps": {Type: "array", Description: "The steps in the order they run, each a job, a workflow, a prompt or a summary. " +
				"The last prompt or summary step's answer is the run's answer.", Items: &step},
			"sidebar": {Type: "boolean", Description: "List the workflow in the sidebar; true unless given."},
		}
	}
	update := fields()
	update["workflow"] = str("The workflow's ID or name.")
	return []Tool{
		{
			Name:        "get_workflow",
			Description: "Show a workflow by ID or name: its steps in order, each a job, a workflow, a prompt or a summary, and its last run.",
			InputSchema: object(map[string]Property{"workflow": str("The workflow's ID or name.")}, "workflow"),
			call:        getWorkflow,
		},
		{
			Name: "create_workflow",
			Description: "Make a workflow as you: a name and its steps in order, each a job of a folder, another workflow, a prompt to Claude " +
				"or a summary of the steps before it. Prompts and summaries take prompt/edit. Test it with run_workflow.",
			InputSchema: object(fields(), "name"),
			call:        createWorkflow,
		},
		{
			Name: "update_workflow",
			Description: "Change a workflow you own, or any as an admin: a field left out stays as it is, steps given replace them all. " +
				"There is no delete.",
			InputSchema: object(update, "workflow"),
			call:        updateWorkflow,
		},
	}
}

// stepArgs is a step as a tool takes it.
type stepArgs struct {
	Kind     string   `json:"kind"`
	Folder   string   `json:"folder"`
	Job      string   `json:"job"`
	Workflow string   `json:"workflow"`
	Prompt   string   `json:"prompt"`
	Dir      string   `json:"dir"`
	Args     []string `json:"args"`
	Enabled  *bool    `json:"enabled"`
}

// workflowFields are the fields create_workflow and update_workflow
// take; a nil one keeps the workflow's own on an update.
type workflowFields struct {
	Workflow    string      `json:"workflow"`
	Name        *string     `json:"name"`
	Description *string     `json:"description"`
	Steps       *[]stepArgs `json:"steps"`
	Sidebar     *bool       `json:"sidebar"`
}

func getWorkflow(ctx context.Context, c *client, args json.RawMessage) (string, error) {
	var in workflowFields
	if err := decode(args, &in); err != nil {
		return "", err
	}
	wf, names, err := findWorkflow(ctx, c, in.Workflow)
	if err != nil {
		return "", err
	}
	return showWorkflow(*wf, names), nil
}

func createWorkflow(ctx context.Context, c *client, args json.RawMessage) (string, error) {
	var in workflowFields
	if err := decode(args, &in); err != nil {
		return "", err
	}
	if in.Name == nil || strings.TrimSpace(*in.Name) == "" {
		return "", errors.New("a workflow needs a name")
	}
	req := api.WorkflowRequest{Name: *in.Name, Sidebar: in.Sidebar}
	return saveWorkflow(ctx, c, http.MethodPost, "/workflows", in, req)
}

func updateWorkflow(ctx context.Context, c *client, args json.RawMessage) (string, error) {
	var in workflowFields
	if err := decode(args, &in); err != nil {
		return "", err
	}
	wf, _, err := findWorkflow(ctx, c, in.Workflow)
	if err != nil {
		return "", err
	}
	req := api.WorkflowRequest{Name: wf.Name, Description: wf.Description, Sidebar: in.Sidebar}
	if in.Name != nil {
		req.Name = *in.Name
	}
	req.Steps = make([]api.StepRequest, 0, len(wf.Steps))
	for _, step := range wf.Steps {
		req.Steps = append(req.Steps, keptStep(step))
	}
	return saveWorkflow(ctx, c, http.MethodPut, "/workflows/"+url.PathEscape(wf.ID), in, req)
}

// keptStep is a workflow's step as a request carries it unchanged.
func keptStep(step model.WorkflowStep) api.StepRequest {
	enabled := !step.Disabled
	out := api.StepRequest{Kind: step.StepKind(), Enabled: &enabled}
	switch {
	case step.IsPrompt():
		out.Prompt, out.Folder, out.Dir = step.Prompt, step.Folder, step.Dir
	case step.IsWorkflow():
		out.Workflow = step.WorkflowRef
	default:
		out.Folder, out.Job, out.Args = step.Folder, step.Job, model.DecodeList(step.Args)
	}
	return out
}

// saveWorkflow fills the request with the fields given, the steps'
// workflows by ID, sends it and shows the workflow saved.
func saveWorkflow(ctx context.Context, c *client, method, path string, in workflowFields, req api.WorkflowRequest) (string, error) {
	if in.Description != nil {
		req.Description = *in.Description
	}
	if in.Steps != nil {
		req.Steps = make([]api.StepRequest, 0, len(*in.Steps))
		for i, step := range *in.Steps {
			out := api.StepRequest{Kind: step.Kind, Folder: step.Folder, Job: step.Job, Prompt: step.Prompt, Dir: step.Dir, Args: step.Args, Enabled: step.Enabled}
			if step.Workflow != "" {
				ref, _, err := findWorkflow(ctx, c, step.Workflow)
				if err != nil {
					return "", fmt.Errorf("step %d: %w", i+1, err)
				}
				out.Workflow = ref.ID
			}
			req.Steps = append(req.Steps, out)
		}
	}
	data, err := c.call(ctx, method, path, req)
	if err != nil {
		return "", err
	}
	var view api.WorkflowView
	if err := json.Unmarshal(data, &view); err != nil {
		return "", err
	}
	_, names, err := findWorkflow(ctx, c, view.ID)
	if err != nil {
		return "", err
	}
	return showWorkflow(view, names), nil
}

// showWorkflow describes a workflow: its description, its steps in
// order and its last run.
func showWorkflow(wf api.WorkflowView, names map[string]string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "workflow %s (id %s) by %s", wf.Name, wf.ID, wf.UserName)
	if wf.SidebarHidden {
		b.WriteString(", not in the sidebar")
	}
	if !wf.CanRun {
		b.WriteString(" [you may not run it]")
	}
	if !wf.CanChange {
		b.WriteString(" [you may not change it]")
	}
	b.WriteString("\n")
	if wf.Description != "" {
		b.WriteString(wf.Description + "\n")
	}
	if len(wf.Steps) == 0 {
		b.WriteString("no steps\n")
	}
	for i, step := range wf.Steps {
		fmt.Fprintf(&b, "  %d. ", i+1)
		switch {
		case step.IsSummary():
			fmt.Fprintf(&b, "summary in %s", model.PlaceLabel(step.Place()))
		case step.IsPrompt():
			fmt.Fprintf(&b, "prompt in %s: %s", model.PlaceLabel(step.Place()), step.Prompt)
		case step.IsWorkflow():
			fmt.Fprintf(&b, "workflow %s (id %s)", names[step.WorkflowRef], step.WorkflowRef)
		default:
			b.WriteString(step.Folder + "/" + step.Job)
			if args := model.DecodeList(step.Args); len(args) > 0 {
				b.WriteString(" -- " + strings.Join(args, " "))
			}
		}
		if step.Disabled {
			b.WriteString(" [off]")
		}
		b.WriteString("\n")
	}
	if wf.LastRun != nil {
		fmt.Fprintf(&b, "last run %s: %s\n", wf.LastRun.ID, flowOutcome(wf.LastRun))
	}
	return b.String()
}
