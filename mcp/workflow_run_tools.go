package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/model"
)

type workflowList struct {
	Workflows []api.WorkflowView `json:"workflows"`
}

func listWorkflows(ctx context.Context, c *client, args json.RawMessage) (string, error) {
	var out workflowList
	if err := c.get(ctx, "/workflows", &out); err != nil {
		return "", err
	}
	if len(out.Workflows) == 0 {
		return "No workflows.", nil
	}
	names := map[string]string{}
	for _, wf := range out.Workflows {
		names[wf.ID] = wf.Name
	}
	var b strings.Builder
	for _, wf := range out.Workflows {
		steps := make([]string, 0, len(wf.Steps))
		for _, step := range wf.Steps {
			if step.WorkflowRef != "" {
				steps = append(steps, "workflow "+names[step.WorkflowRef])
				continue
			}
			steps = append(steps, step.Folder+"/"+step.Job)
		}
		fmt.Fprintf(&b, "%s (id %s): %s", wf.Name, wf.ID, strings.Join(steps, " -> "))
		if !wf.CanRun {
			b.WriteString(" [view only]")
		}
		if wf.LastRun != nil {
			fmt.Fprintf(&b, " | last: %s", flowOutcome(wf.LastRun))
		}
		b.WriteString("\n")
	}
	return b.String(), nil
}

type workflowArgs struct {
	Workflow string `json:"workflow"`
	Input    string `json:"input"`
	Wait     int    `json:"wait"`
	Limit    int    `json:"limit"`
}

// busyFlow is the API's answer to a workflow that is running already.
type busyFlow struct {
	RunID string `json:"run_id"`
}

type flowRunList struct {
	Runs []api.WorkflowRunView `json:"runs"`
}

func runWorkflow(ctx context.Context, c *client, args json.RawMessage) (string, error) {
	var in workflowArgs
	if err := decode(args, &in); err != nil {
		return "", err
	}
	wf, names, err := findWorkflow(ctx, c, in.Workflow)
	if err != nil {
		return "", err
	}
	var body any
	if in.Input != "" {
		body = map[string]string{"input": in.Input}
	}
	data, err := c.call(ctx, http.MethodPost, "/workflows/"+url.PathEscape(wf.ID)+"/runs", body)
	var e *apiError
	if errors.As(err, &e) && e.Status == http.StatusConflict {
		var busy busyFlow
		if json.Unmarshal(e.Body, &busy) == nil && busy.RunID != "" {
			return "", fmt.Errorf("%s is already running as workflow run %s", wf.Name, busy.RunID)
		}
	}
	if err != nil {
		return "", err
	}
	var run model.WorkflowRun
	if err := json.Unmarshal(data, &run); err != nil {
		return "", err
	}
	path := "/workflows/" + url.PathEscape(wf.ID) + "/runs?limit=20"
	deadline := time.Now().Add(min(time.Duration(in.Wait)*time.Second, maxWait))
	for {
		var out flowRunList
		if err := c.get(ctx, path, &out); err != nil {
			return "", err
		}
		for _, view := range out.Runs {
			if view.ID == run.ID && (model.Finished(view.State) || !time.Now().Before(deadline)) {
				return showFlowRun(view, names), nil
			}
		}
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(pollEvery):
		}
	}
}

func listWorkflowRuns(ctx context.Context, c *client, args json.RawMessage) (string, error) {
	in := workflowArgs{Limit: 5}
	if err := decode(args, &in); err != nil {
		return "", err
	}
	wf, names, err := findWorkflow(ctx, c, in.Workflow)
	if err != nil {
		return "", err
	}
	var out flowRunList
	if err := c.get(ctx, fmt.Sprintf("/workflows/%s/runs?limit=%d", url.PathEscape(wf.ID), in.Limit), &out); err != nil {
		return "", err
	}
	if len(out.Runs) == 0 {
		return wf.Name + " has not run.", nil
	}
	shown := make([]string, 0, len(out.Runs))
	for _, view := range out.Runs {
		shown = append(shown, showFlowRun(view, names))
	}
	return strings.Join(shown, "\n"), nil
}

// findWorkflow returns the workflow with an ID or a name, and the names
// of every workflow by ID, for the steps that run another.
func findWorkflow(ctx context.Context, c *client, ref string) (*api.WorkflowView, map[string]string, error) {
	if ref == "" {
		return nil, nil, errors.New("a workflow ID or name is needed")
	}
	var out workflowList
	if err := c.get(ctx, "/workflows", &out); err != nil {
		return nil, nil, err
	}
	names := map[string]string{}
	var found *api.WorkflowView
	for i, wf := range out.Workflows {
		names[wf.ID] = wf.Name
		if found == nil && (wf.ID == ref || strings.EqualFold(wf.Name, ref)) {
			found = &out.Workflows[i]
		}
	}
	if found == nil {
		return nil, nil, fmt.Errorf("no workflow %s; list_workflows lists them", ref)
	}
	return found, names, nil
}

// showFlowRun describes a workflow run, its steps, its input and its answer.
func showFlowRun(view api.WorkflowRunView, names map[string]string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "workflow run %s: %s by %s, %s\n", view.ID, view.Name, view.UserName, flowOutcome(&view.WorkflowRun))
	for i, step := range view.Steps {
		name := step.Folder + "/" + step.Job
		place := model.PlaceLabel(step.Folder)
		switch {
		case step.IsSummary():
			name = "summary in " + place
		case step.IsPrompt():
			name = "prompt in " + place
		case step.WorkflowRef != "":
			name = "workflow " + names[step.WorkflowRef]
		}
		state := step.State
		if state == "" {
			state = "not run"
		}
		fmt.Fprintf(&b, "  %d. %s: %s", i+1, name, state)
		if step.RunID != "" {
			fmt.Fprintf(&b, " (run %s)", step.RunID)
		}
		if step.WorkflowRunRef != "" {
			fmt.Fprintf(&b, " (workflow run %s)", step.WorkflowRunRef)
		}
		b.WriteString("\n")
	}
	if view.Input != "" {
		b.WriteString("input:\n" + view.Input + "\n")
	}
	if view.Answer != "" {
		b.WriteString("answer:\n" + view.Answer + "\n")
	}
	return b.String()
}

// flowOutcome says how a workflow run went.
func flowOutcome(run *model.WorkflowRun) string {
	text := outcome(&model.Run{State: run.State, StartedAt: run.StartedAt, FinishedAt: run.FinishedAt})
	text = strings.Replace(text, " (exit 0)", "", 1)
	if run.Message != "" {
		text += ": " + run.Message
	}
	return text
}
