package api_test

import (
	"encoding/json"
	"testing"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
)

func TestWorkspaceRequest(t *testing.T) {
	t.Parallel()
	var req api.WorkspaceRequest
	assert.NoError(t, json.Unmarshal([]byte(`{"url":"file:///w.git","folder":"malta","title":"Malta","description":"d","branch":"main"}`), &req))
	assert.Equal(t, api.WorkspaceRequest{URL: "file:///w.git", Folder: "malta", Title: "Malta", Description: "d", Branch: "main"}, req)

	var kept api.WorkspaceRequest
	assert.NoError(t, json.Unmarshal([]byte(`{"title":"Malta"}`), &kept))
	assert.Nil(t, kept.Recursive)
	assert.Nil(t, kept.Sort)
	assert.Nil(t, kept.SidebarLimit)

	var limited api.WorkspaceRequest
	assert.NoError(t, json.Unmarshal([]byte(`{"sidebar_limit":3}`), &limited))
	assert.Equal(t, int64(3), *limited.SidebarLimit)
}

// A home request names the repository and, when it has one, the branch.
func TestWorkspaceHomeRequest(t *testing.T) {
	t.Parallel()
	var req api.WorkspaceHomeRequest
	assert.NoError(t, json.Unmarshal([]byte(`{"url":"file:///w.git","branch":"main"}`), &req))
	assert.Equal(t, api.WorkspaceHomeRequest{URL: "file:///w.git", Branch: "main"}, req)
}
