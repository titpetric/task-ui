//go:build integration

package spotify_test

import (
	"context"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/titpetric/task-ui/app"
	"github.com/titpetric/task-ui/integration"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/feature/secret"
	"github.com/titpetric/task-ui/integration/spotify"
	"github.com/titpetric/task-ui/integration/ssh/sshkey"
	"github.com/titpetric/task-ui/storage"
)

// The tokens sit encrypted in the user's own scope: a second user has
// none, no run's variables or secrets page hold them, and Delete takes
// every one out.
func TestSecretTokens(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, err := storage.OpenAndMigrate(ctx, ":memory:", app.Migrations())
	assert.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	keys := &sshkey.Store{Dir: t.TempDir()}
	_, err = keys.Generate("ed25519", "task-ui")
	assert.NoError(t, err)
	rows := storage.NewSecretStorage(db)
	store := secret.New(rows, secret.NewVault(secret.DirKeys{Store: keys}))
	tokens := spotify.SecretTokens{Secrets: store}

	link := spotify.Link{Token: spotify.Token{Access: "access-secret-1", Refresh: "refresh-secret-1", Expires: time.Unix(1800000000, 0), Scope: "a b"}, Name: "Tit"}
	assert.NoError(t, tokens.Save(ctx, "tit", link))
	got, err := tokens.Load(ctx, "tit")
	assert.NoError(t, err)
	assert.Equal(t, link, *got)
	other, err := tokens.Load(ctx, "eve")
	assert.NoError(t, err)
	assert.Nil(t, other)

	// Stored encrypted, in a scope no folder reaches and the page hides.
	raw, err := rows.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, raw, 6)
	for _, row := range raw {
		assert.Equal(t, "user:tit", row.Scope)
		assert.True(t, integration.Internal(row.Scope))
		assert.NotContains(t, row.Content, "secret-1")
	}
	reaching, err := store.For(ctx, "malta/apps")
	assert.NoError(t, err)
	assert.Len(t, reaching, 0)

	link.Broken = true
	assert.NoError(t, tokens.Save(ctx, "tit", link))
	got, _ = tokens.Load(ctx, "tit")
	assert.True(t, got.Broken)

	assert.NoError(t, tokens.Delete(ctx, "tit"))
	raw, _ = rows.List(ctx)
	assert.Len(t, raw, 0)
	got, _ = tokens.Load(ctx, "tit")
	assert.Nil(t, got)
}
