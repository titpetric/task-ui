package integration_test

import (
	"context"
	"testing"

	_ "modernc.org/sqlite" // the driver storage opens

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
	return &integration.Store{Secrets: secret.New(storage.NewSecretStorage(db), secret.NewVault(keys))}
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

	assert.NoError(t, s.Delete(ctx, integration.Claude))
	values, _ = s.Values(ctx, integration.Claude)
	assert.Len(t, values, 0)
	env, _ = s.Env(ctx, "ops", "runner")
	assert.Len(t, env, 0)
	assert.False(t, integration.Internal("tools"))
}
