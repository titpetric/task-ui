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

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/model"
)

// maxWait bounds how long a call waits for a run to end.
const maxWait = 15 * time.Minute

// pollEvery is how often a waiting call looks at the run.
var pollEvery = time.Second

// tools are the server's tools. The set does not grow: a change adds a
// tool only for a need these cannot serve, and states the reason
// (docs/guide/mcp.md).
func tools() []Tool {
	str := func(text string) Property { return Property{Type: "string", Description: text} }
	num := func(text string) Property { return Property{Type: "integer", Description: text} }
	wait := num("Seconds to wait for the run to end, up to 900; 0 returns once it started.")
	lines := num("Lines of output to return from the end, 100 unless given; 0 for none.")
	return append([]Tool{
		{
			Name:        "list_jobs",
			Description: "List the folders you can see and their jobs: whether you may run each, and how it last went.",
			InputSchema: object(map[string]Property{"folder": str("Only this folder, such as deploy/terraform.")}),
			call:        listJobs,
		},
		{
			Name:        "list_workflows",
			Description: "List the workflows, each a chain of jobs and other workflows, with their last run.",
			InputSchema: object(map[string]Property{}),
			call:        listWorkflows,
		},
		{
			Name:        "run_job",
			Description: "Start a job of a folder, and wait for it to end if asked. Returns the run and the end of its output.",
			InputSchema: object(map[string]Property{
				"folder": str("The folder, such as deploy/scripts."),
				"job":    str("The job, as list_jobs names it."),
				"force":  {Type: "boolean", Description: "Run out of hours, or in place of a run of the job that is in progress."},
				"reason": str("Why the run is forced; required out of hours."),
				"wait":   wait,
				"lines":  lines,
			}, "folder", "job"),
			call: runJob,
		},
		{
			Name:        "run_workflow",
			Description: "Start a workflow by ID or name, and wait for it to end if asked. Returns the run and its steps.",
			InputSchema: object(map[string]Property{
				"workflow": str("The workflow's ID or name."),
				"input":    str("Text for the workflow's prompt, which Claude reads after it under Input:, for a workflow with a prompt."),
				"wait":     wait,
			}, "workflow"),
			call: runWorkflow,
		},
		{
			Name:        "list_runs",
			Description: "List the latest runs, newest first, by folder, job or state.",
			InputSchema: object(map[string]Property{
				"folder": str("Only runs of this folder."),
				"job":    str("Only runs of this job; use with folder."),
				"state":  {Type: "string", Description: "Only runs in this state.", Enum: model.States()},
				"limit":  num("How many, 20 unless given."),
			}),
			call: listRuns,
		},
		{
			Name:        "get_run",
			Description: "Show a run: its job, who started it, how it went, and the end of its output.",
			InputSchema: object(map[string]Property{"id": str("The run's ID."), "lines": lines}, "id"),
			call:        getRun,
		},
		{
			Name: "list_notes",
			Description: "List the notes on the board with their replies, newest first. By default the open notes that name you with @: " +
				"each is an instruction to you, from whoever may run jobs, to do and report on with note_comment. Notes to other users are not yours.",
			InputSchema: object(map[string]Property{
				"mention": str("Only notes that name this user; empty is you, * is anyone."),
				"state":   {Type: "string", Description: "Only notes in this state; open unless given, * for any.", Enum: []string{"open", "working", "done", "failed", "*"}},
				"limit":   num("How many, 20 unless given."),
			}),
			call: listNotes,
		},
		{
			Name: "note_comment",
			Description: "Reply to a note, and set the state of its instruction: working when you start, done or failed when you end, " +
				"with what you did or why it failed. The state is for a user the note names, its author or an admin; anyone may reply. " +
				"A change the author wants is a new note.",
			InputSchema: object(map[string]Property{
				"id":    str("The note's ID, from list_notes."),
				"body":  str("The reply: what you did, what changed, where to look."),
				"state": {Type: "string", Description: "The state of the instruction, left as it is unless given.", Enum: []string{"working", "done", "failed"}},
			}, "id"),
			call: noteComment,
		},
		{
			Name:        "list_workflow_runs",
			Description: "List a workflow's latest runs with how each step went.",
			InputSchema: object(map[string]Property{
				"workflow": str("The workflow's ID or name."),
				"limit":    num("How many, 5 unless given."),
			}, "workflow"),
			call: listWorkflowRuns,
		},
	}, append(append(append(noteTools(), artifactTools()...), workflowTools()...), whoamiTool(), keyTool())...)
}

// decode reads a tool's arguments into v; none is an empty object.
func decode(args json.RawMessage, v any) error {
	if len(args) == 0 || string(args) == "null" {
		return nil
	}
	if err := json.Unmarshal(args, v); err != nil {
		return fmt.Errorf("arguments: %w", err)
	}
	return nil
}

type folderArgs struct {
	Folder string `json:"folder"`
}

type folderList struct {
	Folders []api.FolderView `json:"folders"`
}

func listJobs(ctx context.Context, c *client, args json.RawMessage) (string, error) {
	var in folderArgs
	if err := decode(args, &in); err != nil {
		return "", err
	}
	var out folderList
	if err := c.get(ctx, "/folders?folder="+url.QueryEscape(in.Folder), &out); err != nil {
		return "", err
	}
	if len(out.Folders) == 0 {
		return "No folders.", nil
	}
	var b strings.Builder
	for _, f := range out.Folders {
		fmt.Fprintf(&b, "%s\n", f.ID)
		if f.Error != "" {
			fmt.Fprintf(&b, "  (cannot read: %s)\n", f.Error)
		}
		for _, job := range f.Jobs {
			fmt.Fprintf(&b, "  %s", job.ID)
			if job.Description != "" {
				fmt.Fprintf(&b, " - %s", job.Description)
			}
			if !job.CanRun {
				b.WriteString(" [view only]")
			}
			if job.LastRun != nil {
				fmt.Fprintf(&b, " | last: %s", outcome(job.LastRun))
			}
			b.WriteString("\n")
		}
	}
	return b.String(), nil
}

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

type runArgs struct {
	Folder string `json:"folder"`
	Job    string `json:"job"`
	Force  bool   `json:"force"`
	Reason string `json:"reason"`
	Wait   int    `json:"wait"`
	Lines  *int   `json:"lines"`
}

// busyRun is the API's answer to a job that is running already.
type busyRun struct {
	Run api.RunView `json:"run"`
}

func runJob(ctx context.Context, c *client, args json.RawMessage) (string, error) {
	var in runArgs
	if err := decode(args, &in); err != nil {
		return "", err
	}
	data, err := c.call(ctx, http.MethodPost, "/runs", api.RunRequest{Folder: in.Folder, Job: in.Job, Force: in.Force, Reason: in.Reason, Cols: 120, Rows: 40})
	var e *apiError
	if errors.As(err, &e) && e.Status == http.StatusConflict {
		var busy busyRun
		if json.Unmarshal(e.Body, &busy) == nil && busy.Run.ID != "" {
			return "", fmt.Errorf("%s/%s is already running as run %s; force replaces it", in.Folder, in.Job, busy.Run.ID)
		}
	}
	if err != nil {
		return "", err
	}
	var view api.RunView
	if err := json.Unmarshal(data, &view); err != nil {
		return "", err
	}
	return showRun(ctx, c, view.ID, in.Wait, in.Lines)
}

type getRunArgs struct {
	ID    string `json:"id"`
	Lines *int   `json:"lines"`
}

func getRun(ctx context.Context, c *client, args json.RawMessage) (string, error) {
	var in getRunArgs
	if err := decode(args, &in); err != nil {
		return "", err
	}
	if in.ID == "" {
		return "", errors.New("a run ID is needed")
	}
	return showRun(ctx, c, in.ID, 0, in.Lines)
}

// showRun waits up to wait seconds for a run to end, then describes it
// with the end of its output.
func showRun(ctx context.Context, c *client, id string, wait int, lines *int) (string, error) {
	path := "/runs/" + url.PathEscape(id)
	var view api.RunView
	if err := c.get(ctx, path, &view); err != nil {
		return "", err
	}
	if wait > 0 && !view.Run.Finished() {
		deadline := time.Now().Add(min(time.Duration(wait)*time.Second, maxWait))
		for !view.Run.Finished() && time.Now().Before(deadline) {
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(pollEvery):
			}
			if err := c.get(ctx, path, &view); err != nil {
				return "", err
			}
		}
	}
	var b strings.Builder
	fmt.Fprintf(&b, "run %s: %s by %s, %s\n", view.ID, view.Path(), view.UserName, outcome(&view.Run))
	if view.Reason != "" {
		fmt.Fprintf(&b, "forced: %s\n", view.Reason)
	}
	if view.Title != "" {
		fmt.Fprintf(&b, "title: %s\n", view.Title)
	}
	n := 100
	if lines != nil {
		n = *lines
	}
	if n > 0 {
		data, err := c.call(ctx, http.MethodGet, path+"/frame.txt?all=1", nil)
		if err == nil {
			b.WriteString("\n")
			b.WriteString(tail(string(data), n))
		}
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

// showFlowRun describes a workflow run and its steps.
func showFlowRun(view api.WorkflowRunView, names map[string]string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "workflow run %s: %s by %s, %s\n", view.ID, view.Name, view.UserName, flowOutcome(&view.WorkflowRun))
	for i, step := range view.Steps {
		name := step.Folder + "/" + step.Job
		switch {
		case step.IsPrompt():
			name = "prompt in " + step.Folder
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
	if view.Answer != "" {
		b.WriteString("answer:\n" + view.Answer + "\n")
	}
	return b.String()
}

type runsArgs struct {
	Folder string `json:"folder"`
	Job    string `json:"job"`
	State  string `json:"state"`
	Limit  int    `json:"limit"`
}

type runList struct {
	Runs []api.RunView `json:"runs"`
}

func listRuns(ctx context.Context, c *client, args json.RawMessage) (string, error) {
	in := runsArgs{Limit: 20}
	if err := decode(args, &in); err != nil {
		return "", err
	}
	q := url.Values{}
	q.Set("folder", in.Folder)
	q.Set("job", in.Job)
	q.Set("state", in.State)
	q.Set("limit", fmt.Sprint(in.Limit))
	var out runList
	if err := c.get(ctx, "/runs?"+q.Encode(), &out); err != nil {
		return "", err
	}
	if len(out.Runs) == 0 {
		return "No runs.", nil
	}
	var b strings.Builder
	for _, view := range out.Runs {
		fmt.Fprintf(&b, "%s %s by %s: %s\n", view.ID, view.Path(), view.UserName, outcome(&view.Run))
		if view.Title != "" {
			fmt.Fprintf(&b, "  title: %s\n", view.Title)
		}
	}
	return b.String(), nil
}

// outcome says how a run went, when, and for how long.
func outcome(run *model.Run) string {
	when := ""
	if run.StartedAt != nil {
		when = " at " + run.StartedAt.Local().Format("2006-01-02 15:04:05")
	}
	took := run.Elapsed().Round(time.Second).String()
	switch {
	case run.State == model.StateQueued:
		return fmt.Sprintf("queued for %s, behind the run of its job before it, since%s", took, strings.TrimPrefix(when, " at"))
	case !run.Finished():
		return fmt.Sprintf("%s for %s, started%s", run.State, took, when)
	case run.Message != "":
		return fmt.Sprintf("%s (exit %d) in %s, started%s: %s", run.State, run.ExitCode, took, when, run.Message)
	}
	return fmt.Sprintf("%s (exit %d) in %s, started%s", run.State, run.ExitCode, took, when)
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

// tail returns the last n lines of text, without the blank ones at the end.
func tail(text string, n int) string {
	lines := strings.Split(strings.TrimRight(text, " \n"), "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return strings.Join(lines, "\n") + "\n"
}

type notesArgs struct {
	Mention string `json:"mention"`
	State   string `json:"state"`
	Limit   int    `json:"limit"`
}

type noteList struct {
	Notes []api.NoteView `json:"notes"`
}

func listNotes(ctx context.Context, c *client, args json.RawMessage) (string, error) {
	in := notesArgs{State: model.NoteOpen, Limit: 20}
	if err := decode(args, &in); err != nil {
		return "", err
	}
	// The API takes * as any user and any state; empty mention is the
	// caller here.
	if in.Mention == "" {
		if user := access.UserFrom(ctx); user != nil {
			in.Mention = user.Name
		}
	}
	q := url.Values{}
	q.Set("mention", in.Mention)
	q.Set("state", in.State)
	q.Set("limit", fmt.Sprint(in.Limit))
	var out noteList
	if err := c.get(ctx, "/notes?"+q.Encode(), &out); err != nil {
		return "", err
	}
	if len(out.Notes) == 0 {
		return "No notes.", nil
	}
	shown := make([]string, 0, len(out.Notes))
	for _, view := range out.Notes {
		shown = append(shown, showNote(view))
	}
	return strings.Join(shown, "\n"), nil
}

// showNote is a note with its replies, as text.
func showNote(view api.NoteView) string {
	var b strings.Builder
	state := view.State
	if state == "" {
		state = "message"
	}
	fmt.Fprintf(&b, "note %s by %s, %s, %s\n%s\n", view.ID, view.UserName, state, when(view.CreatedAt), view.Body)
	if view.RunID != "" {
		fmt.Fprintf(&b, "(run %s)\n", view.RunID)
	}
	for _, reply := range view.Replies {
		fmt.Fprintf(&b, "  reply by %s, %s: %s\n", reply.UserName, when(reply.CreatedAt), strings.ReplaceAll(reply.Body, "\n", "\n    "))
	}
	return b.String()
}

// when is a time as the tools print it.
func when(t *time.Time) string {
	if t == nil {
		return "-"
	}
	return t.Local().Format("2006-01-02 15:04:05")
}

type commentArgs struct {
	ID    string `json:"id"`
	Body  string `json:"body"`
	State string `json:"state"`
}

func noteComment(ctx context.Context, c *client, args json.RawMessage) (string, error) {
	var in commentArgs
	if err := decode(args, &in); err != nil {
		return "", err
	}
	if in.ID == "" {
		return "", errors.New("a note ID is needed")
	}
	if strings.TrimSpace(in.Body) == "" && in.State == "" {
		return "", errors.New("a reply, a state, or both are needed")
	}
	if _, err := c.call(ctx, http.MethodPost, "/notes/"+url.PathEscape(in.ID)+"/comments", api.CommentRequest{Body: in.Body, State: in.State}); err != nil {
		return "", err
	}
	text := "Replied to note " + in.ID
	if in.State != "" {
		text += "; it is " + in.State
	}
	return text + ".", nil
}
