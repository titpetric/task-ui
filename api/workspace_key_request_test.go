package api_test

import (
	"encoding/json"
	"testing"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
)

// A request names the key and nothing about who gets it: links of an
// older client are read and dropped, because the workspace it is added
// to is the answer.
func TestWorkspaceKeyRequest(t *testing.T) {
	t.Parallel()
	var req api.WorkspaceKeyRequest
	assert.NoError(t, json.Unmarshal([]byte(`{"type":"ed25519","name":"deploy","comment":"c","policies":["admin"],"users":["alice"]}`), &req))
	assert.Equal(t, api.WorkspaceKeyRequest{Type: "ed25519", Name: "deploy", Comment: "c"}, req)
	assert.NoError(t, json.Unmarshal([]byte(`{"encrypt":true}`), &req))
	assert.True(t, req.Encrypt)
}
