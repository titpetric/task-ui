//go:build integration

package mcp_test

import (
	"testing"

	"github.com/titpetric/task-ui/config"
	"github.com/titpetric/task-ui/internal/assert"
)

// resources/list and resources/read return context.md as a resource.
func TestServer_resources(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	list := f.rpc(t, "guest", "resources/list", nil)
	assert.Len(t, list.Result.Resources, 1)
	assert.Equal(t, "taskui://context", list.Result.Resources[0].URI)

	read := f.rpc(t, "guest", "resources/read", map[string]any{"uri": "taskui://context"})
	assert.Len(t, read.Result.Contents, 1)
	assert.Equal(t, "text/markdown", read.Result.Contents[0].MimeType)
	assert.Equal(t, config.Context(), read.Result.Contents[0].Text)

	assert.Equal(t, -32602, f.rpc(t, "guest", "resources/read", map[string]any{"uri": "taskui://nothing"}).Error.Code)
}
