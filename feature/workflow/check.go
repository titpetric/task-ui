package workflow

import (
	"context"
	"fmt"
	"strings"

	"github.com/titpetric/task-ui/app/model"
)

// Output reads the text a step of a workflow run left: a job step's
// terminal output, replayed from its recording. A prompt step keeps its
// answer on its row and needs no reader. note.Board is one.
type Output interface {
	StepText(ctx context.Context, row *model.WorkflowRunStep) string
}

// checked reads the step a check follows and says what it found: the
// answer a reader sees, and whether the steps go round again, which they
// do while a marker is there and the check has a pass left.
func (r *Runner) checked(ctx context.Context, step Step, before []model.WorkflowRunStep, pass int) (answer string, again bool) {
	row := ranLast(before)
	if row == nil {
		return "nothing ran before the check", false
	}
	at := fmt.Sprintf("step %d", row.Position+1)
	text := r.stepText(ctx, row)
	marker := model.Marked(step.Markers, text)
	switch {
	case text == "":
		return at + " left nothing to read", false
	case marker == "":
		return fmt.Sprintf("no %s in %s", strings.Join(step.Markers, ", "), at), false
	case pass >= step.Passes:
		return fmt.Sprintf("found %q in %s, and pass %d of %d was the last", marker, at, min(pass, step.Passes), step.Passes), false
	}
	return fmt.Sprintf("found %q in %s: the steps go round again, pass %d of %d", marker, at, pass+1, step.Passes), true
}

// stepText is the text a step left: a prompt's answer off its row, a
// job's output through the reader, "" on an instance without one.
func (r *Runner) stepText(ctx context.Context, row *model.WorkflowRunStep) string {
	if row.Answer != "" {
		return row.Answer
	}
	if r.Output == nil {
		return ""
	}
	return r.Output.StepText(ctx, row)
}

// groupStart is the first step of a project's group, where a check of
// that project sends the walk back to: the first step of all for a
// workflow that runs over no workspace, so a rerun keeps the projects
// walked before this one.
func groupStart(steps []Step, group int) int {
	for i, step := range steps {
		if step.Group == group {
			return i
		}
	}
	return 0
}

// groupAt is the project a step belongs to, 0 for a step past the end.
func groupAt(steps []Step, i int) int {
	if i < 0 || i >= len(steps) {
		return 0
	}
	return steps[i].Group
}

// ranLast is the last of the steps walked that ran and left something to
// read, the checks among them passed over, so a skipped or disabled step
// reads past to its predecessor. A row of a pass that skipped it keeps
// the run and the answer of the pass before, which is nothing to read.
func ranLast(before []model.WorkflowRunStep) *model.WorkflowRunStep {
	for i := len(before) - 1; i >= 0; i-- {
		row := &before[i]
		switch {
		case row.IsCheck(), row.State == model.StepSkipped:
		case row.RunID != "", row.Answer != "":
			return row
		}
	}
	return nil
}
