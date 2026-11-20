package api

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
)

// A filter's * is no filter, as its empty value is.
func TestAnyMatch(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "", anyMatch("*"))
	assert.Equal(t, "", anyMatch(""))
	assert.Equal(t, "open", anyMatch("open"))
}
