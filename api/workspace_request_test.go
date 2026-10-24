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

// An update names the home project by folder; "" unsets it, and left
// out it stays.
func TestWorkspaceRequest_home(t *testing.T) {
	t.Parallel()
	var req api.WorkspaceRequest
	assert.NoError(t, json.Unmarshal([]byte(`{"title":"Malta","home_project":"malta/workspace"}`), &req))
	assert.Equal(t, "malta/workspace", *req.HomeProject)

	var unset api.WorkspaceRequest
	assert.NoError(t, json.Unmarshal([]byte(`{"home_project":""}`), &unset))
	assert.Equal(t, "", *unset.HomeProject)

	var kept api.WorkspaceRequest
	assert.NoError(t, json.Unmarshal([]byte(`{"title":"Malta"}`), &kept))
	assert.Nil(t, kept.HomeProject)

	var created api.WorkspaceRequest
	assert.NoError(t, json.Unmarshal([]byte(`{"folder":"malta","url":"file:///w.git","project":"workspace"}`), &created))
	assert.Equal(t, "workspace", created.Project)
}
