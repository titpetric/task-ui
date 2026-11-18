package mcp_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/titpetric/task-ui/integration/mcp"
	"github.com/titpetric/task-ui/internal/assert"
)

// callThread asks one of the thread tools of a handler and returns the
// text the tool answered with.
func callThread(t *testing.T, handler http.HandlerFunc, name string, args map[string]any) string {
	t.Helper()
	server := mcp.New(handler, "")
	data, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": 1, "method": "tools/call", "params": map[string]any{"name": name, "arguments": args}})
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(data)))
	return rec.Body.String()
}

// list_threads names each thread with its state, place, count and agent.
func TestListThreads_ShowsThreads(t *testing.T) {
	t.Parallel()
	body := callThread(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/threads", r.URL.Path)
		assert.Equal(t, "1", r.URL.Query().Get("threads"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"threads":[{"id":"fx1","user_name":"ann","state":"working","dir":"malta","agent":"claude","created_at":"2026-11-16T10:00:00Z","messages":2,"body":"why did the build fail"}]}`))
	}, "list_threads", nil)
	assert.Contains(t, body, "thread fx1 by ann, working, malta")
	assert.Contains(t, body, "2 messages, agent claude")
	assert.Contains(t, body, "why did the build fail")
}

// read_thread shows the question, the work of each run and the replies.
func TestReadThread_ShowsConversation(t *testing.T) {
	t.Parallel()
	body := callThread(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/threads/fx1", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"thread":{"id":"fx1","user_name":"ann","state":"done","dir":"malta","body":"why"},
			"agent":"claude",
			"messages":[{"id":"fx1","user_name":"ann","body":"why"},{"id":"fx2","user_name":"claude","body":"because"}],
			"runs":[{"question":"fx1","run":"r1","answer":"fx2","events":[{"kind":"text","body":"looking"}]}]}`))
	}, "read_thread", map[string]any{"id": "fx1"})
	assert.Contains(t, body, "run r1:")
	assert.Contains(t, body, "text: looking")
	assert.Contains(t, body, "reply by claude")
}
