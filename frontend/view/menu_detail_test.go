package view_test

import (
	"testing"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
)

// A detail takes a title and mono; another option is an error.
func TestNewMenuDetail(t *testing.T) {
	t.Parallel()

	d, err := view.NewMenuDetail("Fingerprint", "SHA256:abc", "mono", "", "title", "the whole line")
	assert.NoError(t, err)
	assert.Equal(t, view.MenuDetail{Label: "Fingerprint", Value: "SHA256:abc", Title: "the whole line", Mono: true}, d)

	_, err = view.NewMenuDetail("Created", "now", "bold", "")
	assert.Error(t, err)
	_, err = view.NewMenuDetail("Created", "now", "title")
	assert.Error(t, err)
}
