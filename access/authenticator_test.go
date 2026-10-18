package access_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

func TestAuthenticator_Authenticate(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, err := storage.OpenAndMigrate(ctx, ":memory:")
	assert.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	users, emails, passwords := storage.NewUserStorage(db), storage.NewUserEmailStorage(db), storage.NewUserAuthPasswordStorage(db)
	auth := access.NewAuthenticator(users, emails, passwords)

	user := &model.User{Name: "black", Role: model.RoleAdmin}
	assert.NoError(t, users.Create(ctx, user))
	hash, err := access.HashPassword("default123")
	assert.NoError(t, err)
	assert.NoError(t, passwords.SetPassword(ctx, user.ID, hash))
	assert.NoError(t, emails.Add(ctx, user.ID, "black@example.com"))
	assert.NoError(t, emails.Add(ctx, user.ID, "ops@example.com"))
	codes, hashes, err := access.BackupPasswords(2)
	assert.NoError(t, err)
	assert.NoError(t, passwords.SetBackups(ctx, user.ID, hashes))

	// The name or any of the addresses, with the password.
	for _, login := range []string{"black", " black ", "black@example.com", "OPS@example.com"} {
		got, err := auth.Authenticate(ctx, login, "default123")
		assert.NoError(t, err, login)
		assert.Equal(t, user.ID, got.ID)
	}

	// A backup password signs in once; the other stays good.
	_, err = auth.Authenticate(ctx, "black", codes[0])
	assert.NoError(t, err)
	_, err = auth.Authenticate(ctx, "black", codes[0])
	assert.ErrorIs(t, err, access.ErrWrongCredentials)
	_, err = auth.Authenticate(ctx, "black@example.com", codes[1])
	assert.NoError(t, err)

	// Every failure is the same failure.
	for _, c := range [][2]string{{"black", "wrong"}, {"nobody", "default123"}, {"nobody@example.com", "default123"}, {"", "default123"}} {
		_, err := auth.Authenticate(ctx, c[0], c[1])
		assert.ErrorIs(t, err, access.ErrWrongCredentials, c[0])
	}

	// A user with nothing to sign in with cannot.
	bare := &model.User{Name: "bare", Role: model.RoleViewer}
	assert.NoError(t, users.Create(ctx, bare))
	_, err = auth.Authenticate(ctx, "bare", "")
	assert.ErrorIs(t, err, access.ErrWrongCredentials)
}
