package phpfront

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/markdown"
)

// The policy is on every answer under the mount, whatever the handler under it
// did: a page, an asset and a refusal all carry it, because it is set before
// the handler runs and not by the handler.
func TestPolicy(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		next http.Handler
	}{{
		name: "a page",
		next: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("<html></html>"))
		}),
	}, {
		name: "a refusal written before any page ran",
		next: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "no", http.StatusForbidden)
		}),
	}, {
		name: "a handler that writes nothing at all",
		next: http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}),
	}} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			w := httptest.NewRecorder()
			Policy(tc.next).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
			assert.Equal(t, markdown.CSP, w.Header().Get("Content-Security-Policy"))
		})
	}
}

// It is the policy the Go pages carry and not a second one: the renderer owns
// the line, so a change to what markdown allows reaches both front ends.
func TestPolicy_Markdown(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	Policy(http.NotFoundHandler()).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, markdown.CSP, w.Header().Get("Content-Security-Policy"))
	assert.Contains(t, markdown.CSP, "default-src", "a policy, not an empty string")
	assert.Equal(t, http.StatusNotFound, w.Code, "and the handler under it answered")
}
