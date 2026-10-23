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
