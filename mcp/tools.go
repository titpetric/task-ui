package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/model"
)

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
			Description: "List the workflows, each a chain of steps (jobs, other workflows, prompts and summaries), with their last run.",
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
			Description: "Start a workflow by ID or name, and wait for it to end if asked. Returns the run, its steps, its input and, once it ended, its answer.",
			InputSchema: object(map[string]Property{
				"workflow": str("The workflow's ID or name."),
				"input":    str("Text for the workflow's prompt steps, which Claude reads after each prompt under Input:, for a workflow with prompt or summary steps."),
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
			Description: "List a workflow's latest runs with how each step went, the input and the answer.",
			InputSchema: object(map[string]Property{
				"workflow": str("The workflow's ID or name."),
				"limit":    num("How many, 5 unless given."),
			}, "workflow"),
			call: listWorkflowRuns,
		},
	}, append(append(append(noteTools(), artifactTools()...), workflowTools()...), whoamiTool())...)
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
		for _, finding := range f.Errors {
			fmt.Fprintf(&b, "  ! %s", finding.Path)
			if finding.Job != "" {
				fmt.Fprintf(&b, " job %s", finding.Job)
			}
			fmt.Fprintf(&b, ": %s\n", finding.Detail)
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
