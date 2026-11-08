package api

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
)

// A commit is named in full, or by its first seven characters or more.
func TestHasVersion(t *testing.T) {
	t.Parallel()
	versions := []ArtifactVersion{{Commit: "9f1c2d3e4b5a69788796a5b4c3d2e1f0a9b8c7d6"}}
	assert.True(t, hasVersion(versions, "9f1c2d3e4b5a69788796a5b4c3d2e1f0a9b8c7d6"))
	assert.True(t, hasVersion(versions, "9f1c2d3"))
	assert.False(t, hasVersion(versions, "9f1c2d"))
	assert.False(t, hasVersion(versions, "0000000"))
	assert.False(t, hasVersion(nil, "9f1c2d3"))
}
