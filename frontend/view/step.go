package view

import (
	"strings"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/model"
)

// StepScript is the kind the editor shows a job step of @script as: a
// command the step runs with bash, which the API keeps as a job step.
const StepScript = "script"

// Step is a step of a workflow, or of a run of one, as a page shows
// it: a job of a folder, another workflow, a prompt to Claude or a
// summary by Claude, as Kind says.
type Step struct {
	// Kind is model.StepJob, StepWorkflow, StepPrompt or StepSummary;
	// empty reads as the fields say, as model.WorkflowStep.StepKind.
	Kind   string
	Folder string
	Job    string
	// WorkflowID and WorkflowName are the workflow a step runs in place
	// of a job.
	WorkflowID   string
	WorkflowName string
	// Disabled is a workflow's step its runs skip.
	Disabled bool
	// Args are what the step passes its job, a JSON list: the command of
	// @host:<host>.
	Args string
	// Prompt is what a prompt step asks Claude, in Folder, internal for the
	// root, and Dir inside it; empty for any other step. A summary step
	// asks in the same place.
	Prompt string
	Dir    string
	// Missing says what of the step a move to another workspace did not
	// find there, for the editor to show under it.
	Missing string
	// State is how a run's step ended; RunID the run it made of a job,
	// WorkflowRunID the run it made of a workflow.
	State         string
	RunID         string
	WorkflowRunID string
}

// NewStep returns a blank step of a kind, as the workflow editor adds
// one: a prompt or a summary asks in the root, a script is the job @script.
func NewStep(kind string) Step {
	switch kind {
	case model.StepPrompt, model.StepSummary:
		return Step{Kind: kind, Folder: model.Default}
	case StepScript:
		return Step{Kind: model.StepJob, Job: folder.ScriptID}
	}
	return Step{Kind: kind}
}

// StepKind is what the step does, from its fields when Kind is empty: StepScript for the job @script.
func (s Step) StepKind() string {
	kind := model.WorkflowStep{Kind: s.Kind, Prompt: s.Prompt, WorkflowRef: s.WorkflowID}.StepKind()
	if kind == model.StepJob && s.Job == folder.ScriptID {
		return StepScript
	}
	return kind
}

// Command is the command of a script step, its one arg.
func (s Step) Command() string {
	return strings.Join(model.DecodeList(s.Args), " ")
}

// KindLabel names the step's kind as the editor shows it.
func (s Step) KindLabel() string {
	switch s.StepKind() {
	case model.StepWorkflow:
		return "Workflow"
	case model.StepPrompt:
		return "Prompt"
	case model.StepSummary:
		return "Summary"
	case StepScript:
		return "Script"
	}
	return "Command"
}

// Icon is the symbol of the step's kind, #i-<Icon> of the sprite.
func (s Step) Icon() string {
	switch s.StepKind() {
	case model.StepWorkflow:
		return "flow"
	case model.StepPrompt:
		return "prompt"
	case model.StepSummary:
		return "summary"
	case StepScript:
		return "script"
	}
	return "terminal"
}

// Place is where a prompt or summary step asks Claude: the root, the workspace root, or folder/dir.
func (s Step) Place() string {
	return model.PlaceLabel(model.WorkflowStep{Folder: s.Folder, Dir: s.Dir}.Place())
}

// Line is the step in one line, its card's heading while folded: the
// job and its args, the workflow, the first line of the prompt, or
// where the summary asks.
func (s Step) Line() string {
	switch s.StepKind() {
	case model.StepWorkflow:
		if s.WorkflowName == "" {
			return "a workflow that is gone"
		}
		return s.WorkflowName
	case model.StepPrompt:
		line, _, _ := strings.Cut(strings.TrimSpace(s.Prompt), "\n")
		return line
	case model.StepSummary:
		return "of the steps before it, in " + s.Place()
	case StepScript:
		line, _, _ := strings.Cut(strings.TrimSpace(s.Command()), "\n")
		if s.Folder == "" || line == "" {
			return ""
		}
		return s.Folder + ": " + line
	case model.StepJob:
		if s.Job == "" {
			return ""
		}
	}
	line := s.Folder + " / " + s.Job
	if args := ArgLine(s.Job, s.Args); args != "" {
		line += " " + args
	}
	return line
}
