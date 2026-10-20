package integration_test

import (
	"testing"

	"github.com/titpetric/task-ui/integration"
	"github.com/titpetric/task-ui/internal/assert"
)

func TestKinds(t *testing.T) {
	t.Parallel()
	kind := integration.KindOf(integration.Beszel)
	assert.NotNil(t, kind)
	assert.True(t, kind.Field("token").Secret)
	assert.Nil(t, kind.Field("colour"))
	assert.Nil(t, integration.KindOf("nothing"))
	assert.Len(t, integration.Kinds(), 2)
	assert.NotNil(t, integration.KindOf(integration.Claude).Field("who"))

	assert.False(t, integration.Configured(*kind, map[string]string{"url": "http://beszel:8090"}))
	assert.True(t, integration.Configured(*kind, map[string]string{"url": "u", "token": "t"}))
}
