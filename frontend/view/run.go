package view

import (
	"html/template"

	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/recording"
)

// Run is a run as the pages show it: the row, and what the runner knows
// while the session is in memory.
type Run struct {
	model.Run
	// Live holds true while the session accepts an attach.
	Live bool
	// Attached counts the windows on the session.
	Attached int
	// Message is what the status says, when it says something.
	Message string
	// FrameSVG is the last frame of the terminal, rendered, when the
	// page shows it.
	FrameSVG template.HTML
	// Expired says the Data retention settings removed the run's
	// recording; the run's page says so in place of its output.
	Expired bool
}

// ExpiredText is what a page says in place of the output of a run
// whose recording expired.
func (r Run) ExpiredText() string {
	return recording.Expired
}

// Initials returns the initials of the user who started the run.
func (r Run) Initials() string {
	return (&model.User{Name: r.UserName}).Initials()
}

// IsScheduled reports whether the scheduler started the run.
func (r Run) IsScheduled() bool {
	return r.Run.Scheduled != 0
}

// IsForced reports whether the run killed the one before it.
func (r Run) IsForced() bool {
	return r.Run.Forced != 0
}
