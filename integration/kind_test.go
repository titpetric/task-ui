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
	assert.Len(t, integration.Kinds(), 3)
	assert.NotNil(t, integration.KindOf(integration.Claude).Field("who"))
	forgejo := integration.KindOf(integration.Forgejo)
	assert.Equal(t, []string{"ssh", "https"}, forgejo.Field("clone").Options)
	assert.True(t, forgejo.Field("secret").Generated)
	assert.True(t, forgejo.Field("token").Secret)
	assert.Equal(t, "/api/webhooks/forgejo", integration.WebhookPath)
	assert.False(t, integration.Configured(*forgejo, map[string]string{"url": "https://git.cdn.si"}))
	assert.True(t, integration.Configured(*forgejo, map[string]string{"url": "https://git.cdn.si", "token": "t"}))

	assert.False(t, integration.Configured(*kind, map[string]string{"url": "http://beszel:8090"}))
	assert.True(t, integration.Configured(*kind, map[string]string{"url": "u", "token": "t"}))
}
