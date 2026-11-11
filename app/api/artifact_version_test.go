package api_test

import (
	"encoding/json"
	"testing"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
)

// A version names its commit and the run that wrote it.
func TestArtifactVersion(t *testing.T) {
	t.Parallel()
	data, err := json.Marshal(api.ArtifactVersion{Commit: "abc1234", Folder: "infra", Job: "build", RunID: "r1", ByteSize: 3})
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"commit":"abc1234"`)
	assert.Contains(t, string(data), `"run_id":"r1"`)
}
