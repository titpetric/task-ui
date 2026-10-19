package view

import (
	"html/template"
)

// Step is a step of a workflow, or of a run of one, as a page shows
// it: a job of a folder, or another workflow.
type Step struct {
	Folder string
	Job    string
	// WorkflowID and WorkflowName are the workflow a step runs in place
	// of a job.
	WorkflowID   string
	WorkflowName string
	// State is how a run's step ended; RunID the run it made of a job,
	// WorkflowRunID the run it made of a workflow.
	State         string
	RunID         string
	WorkflowRunID string
}

// Link is one step of a chain as a page draws it.
type Link struct {
	Text  string
	Href  string
	State string
	// RunID is the run a run's step made of a job, to attach to.
	RunID string
	// Workflow says the step runs a workflow: WorkflowID, and for a run's
	// step WorkflowRunID, the run it made of it.
	Workflow      bool
	WorkflowID    string
	WorkflowRunID string
	// Sub is the chain of the run the step made of a workflow, when the
	// page shows it.
	Sub []Link
	// FrameSVG is the last screen of the step's run, when the page shows
	// it.
	FrameSVG template.HTML
	// Complete says the frame already shows the run's whole output, so
	// there is nothing to expand; Lines is how many lines that output
	// has, 0 when not known.
	Complete bool
	Lines    int
}

// Chain turns steps into links. A folder is named once, for the steps
// that follow in it; a workflow step is named by its name. A run's step
// links to the run it made.
func Chain(steps []Step) []Link {
	links := make([]Link, 0, len(steps))
	last := ""
	for _, s := range steps {
		l := Link{State: s.State, RunID: s.RunID}
		switch {
		case s.WorkflowID != "":
			l.Workflow, l.Text, l.Href = true, s.WorkflowName, "/workflows/"+s.WorkflowID
			l.WorkflowID, l.WorkflowRunID = s.WorkflowID, s.WorkflowRunID
			if l.Text == "" {
				l.Text = "a workflow that is gone"
			}
			last = ""
		default:
			l.Text, l.Href = s.Job, "/f/"+s.Folder+"/"+s.Job
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
