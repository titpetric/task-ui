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

// A tool says what the API answered when it is not a success: its
// message, or the body when it sent no JSON.
func TestClient_Errors(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		`{"error":"no such folder"}`: "no such folder (404)",
		"plain trouble\n":            "plain trouble (404)",
	}
	for body, want := range cases {
		server := mcp.New(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, body, http.StatusNotFound)
		}), "")
		data, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": 1, "method": "tools/call", "params": map[string]any{"name": "list_jobs"}})
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(data)))
		assert.Contains(t, rec.Body.String(), `"isError":true`)
		assert.Contains(t, rec.Body.String(), want)
	}
}
