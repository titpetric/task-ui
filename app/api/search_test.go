package api

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
)

// A browse path joins its prefix and its child.
func TestJoinPath(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "docs", joinPath("", "docs"))
	assert.Equal(t, "src/acme/docs", joinPath("src/acme", "docs"))
}

// A folder is listed once.
func TestUniqueStrings(t *testing.T) {
	t.Parallel()
	assert.Equal(t, []string{"a", "b"}, uniqueStrings([]string{"a", "b", "a"}))
}
