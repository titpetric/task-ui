//go:build integration

package app_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/app"
	"github.com/titpetric/task-ui/internal/assert"
)

// TestNew builds a whole instance over a temporary data directory and asks it
// for the three paths that answer at the root.
//
// Nothing else builds one: wire mounts the front end, the byte streams, the
// API and the assets on one router, and chi panics on a second Mount of one
// pattern. A unit test over a handler proves nothing about that.
func TestNew(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "workspaces")
	assert.NoError(t, os.MkdirAll(root, 0o755))
	t.Setenv("TASK_UI_DATA", filepath.Join(base, "data"))
	t.Setenv("TASK_UI_SSH_DIR", filepath.Join(base, "keys"))
	t.Setenv("TASK_UI_WORKSPACE", root)
	t.Setenv("TASK_UI_DISCOVER", "false")

	cfg, err := app.NewConfigAt(base, nil)
	assert.NoError(t, err)

	instance, err := app.New(context.Background(), cfg, "test")
	assert.NoError(t, err)
	if instance == nil {
		return
	}
	defer instance.Close()

	server := httptest.NewServer(instance.Handler())
	defer server.Close()

	client := server.Client()
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }

	tests := []struct {
		path string
		want int
	}{
		{path: "/healthz", want: http.StatusOK},
		{path: "/login", want: http.StatusOK},
		{path: "/", want: http.StatusSeeOther},
	}
	for _, test := range tests {
		answer, err := client.Get(server.URL + test.path)
		assert.NoError(t, err)
		if err != nil {
			continue
		}
		assert.Equal(t, test.want, answer.StatusCode, test.path)
		answer.Body.Close()
	}
}
