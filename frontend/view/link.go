package view

import (
	"html/template"
)

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
