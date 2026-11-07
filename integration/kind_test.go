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
	assert.Len(t, integration.Kinds(), 6)
	assert.NotNil(t, integration.KindOf(integration.Claude).Field("who"))
	openrouter := integration.KindOf(integration.OpenRouter)
	assert.NotNil(t, openrouter)
	assert.True(t, openrouter.Field("token").Secret)
	assert.Equal(t, "https://openrouter.ai/api", openrouter.Field("base_url").Default)
	assert.False(t, openrouter.Field("model").Toggle, "the API key sets every variable")
	assert.Equal(t, "token", openrouter.Field("auth_token").Source)
	assert.True(t, openrouter.Field("api_key").Empty)
	assert.Equal(t, "ANTHROPIC_API_KEY", openrouter.Field("api_key").Env, "the empty key goes out explicitly")
	assert.Equal(t, "ANTHROPIC_DEFAULT_FABLE_MODEL", openrouter.Field("default_fable_model").Env)
	assert.Equal(t, "ANTHROPIC_DEFAULT_OPUS_MODEL", openrouter.Field("default_opus_model").Env)
	assert.Equal(t, "ANTHROPIC_DEFAULT_SONNET_MODEL", openrouter.Field("default_sonnet_model").Env)
	assert.Equal(t, "ANTHROPIC_DEFAULT_HAIKU_MODEL", openrouter.Field("default_haiku_model").Env)
	assert.Equal(t, "CLAUDE_CODE_SUBAGENT_MODEL", openrouter.Field("subagent_model").Env)
	assert.Equal(t, "CLAUDE_CODE_DISABLE_UNKNOWN_MODEL_WINDOW_ENFORCEMENT", openrouter.Field("unknown_model_window").Env)
	assert.Nil(t, openrouter.ToggleField("model.enabled"))
	assert.Nil(t, openrouter.ToggleField("token.enabled"), "a secret has no switch")
	assert.Nil(t, openrouter.ToggleField("model"))
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
	assert.Equal(t, "API key", integration.KindOf(integration.OpenRouter).Credential())
	assert.Equal(t, "token", integration.KindOf(integration.Forgejo).Credential())
	assert.Equal(t, "token", integration.KindOf(integration.Beszel).Credential())
	assert.Equal(t, "", (&integration.Kind{Fields: []integration.Field{{Name: "secret", Secret: true, Generated: true}}}).Credential())
	assert.Contains(t, integration.KindOf(integration.Claude).Steps[0], "`claude setup-token`")
	forgejo := integration.KindOf(integration.Forgejo)
	assert.Contains(t, forgejo.Steps[1], "`read:user`")
	assert.Contains(t, forgejo.Field("token").Help, "read:user")
	assert.Contains(t, forgejo.Steps[4], "`<login>/<login>`")
}

// Spotify keeps its credential out of every run: no setting sets a variable.
func TestKind_Spotify(t *testing.T) {
	t.Parallel()
	kind := integration.KindOf(integration.Spotify)
	assert.NotNil(t, kind)
	assert.True(t, kind.Field("client_secret").Secret)
	assert.False(t, kind.Field("client_id").Secret)
	assert.Equal(t, "client secret", kind.Credential())
	assert.Equal(t, []string{"read", "write"}, kind.Field("scopes").Options, "read is the default")
	for _, field := range kind.Fields {
		assert.Equal(t, "", field.Env, field.Name)
	}
	assert.False(t, integration.Configured(*kind, map[string]string{"client_id": "x"}))
	assert.True(t, integration.Configured(*kind, map[string]string{"client_id": "x", "client_secret": "y"}))
}

// Speech to text names two paths on this machine and no credential, so
// nothing of it is a secret and nothing reaches a run as a variable.
func TestKind_STT(t *testing.T) {
	t.Parallel()
	kind := integration.KindOf(integration.STT)
	assert.NotNil(t, kind)
	assert.Equal(t, "", kind.Credential())
	for _, field := range kind.Fields {
		assert.False(t, field.Secret, field.Name)
		assert.Equal(t, "", field.Env, field.Name)
	}
	assert.True(t, kind.Field("threads").Optional)
	assert.False(t, integration.Configured(*kind, map[string]string{"bin": "/x/whisper-cli"}))
	assert.True(t, integration.Configured(*kind, map[string]string{"bin": "/x/whisper-cli", "model": "/x/m.bin"}))
}
