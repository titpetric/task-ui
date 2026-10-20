package view_test

import (
	"testing"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/integration"
	"github.com/titpetric/task-ui/internal/assert"
)

// A form shows the plain settings, and of a secret only whether it has a value.
func TestIntegration(t *testing.T) {
	t.Parallel()
	kind := *integration.KindOf(integration.Beszel)
	var got view.Integration = view.NewIntegration(kind, map[string]string{"url": "http://beszel:8090", "token": "t1"})
	assert.True(t, got.Configured)
	assert.Equal(t, "http://beszel:8090", got.Values["url"])
	assert.True(t, got.Set["token"])
	_, leaked := got.Values["token"]
	assert.False(t, leaked)
	assert.False(t, view.NewIntegration(kind, nil).Configured)
}
