package claude_test

import (
	"strings"
	"testing"

	"github.com/titpetric/task-ui/claude"
	"github.com/titpetric/task-ui/config"
	"github.com/titpetric/task-ui/internal/assert"
)

// The context names the user, the role and what the user may do, ahead
// of context.md.
func TestContext(t *testing.T) {
	t.Parallel()
	text := claude.Context("claude", "runner", "note/view, run; everything else is denied")
	assert.True(t, strings.HasPrefix(text, "You work in task-ui for the user claude, role runner, who may: note/view, run; everything else is denied."))
	assert.True(t, strings.HasSuffix(text, config.Context()))
	assert.Equal(t, config.Context(), claude.Context("", "", ""))
}
