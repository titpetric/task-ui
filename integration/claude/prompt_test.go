package claude_test

import (
	"testing"

	"github.com/titpetric/task-ui/claude"
	"github.com/titpetric/task-ui/internal/assert"
)

// A run's summary is short: a paragraph or two, a table of what is worth
// seeing, and no changes.
func TestRunSummary(t *testing.T) {
	t.Parallel()
	assert.Contains(t, claude.RunSummary, "one or two short paragraphs")
	assert.Contains(t, claude.RunSummary, "a markdown table of them")
	assert.Contains(t, claude.RunSummary, "No preamble, no headings")
	assert.Contains(t, claude.RunSummary, "Change nothing.")
}

func TestHostsSummary(t *testing.T) {
	t.Parallel()
	assert.Contains(t, claude.HostsSummary, "one or two short paragraphs")
	assert.Contains(t, claude.HostsSummary, "a markdown table with a row a host")
	assert.Contains(t, claude.HostsSummary, "Change nothing.")
}

func TestWorkflowBuilder(t *testing.T) {
	t.Parallel()
	assert.Contains(t, claude.WorkflowBuilder, "create_workflow or update_workflow")
	assert.Contains(t, claude.WorkflowBuilder, "at most 3 runs")
	assert.Contains(t, claude.WorkflowBuilder, "verbatim")
}

// A triage finds the cause, fixes it in the folder alone and ends with
// the one line the scheduler reads.
func TestRunTriage(t *testing.T) {
	t.Parallel()
	assert.Contains(t, claude.RunTriage, "Find the root cause")
	assert.Contains(t, claude.RunTriage, "Do not change task-ui's settings, its users or its secrets")
	assert.Contains(t, claude.RunTriage, "`Result: resolved`, `Result: not resolved` or `Result: no action`")
}
