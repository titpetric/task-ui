package agent

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
)

// The prompt of a note names the agent the run answers as, and says that
// a message naming several users splits the work; without an agent it
// speaks to whoever runs it.
func TestNotePrompt(t *testing.T) {
	t.Parallel()
	prompt := notePrompt("n1", "ops", "consuela", "ops:\n@consuela and @claude, split it")
	assert.Contains(t, prompt, "note n1 by ops, for you, @consuela, to work on")
	assert.Contains(t, prompt, "splits the work between them")
	assert.Contains(t, prompt, "@consuela and @claude, split it")
	assert.Contains(t, notePrompt("n1", "ops", "", "do it"), "by ops, for you to work on")
}
