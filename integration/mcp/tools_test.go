package mcp_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/mcp"
)

// list_jobs names each lint finding of a pipeline atkins refuses, which
// otherwise lists as a folder with no jobs.
func TestListJobs_Errors(t *testing.T) {
	t.Parallel()
	server := mcp.New(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"folders":[{"id":"app","title":"app","jobs":[],"error":"atkins.yml: pipeline 'app' has errors",
			"errors":[{"path":"atkins.yml","job":"build","detail":"step references task 'go:fmt'"},{"path":"atkins.yml","detail":"bare"}]}]}`))
	}), "")
	data, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": 1, "method": "tools/call", "params": map[string]any{"name": "list_jobs"}})
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(data)))
	body := rec.Body.String()
	assert.Contains(t, body, `! atkins.yml job build: step references task 'go:fmt'`)
	assert.Contains(t, body, `! atkins.yml: bare`)
}
