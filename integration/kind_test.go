package integration_test

import (
	"encoding/json"
	"strings"
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

// Every kind carries the steps to its credential, each one with words in
// it, in the JSON the API serves too; the heading names the credential
// as its field's label, lower case unless it starts with an acronym.
func TestKind_Steps(t *testing.T) {
	t.Parallel()
	for _, kind := range integration.Kinds() {
		assert.True(t, len(kind.Steps) >= 3, kind.ID)
		for i, step := range kind.Steps {
			assert.True(t, strings.TrimSpace(step) != "", kind.ID, i)
		}
		raw, err := json.Marshal(kind)
		assert.NoError(t, err)
		assert.Contains(t, string(raw), `"steps":["`)
	}
	assert.Equal(t, "OAuth token", integration.KindOf(integration.Claude).Credential())
	assert.Equal(t, "token", integration.KindOf(integration.Forgejo).Credential())
	assert.Equal(t, "token", integration.KindOf(integration.Beszel).Credential())
	assert.Equal(t, "", (&integration.Kind{Fields: []integration.Field{{Name: "secret", Secret: true, Generated: true}}}).Credential())
	assert.Contains(t, integration.KindOf(integration.Claude).Steps[0], "`claude setup-token`")
}
