package agent

import (
	"strings"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
)

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
