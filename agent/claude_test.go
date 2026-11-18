package agent

import (
	"strings"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/internal/claudesession"
)

// A run names its session on the first start and resumes it after a restart; no session takes no flags.
func TestAgent_sessionArgs(t *testing.T) {
	t.Parallel()
	id := "01arz3ndektsv4rrffq69g5fav"
	fresh := &Agent{Env: []string{"TASK_UI_CLAUDE_SESSION=" + id}}
	assert.Equal(t, []string{"--session-id", claudesession.UUID(id)}, fresh.sessionArgs())
	resumed := &Agent{Env: []string{"TASK_UI_CLAUDE_SESSION=" + id, "TASK_UI_CLAUDE_RESUME=1"}}
	assert.Equal(t, []string{"--resume", claudesession.UUID(id)}, resumed.sessionArgs())
	assert.Len(t, (&Agent{}).sessionArgs(), 0)
}

// A run with a session config reads Claude Code's config directory from it.
func TestAgent_environ(t *testing.T) {
	t.Parallel()
	a := &Agent{Env: []string{"TASK_UI_CLAUDE_CONFIG=/data/claude/ana_x"}}
	assert.Contains(t, a.environ(), "CLAUDE_CONFIG_DIR=/data/claude/ana_x")
}

// A streamed run writes Claude's text and tools to the terminal and its result to the answer.
func TestStreamClaude(t *testing.T) {
	t.Parallel()
	stream := strings.Join([]string{
		`{"type":"system","subtype":"init"}`,
		`{"type":"assistant","message":{"content":[{"type":"text","text":"Looking."}]}}`,
		`{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Bash","input":{"command":"go test ./..."}}]}}`,
		`{"type":"assistant","message":{"content":[{"type":"text","text":"All green."}]}}`,
		`{"type":"result","result":"All green."}`,
	}, "\n")
	var progress, answer strings.Builder
	streamClaude(strings.NewReader(stream), &progress, &answer)
	assert.Contains(t, progress.String(), "Looking.")
	assert.Contains(t, progress.String(), "> Bash: go test ./...")
	assert.Contains(t, progress.String(), "All green.")
	assert.Equal(t, "All green.\n", answer.String())
}

// A line that is not stream-json reaches the terminal, and a run with no answer writer still streams.
func TestStreamClaudePlainLine(t *testing.T) {
	t.Parallel()
	var progress strings.Builder
	streamClaude(strings.NewReader("a warning\n"), &progress, nil)
	assert.Equal(t, "a warning\n", progress.String())
}

// A tool line takes the first value that says what the tool works on, on one line.
func TestToolLine(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "> Read", toolLine("Read", nil))
	assert.Equal(t, "> Bash: go build ./...", toolLine("Bash", []byte(`{"command":"go build\n./..."}`)))
}
