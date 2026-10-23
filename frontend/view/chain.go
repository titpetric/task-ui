package view

import (
	"github.com/titpetric/task-ui/model"
)

// Chain turns steps into links, a folder named once for the steps that follow in it, each linked to its page or the run it made.
func Chain(steps []Step, workspaces map[string]int) []Link {
	links := make([]Link, 0, len(steps))
	last := ""
	for _, s := range steps {
		l := Link{State: s.State, RunID: s.RunID, Disabled: s.Disabled}
		switch s.StepKind() {
		case model.StepSummary:
			place := model.WorkflowStep{Folder: s.Folder, Dir: s.Dir}.Place()
			l.Summary, l.Text, l.Href = true, "summary in "+place, FolderHref(s.Folder, workspaces)
			last = ""
		case model.StepPrompt:
			place := model.WorkflowStep{Folder: s.Folder, Dir: s.Dir}.Place()
			l.Prompt, l.Text, l.Href = s.Prompt, "prompt in "+place, FolderHref(s.Folder, workspaces)
			last = ""
		case model.StepWorkflow:
			l.Workflow, l.Text, l.Href = true, s.WorkflowName, "/workflows/"+s.WorkflowID
			l.WorkflowID, l.WorkflowRunID = s.WorkflowID, s.WorkflowRunID
			if l.Text == "" {
				l.Text = "a workflow that is gone"
			}
			last = ""
		default:
			l.Text, l.Href = s.Job, FolderHref(s.Folder, workspaces)+"/"+s.Job
			if s.Folder != last {
				l.Text = s.Folder + " / " + s.Job
			}
			last = s.Folder
		}
		if s.RunID != "" {
			l.Href = "/runs/" + s.RunID
		}
		links = append(links, l)
	}
	return links
}
