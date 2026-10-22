package claude_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/titpetric/task-ui/claude"
	"github.com/titpetric/task-ui/internal/assert"
)

// fakeClaude echoes its system prompt, its prompt, the token it signed in with and the
// document it read; with the prompt "fail" it fails.
func fakeClaude(t *testing.T) string {
	t.Helper()
	script := filepath.Join(t.TempDir(), "claude")
	body := "#!/bin/sh\nfor last; do :; done\n[ \"$last\" = fail ] && { echo broken >&2; exit 1; }\n[ \"$last\" = slow ] && sleep 5\n[ \"$4\" = --append-system-prompt ] && echo \"system=$5\"\necho \"prompt=$last token=$CLAUDE_CODE_OAUTH_TOKEN\"\ncat\n"
	assert.NoError(t, os.WriteFile(script, []byte(body), 0o755))
	return script
}

func TestSummarizer_Summarize(t *testing.T) {
	t.Parallel()
	s := &claude.Summarizer{Command: fakeClaude(t), Timeout: 5 * time.Second}
	out, err := s.Summarize(context.Background(), "sk-1", "", "Summarize.", "## a\n\nok\n")
	assert.NoError(t, err)
	assert.Equal(t, "prompt=Summarize. token=sk-1\n## a\n\nok", out)

	_, err = s.Summarize(context.Background(), "", "", "Summarize.", "x")
	assert.ErrorIs(t, err, claude.ErrNoToken)
	_, err = s.Summarize(context.Background(), "sk-1", "", "fail", "x")
	assert.ErrorContains(t, err, "broken")

	quick := &claude.Summarizer{Command: fakeClaude(t), Timeout: 200 * time.Millisecond}
	_, err = quick.Summarize(context.Background(), "sk-1", "", "slow", "x")
	assert.ErrorContains(t, err, "ran past")

	// The system prompt goes ahead of the prompt.
	out, err = s.Summarize(context.Background(), "sk-1", "You work in task-ui.", "Summarize.", "x")
	assert.NoError(t, err)
	assert.Equal(t, "system=You work in task-ui.\nprompt=Summarize. token=sk-1\nx", out)
}
