package api_test

import (
	"encoding/json"
	"testing"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
)

// An entry a user wrote names the user and no run; a directory has no
// writer.
func TestArtifactEntry_JSON(t *testing.T) {
	t.Parallel()
	data, err := json.Marshal(api.ArtifactEntry{Name: "todo.txt", Path: "todo.txt", ByteSize: 3, UserName: "ops"})
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"user_name":"ops"`)
	assert.NotContains(t, string(data), `"run_id"`)

	data, err = json.Marshal(api.ArtifactEntry{Name: "reports", Path: "reports", Dir: true})
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"dir":true`)
	assert.NotContains(t, string(data), `"folder"`)
}
