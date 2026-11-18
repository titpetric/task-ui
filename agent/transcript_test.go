package agent

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
)

// lineRecord is one line of a transcript as the test reads it back.
type lineRecord struct {
	Seq    int64          `json:"seq"`
	Kind   string         `json:"kind"`
	Body   string         `json:"body"`
	ToolID string         `json:"tool_id"`
	Data   map[string]any `json:"data"`
}

// TestTranscript_WritesRecords asserts each kind is one JSON line, in order.
func TestTranscript_WritesRecords(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "events.jsonl")
	trace, err := openTranscript(path)
	if err != nil {
		t.Fatal(err)
	}

	trace.status("Thinking", "thinking")
	trace.text("on it", "msg1", 0)
	trace.toolUse("Bash", "tool1", json.RawMessage(`{"command":"go test ./..."}`))
	trace.toolResult("tool1", "ok", false)
	trace.result(claudeEvent{Subtype: "success", Result: "done", DurationMS: 1200})
	trace.error("boom")
	trace.close()

	records := readRecords(t, path)
	assert.Len(t, records, 6)
	assert.Equal(t, "status", records[0].Kind)
	assert.Equal(t, "on it", records[1].Body)
	assert.Equal(t, "tool1", records[2].ToolID)
	assert.Equal(t, "tool_result", records[3].Kind)
	assert.Equal(t, "result", records[4].Kind)
	assert.Equal(t, "error", records[5].Kind)
	for i, record := range records {
		assert.Equal(t, int64(i+1), record.Seq)
	}
}

// TestTranscript_CutsText asserts a body over its cap is cut and marked.
func TestTranscript_CutsText(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "events.jsonl")
	trace, err := openTranscript(path)
	if err != nil {
		t.Fatal(err)
	}
	trace.text(strings.Repeat("x", maxText+10), "msg1", 0)
	trace.close()

	records := readRecords(t, path)
	assert.Len(t, records, 1)
	assert.True(t, strings.HasSuffix(records[0].Body, "... (truncated)"), "the body says it was cut")
	assert.True(t, records[0].Data["truncated"] == true, "the data says it was cut")
}

// TestTranscript_Empty asserts a run with no file keeps nothing and closes quietly.
func TestTranscript_Empty(t *testing.T) {
	t.Parallel()
	trace, err := openTranscript("")
	assert.NoError(t, err)
	trace.status("Thinking", "thinking")
	trace.close()
}

// readRecords parses a transcript file.
func readRecords(t *testing.T, path string) []lineRecord {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	var out []lineRecord
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64<<10), 2<<20)
	for scanner.Scan() {
		var record lineRecord
		assert.NoError(t, json.Unmarshal(scanner.Bytes(), &record))
		out = append(out, record)
	}
	return out
}
