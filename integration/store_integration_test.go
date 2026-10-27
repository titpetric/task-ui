//go:build integration

package integration_test

import (
	"context"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/titpetric/task-ui/integration"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/secret"
	"github.com/titpetric/task-ui/sshkey"
	"github.com/titpetric/task-ui/storage"
)

func newStore(t *testing.T) *integration.Store {
	t.Helper()
	db, err := storage.OpenAndMigrate(context.Background(), ":memory:")
	assert.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	keys := &sshkey.Store{Dir: t.TempDir()}
	_, err = keys.Generate("ed25519", "test")
	assert.NoError(t, err)
	return &integration.Store{Secrets: secret.New(storage.NewSecretStorage(db), secret.NewVault(secret.DirKeys{Store: keys}))}
}

// Settings save encrypted, an empty secret keeps its value, an empty plain
// setting clears, and the Claude token reaches only the folders named.
func TestStore(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	s := newStore(t)

	assert.NoError(t, s.Save(ctx, integration.Claude, map[string]string{"token": "sk-1"}))
	values, err := s.Values(ctx, integration.Claude)
	assert.NoError(t, err)
	assert.Equal(t, "sk-1", values["token"])

	rows, _ := s.Secrets.Secrets.List(ctx)
	assert.Len(t, rows, 1)
	for _, row := range rows {
		assert.True(t, secret.Encrypted(row.Content), row.Name)
		assert.True(t, integration.Internal(row.Scope))
	}

	env, err := s.Env(ctx, "ops", "runner")
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"CLAUDE_CODE_OAUTH_TOKEN": "sk-1"}, env, "naming nobody is everyone")
	secrets, _ := s.Secrets.For(ctx, "tools")
	assert.Len(t, secrets, 0, "no run sees an integration as a secret")

	// Who keeps the token to the users and roles it names.
	assert.NoError(t, s.Save(ctx, integration.Claude, map[string]string{"who": "admin, claude"}))
	for _, c := range []struct {
		user, role string
		gets       bool
	}{{"black", "admin", true}, {"claude", "runner", true}, {"ops", "runner", false}} {
		env, _ = s.Env(ctx, c.user, c.role)
		assert.Equal(t, c.gets, env["CLAUDE_CODE_OAUTH_TOKEN"] == "sk-1", c.user)
	}

	assert.NoError(t, s.Save(ctx, integration.Claude, map[string]string{"token": "", "who": ""}))
	values, _ = s.Values(ctx, integration.Claude)
	assert.Equal(t, "sk-1", values["token"], "an empty secret keeps its value")
	assert.Equal(t, "", values["who"], "an empty setting clears")

	assert.ErrorIs(t, s.Save(ctx, "nothing", nil), integration.ErrUnknown)
	assert.ErrorIs(t, s.Save(ctx, integration.Claude, map[string]string{"colour": "red"}), integration.ErrUnknown)
	_, err = s.Values(ctx, "nothing")
	assert.ErrorIs(t, err, integration.ErrUnknown)
	assert.ErrorIs(t, s.Delete(ctx, "nothing"), integration.ErrUnknown)

	// Forgejo: an option takes only its values, and the webhook secret
	// comes into being on the first save and stays.
	assert.ErrorIs(t, s.Save(ctx, integration.Forgejo, map[string]string{"url": "https://git.example", "token": "t", "clone": "ftp"}), integration.ErrUnknown)
	assert.NoError(t, s.Save(ctx, integration.Forgejo, map[string]string{"url": "https://git.example", "token": "t"}))
	values, _ = s.Values(ctx, integration.Forgejo)
	assert.Len(t, values["secret"], 64, "a generated secret")
	assert.Equal(t, "ssh", values["clone"], "an empty option takes the first")
	made := values["secret"]
	assert.NoError(t, s.Save(ctx, integration.Forgejo, map[string]string{"url": "https://git.example", "clone": "https"}))
	values, _ = s.Values(ctx, integration.Forgejo)
	assert.Equal(t, made, values["secret"], "kept on the next save")
	assert.Equal(t, "https", values["clone"])
	assert.NoError(t, s.Save(ctx, integration.Forgejo, map[string]string{"secret": "mine"}))
	values, _ = s.Values(ctx, integration.Forgejo)
	assert.Equal(t, "mine", values["secret"], "a given secret replaces it")

	assert.NoError(t, s.Delete(ctx, integration.Claude))
	values, _ = s.Values(ctx, integration.Claude)
	assert.Len(t, values, 0)
	env, _ = s.Env(ctx, "ops", "runner")
	assert.Len(t, env, 0)
	assert.False(t, integration.Internal("tools"))

	// OpenRouter sets a run's environment from its key and the defaults of
	// its variables, and its auth token stands in for the Claude token.
	assert.NoError(t, s.Save(ctx, integration.Claude, map[string]string{"token": "sk-1"}))
	assert.NoError(t, s.Save(ctx, integration.OpenRouter, map[string]string{"token": "sk-or-1"}))
	env, _ = s.Env(ctx, "ops", "runner")
	assert.Equal(t, "sk-or-1", env["OPENROUTER_API_KEY"])
	assert.Equal(t, "sk-or-1", env["ANTHROPIC_AUTH_TOKEN"])
	assert.Equal(t, "https://openrouter.ai/api", env["ANTHROPIC_BASE_URL"])
	assert.Equal(t, "deepseek/deepseek-v4.1-flash-1017", env["ANTHROPIC_MODEL"])
	assert.Equal(t, "1000000", env["CLAUDE_CODE_MAX_CONTEXT_TOKENS"])
	assert.Equal(t, "400000", env["CLAUDE_CODE_AUTO_COMPACT_WINDOW"])
	assert.Equal(t, "1", env["CLAUDE_CODE_ENABLE_GATEWAY_MODEL_DISCOVERY"])
	assert.Equal(t, "1", env["CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC"])
	assert.Equal(t, "1", env["CLAUDE_CODE_ENABLE_TODO_TOOLS"])
	assert.Equal(t, "", env["ANTHROPIC_API_KEY"])
	assert.Equal(t, "", env["CLAUDE_CODE_OAUTH_TOKEN"], "OpenRouter stands in for the Claude token")

	// A switch leaves its variable out and the value stays for the next
	// switch on.
	assert.NoError(t, s.Save(ctx, integration.OpenRouter, map[string]string{"model": "other/model", "model.enabled": "0"}))
	values, _ = s.Values(ctx, integration.OpenRouter)
	assert.Equal(t, "other/model", values["model"])
	assert.Equal(t, "0", values["model_enabled"])
	env, _ = s.Env(ctx, "ops", "runner")
	_, ok := env["ANTHROPIC_MODEL"]
	assert.False(t, ok)
	assert.NoError(t, s.Save(ctx, integration.OpenRouter, map[string]string{"model": "other/model", "model.enabled": "1"}))
	env, _ = s.Env(ctx, "ops", "runner")
	assert.Equal(t, "other/model", env["ANTHROPIC_MODEL"])

	// Without the auth token, the Claude integration signs in again.
	assert.NoError(t, s.Save(ctx, integration.OpenRouter, map[string]string{"token": "sk-or-1", "auth_token.enabled": "0"}))
	env, _ = s.Env(ctx, "ops", "runner")
	assert.Equal(t, "", env["ANTHROPIC_AUTH_TOKEN"])
	assert.Equal(t, "sk-1", env["CLAUDE_CODE_OAUTH_TOKEN"])

	assert.ErrorIs(t, s.Save(ctx, integration.OpenRouter, map[string]string{"model.enabled": "yes"}), integration.ErrUnknown)
}
