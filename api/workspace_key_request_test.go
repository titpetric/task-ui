package api_test

import (
	"encoding/json"
	"testing"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
)

func TestWorkspaceKeyRequest(t *testing.T) {
	t.Parallel()
	var req api.WorkspaceKeyRequest
	assert.NoError(t, json.Unmarshal([]byte(`{"type":"ed25519","name":"deploy","comment":"c","policies":["admin"],"users":["alice"]}`), &req))
	assert.Equal(t, api.WorkspaceKeyRequest{Type: "ed25519", Name: "deploy", Comment: "c", Policies: []string{"admin"}, Users: []string{"alice"}}, req)
	assert.NoError(t, json.Unmarshal([]byte(`{"encrypt":true}`), &req))
	assert.True(t, req.Encrypt)
}
