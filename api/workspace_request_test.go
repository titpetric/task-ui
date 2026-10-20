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
}
