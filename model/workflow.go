package model

// The states a workflow run's step has beyond those of a run: it waits
// for the steps before it, and it skips when one of them did not
// end ok. The values mirror the CHECK in schema/workflow_run_step.up.sql.
const (
	// StepPending is a step whose turn has not come.
	StepPending State = "pending"
	// StepSkipped is a step that did not run: a disabled one, one after
	// a failure, or one after something ended the workflow run.
	StepSkipped State = "skipped"
)

// The kinds of a workflow step, its kind column: a job of a folder,
// another workflow, a prompt to Claude, or task-ui's summary of the
// steps before it. The values mirror the CHECK in
// schema/workflow_step_with_kind.up.sql.
const (
	StepJob      = "job"
	StepWorkflow = "workflow"
	StepPrompt   = "prompt"
	StepSummary  = "summary"
)

// stepKind is a step's kind, from its fields for a row written
// without one.
func stepKind(kind, prompt, ref string) string {
	switch {
	case kind != "":
		return kind
	case prompt != "":
		return StepPrompt
	case ref != "":
		return StepWorkflow
	}
	return StepJob
}

// StepKind is what the step does: StepJob, StepWorkflow, StepPrompt or StepSummary.
func (s WorkflowStep) StepKind() string { return stepKind(s.Kind, s.Prompt, s.WorkflowRef) }

// StepKind is what the step did, as WorkflowStep.StepKind says.
func (s WorkflowRunStep) StepKind() string { return stepKind(s.Kind, s.Prompt, s.WorkflowRef) }

// IsWorkflow says the step runs a workflow in place of a job.
func (s WorkflowStep) IsWorkflow() bool { return s.StepKind() == StepWorkflow }

// IsWorkflow says the step ran a workflow in place of a job.
func (s WorkflowRunStep) IsWorkflow() bool { return s.StepKind() == StepWorkflow }

// IsPrompt says the step asks Claude in place of running a job: a prompt or a summary.
func (s WorkflowStep) IsPrompt() bool { return asks(s.StepKind()) }

// IsPrompt says the step asked Claude in place of running a job.
func (s WorkflowRunStep) IsPrompt() bool { return asks(s.StepKind()) }

// IsSummary says the step asks Claude for task-ui's summary of the steps before it.
func (s WorkflowStep) IsSummary() bool { return s.StepKind() == StepSummary }

// IsSummary says the step asked Claude for the summary of the steps before it.
func (s WorkflowRunStep) IsSummary() bool { return s.StepKind() == StepSummary }

func asks(kind string) bool { return kind == StepPrompt || kind == StepSummary }

// Place is where a prompt step asks Claude: its folder and the
// directory inside it, or _ for the root.
func (s WorkflowStep) Place() string {
	folder := s.Folder
	if folder == "" {
		folder = Default
	}
	if s.Dir == "" {
		return folder
	}
	return folder + "/" + s.Dir
}

// Runs reports whether the workflow has something to run with these
// steps: an enabled step.
func (w *Workflow) Runs(steps []WorkflowStep) bool {
	for _, step := range steps {
		if !step.Disabled {
			return true
		}
	}
	return false
}

// Asks reports whether a step of the workflow asks Claude, so its runs
// take input.
func (w *Workflow) Asks(steps []WorkflowStep) bool {
	for _, step := range steps {
		if step.IsPrompt() && !step.Disabled {
			return true
		}
	}
	return false
}

// RunsAs is the ID of the user every run of the workflow runs as: the
// one set on it, else its owner.
func (w *Workflow) RunsAs() string {
	if w.RunAsUserID != "" {
		return w.RunAsUserID
	}
	return w.UserID
}
