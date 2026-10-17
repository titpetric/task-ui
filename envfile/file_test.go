package envfile_test

import (
	"encoding/json"
	"testing"

	"github.com/titpetric/task-ui/envfile"
	"github.com/titpetric/task-ui/internal/assert"
)

func TestFile_State(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "managed", envfile.File{Managed: true, Exists: true}.State())
	assert.Equal(t, "in folder", envfile.File{Exists: true}.State())
	assert.Equal(t, "missing", envfile.File{}.State())
}

func TestFile(t *testing.T) {
	t.Parallel()
	data, err := json.Marshal(envfile.File{Folder: "infra", Path: ".env", Content: "secret", Keys: []string{"A"}})
	assert.NoError(t, err)
	assert.NotContains(t, string(data), "secret")
	var back envfile.File
	assert.NoError(t, json.Unmarshal(data, &back))
	assert.Equal(t, ".env", back.Path)
}
