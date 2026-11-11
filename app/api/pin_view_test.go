package api_test

import (
	"testing"

	"github.com/titpetric/task-ui/app/api"
	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/internal/assert"
)

// A pin shows its args as a list, none for args it cannot read.
func TestPinView(t *testing.T) {
	t.Parallel()
	assert.Equal(t, []string{"a"}, api.NewPinView(model.Pin{Args: `["a"]`}).Args)
	assert.Equal(t, []string{}, api.NewPinView(model.Pin{Args: "nope"}).Args)
}
