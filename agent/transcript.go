package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// The caps a transcript puts on one record, in bytes.
const (
	maxText       = 512 << 10
	maxThinking   = 256 << 10
	maxToolResult = 64 << 10
	maxToolInput  = 64 << 10
	maxSystem     = 8 << 10
)

// record holds a line of the transcript a run writes for its thread.
type record struct {
	V      int    `json:"v"`
	Seq    int64  `json:"seq"`
	TS     string `json:"ts,omitempty"`
	Kind   string `json:"kind"`
	Body   string `json:"body,omitempty"`
	ToolID string `json:"tool_id,omitempty"`
	Data   any    `json:"data,omitempty"`
}

// transcript writes a run's JSONL records, one flushed line at a time.
type transcript struct {
	mu   sync.Mutex
	file *os.File
	seq  int64
	done bool
}

// openTranscript opens the file a run writes its transcript to, nil for an empty path.
func openTranscript(path string) (*transcript, error) {
	if path == "" {
		return nil, nil
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return nil, err
	}
	return &transcript{file: file}, nil
}

// write appends one record; a nil transcript keeps nothing.
func (t *transcript) write(kind, body, toolID string, data any) {
	if t == nil || t.done {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	t.seq++
	line, err := json.Marshal(record{
		V: 1, Seq: t.seq, TS: time.Now().UTC().Format(time.RFC3339Nano),
		Kind: kind, Body: body, ToolID: toolID, Data: data,
	})
	if err != nil {
		return
	}
	_, _ = t.file.Write(append(line, '\n'))
}

// close stops the transcript and closes its file.
func (t *transcript) close() {
	if t == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	t.done = true
	_ = t.file.Close()
}

// status records what Claude is doing now, and the tool it names.
func (t *transcript) status(body, state, tool string) {
	t.write("status", body, "", map[string]string{"state": state, "tool": tool})
}

// system records the run's opening facts, without its environment.
func (t *transcript) system(event claudeEvent) {
	names := make([]string, 0, len(event.MCPServers))
	for _, server := range event.MCPServers {
		names = append(names, server.Name)
	}
	data := map[string]any{
		"subtype": event.Subtype, "model": event.Model, "cwd": event.Cwd,
		"tools": len(event.Tools), "mcp_servers": names, "session": event.SessionID,
	}
	t.write("system", "", "", data)
}

// text records a complete assistant message.
func (t *transcript) text(text, messageID string, block int) {
	body, cut := cutText(strings.TrimSpace(text), maxText)
	if body == "" {
		return
	}
	t.write("text", body, "", map[string]any{"message_id": messageID, "block": block, "truncated": cut})
}

// thinking records a complete thinking block.
func (t *transcript) thinking(text string) {
	body, cut := cutText(strings.TrimSpace(text), maxThinking)
	if body == "" {
		return
	}
	t.write("thinking", body, "", map[string]bool{"truncated": cut})
}

// toolUse records a tool call, with its arguments.
func (t *transcript) toolUse(name, id string, input json.RawMessage) {
	data := map[string]any{"tool_id": id, "name": name}
	if len(input) > maxToolInput {
		data["input"] = string(input[:maxToolInput]) + "\n... (truncated)"
		data["truncated"] = true
	} else {
		data["input"] = input
	}
	t.write("tool_use", toolLine(name, input), id, data)
}

// toolResult records a tool's output, cut when it is large.
func (t *transcript) toolResult(id, content string, isError bool) {
	body, cut := cutText(content, maxToolResult)
	t.write("tool_result", body, id, map[string]bool{"is_error": isError, "truncated": cut})
}

// result records the final answer, what the run read, and how it ended.
func (t *transcript) result(event claudeEvent) {
	input, output := event.Usage.InputTokens, event.Usage.OutputTokens
	cached, created, cost := event.Usage.CacheReadInputTokens, event.Usage.CacheCreationInputTokens, event.TotalCostUSD
	if len(event.ModelUsage) > 0 {
		input, output, cached, created = 0, 0, 0, 0
		for _, model := range event.ModelUsage {
			input, output = input+model.InputTokens, output+model.OutputTokens
			cached, created = cached+model.CacheReadInputTokens, created+model.CacheCreationInputTokens
			cost += model.CostUSD
		}
		if event.TotalCostUSD > 0 {
			cost = event.TotalCostUSD
		}
	}
	data := map[string]any{
		"subtype": event.Subtype, "is_error": event.IsError,
		"duration_ms": event.DurationMS, "num_turns": event.NumTurns,
		"input_tokens": input, "output_tokens": output,
		"cache_read_input_tokens": cached, "cache_creation_input_tokens": created,
		"cost_usd": cost,
	}
	t.write("result", cutTextUncut(event.Result), "", data)
}

// error records a process-level failure.
func (t *transcript) error(text string) {
	body, _ := cutText(strings.TrimSpace(text), maxSystem)
	t.write("error", body, "", nil)
}

// cutText cuts text to size bytes on a line boundary and reports whether it cut.
func cutText(text string, size int) (string, bool) {
	if len(text) <= size {
		return text, false
	}
	text = text[:size]
	if at := strings.LastIndexByte(text, '\n'); at > 0 {
		text = text[:at]
	}
	return text + "\n... (truncated)", true
}

// cutTextUncut bounds the final answer without marking prose that was not cut.
func cutTextUncut(text string) string {
	body, _ := cutText(text, maxText)
	return body
}

// toolResultText flattens a tool_result's content to the text it holds.
func toolResultText(raw json.RawMessage) string {
	var text string
	if json.Unmarshal(raw, &text) == nil {
		return text
	}
	var blocks []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if json.Unmarshal(raw, &blocks) != nil {
		return string(raw)
	}
	var out strings.Builder
	for _, block := range blocks {
		if block.Text != "" {
			out.WriteString(block.Text)
			out.WriteString("\n")
		} else if block.Type != "" {
			fmt.Fprintf(&out, "[%s]\n", block.Type)
		}
	}
	return strings.TrimRight(out.String(), "\n")
}
