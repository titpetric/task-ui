package view_test

import (
	"testing"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
)

func TestFolderHref(t *testing.T) {
	t.Parallel()
	workspaces := map[string]int{"acme": 0}
	assert.Equal(t, "/f/_/web", view.FolderHref("web", workspaces))
	assert.Equal(t, "/f/_/web/api", view.FolderHref("web/api", workspaces))
	assert.Equal(t, "/f/acme", view.FolderHref("acme", workspaces))
	assert.Equal(t, "/f/acme/api", view.FolderHref("acme/api", workspaces))
	assert.Equal(t, "/f/_/acmes", view.FolderHref("acmes", workspaces))
	assert.Equal(t, "/f/_", view.FolderHref("_", workspaces))
}
