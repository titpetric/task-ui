package view

import (
	"time"

	"github.com/titpetric/task-ui/model"
)

// Workflow is a workflow as a page shows it.
type Workflow struct {
	model.Workflow
	Steps []Step
	// Running is the ID of its run in progress, if any.
	Running string
	// LastRun is its newest run.
	LastRun *model.WorkflowRun
	// RunAs names the user its runs run as, its owner without one; ""
	// for a user nobody holds any more.
	RunAs string
	// CanRun says whether its runs may go: the user it runs as may run
	// every step.
	CanRun bool
	// CanChange says whether the user may change or delete it.
	CanChange bool
	// Schedules are the schedules that start it, for the list's hint
	// under its name.
	Schedules []model.Schedule
}

// Prompted reports whether a step of the workflow asks Claude, a prompt or a summary, so its Run takes input.
func (w Workflow) Prompted() bool {
	for _, step := range w.Steps {
		if kind := step.StepKind(); !step.Disabled && (kind == model.StepPrompt || kind == model.StepSummary) {
			return true
		}
	}
	return false
}

// PromptPlace is where the workflow's first prompt or summary step asks
// Claude, the place its thread shows before it starts: _ without one.
func (w Workflow) PromptPlace() string {
	for _, step := range w.Steps {
		if kind := step.StepKind(); kind == model.StepPrompt || kind == model.StepSummary {
			return model.WorkflowStep{Folder: step.Folder, Dir: step.Dir}.Place()
		}
	}
	return model.Default
}

// WorkflowRun is a run of a workflow with its steps, and the thread it
// is: Question, what it started with, under its starter, and its answer
// under Agent, the reply.
type WorkflowRun struct {
	model.WorkflowRun
	Steps []Step
	// Question is the run's input, or else its first prompt: none for a
	// run of commands started without input.
	Question string
	// Agent is the name Claude answers with, for a run that asks Claude;
	// empty for one of commands alone, whose reply is how it ended.
	Agent string
}

// Running reports whether the run has not ended, its reply still to come.
func (r WorkflowRun) Running() bool {
	return !model.Finished(r.State)
}

// NoteState is the run's state as its thread shows it, a note's: working
// while it runs, done once it ended ok, else failed.
func (r WorkflowRun) NoteState() string {
	switch {
	case r.Running():
		return model.NoteWorking
	case r.State == model.StateOK:
		return model.NoteDone
	}
	return model.NoteFailed
}

// Took is how long the run took so far, on a copy, which the pointer method Elapsed does not allow.
func (r WorkflowRun) Took() time.Duration {
	return r.WorkflowRun.Elapsed()
}
