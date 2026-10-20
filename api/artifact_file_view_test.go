package api_test

import (
	"encoding/json"
	"testing"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
)

// A file's view leaves out the content it has none of, and always lists
// the versions.
func TestArtifactFileView(t *testing.T) {
	t.Parallel()
	data, err := json.Marshal(api.ArtifactFileView{Path: "a.png", Kind: "image", Versions: []api.ArtifactVersion{}})
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"kind":"image"`)
	assert.Contains(t, string(data), `"versions":[]`)
	assert.NotContains(t, string(data), `"html"`)
}
