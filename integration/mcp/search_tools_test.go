package mcp_test

import (
	"net/http"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
)

// The search tools call the API's routes and print the hit, the outline and a
// job target with the words the agent needs to read it next.
func TestSearchTools(t *testing.T) {
	t.Parallel()
	var got []string
	api := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = append(got, r.Method+" "+r.URL.RequestURI())
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/search":
			_, _ = w.Write([]byte(`{"query":"daily","hits":[{"document":"d1","ordinal":1,"kind":"doc","path":"src/acme/README.md","folder":"acme","heading":"Install","breadcrumb":"Acme > Install","line":7,"snippet":"run the daily deploy","score":4}]}`))
		case "/search/browse":
			_, _ = w.Write([]byte(`{"path":"src/acme","folders":["src/acme/docs"],"documents":[{"id":"d2","kind":"job","path":"src/acme/atkins.yml#ci:daily","name":"ci:daily","title":"ci:daily - Run the gate.","dir":""}]}`))
		case "/search/documents/d1":
			_, _ = w.Write([]byte(`{"document":{"id":"d1","kind":"doc","path":"src/acme/README.md"},"sections":[{"ordinal":0,"level":1,"heading":"Acme"},{"ordinal":1,"level":2,"heading":"Install"}]}`))
		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":"not found"}`))
		}
	})

	out := callTool(api, "search_docs", map[string]any{"query": "daily"})
	assert.Contains(t, out, "src/acme/README.md")
	assert.Contains(t, out, "Install")
	assert.Contains(t, out, "document d1 section 1")
	assert.Contains(t, out, "run the daily deploy")

	out = callTool(api, "search_outline", map[string]any{"id": "d1"})
	assert.Contains(t, out, "Install")

	out = callTool(api, "search_browse", map[string]any{"path": "src/acme"})
	assert.Contains(t, out, "src/acme/docs/")
	assert.Contains(t, out, "ci:daily")

	assert.Equal(t, "GET /search?q=daily", got[0])
}
