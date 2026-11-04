//go:build integration

package storage_test

import (
	"context"
	"testing"
	"time"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

func TestUserAuthPasswordStorage(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := newDB(t)
	users, passwords := storage.NewUserStorage(db), storage.NewUserAuthPasswordStorage(db)
	user := &model.User{Name: "ops", Role: model.RoleRunner}
	assert.NoError(t, users.Create(ctx, user))

	hash, err := passwords.PasswordHash(ctx, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, "", hash)

	assert.NoError(t, passwords.SetPassword(ctx, user.ID, "$2first"))
	// A new password ends the sessions the old one opened.
	sessions := storage.NewUserSessionStorage(db)
	_, err = sessions.Create(ctx, user.ID, "h1", time.Now().Add(time.Hour))
	assert.NoError(t, err)
	assert.NoError(t, passwords.SetPassword(ctx, user.ID, "$2second"))
	_, err = sessions.Find(ctx, "h1")
	assert.ErrorIs(t, err, storage.ErrNotFound)
	hash, err = passwords.PasswordHash(ctx, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, "$2second", hash)

	assert.NoError(t, passwords.SetBackups(ctx, user.ID, []string{"$2b1", "$2b2"}))
	list, err := passwords.List(ctx, user.ID)
	assert.NoError(t, err)
	assert.Len(t, list, 3)
	assert.Equal(t, model.AuthPassword, list[0].Kind)
	assert.Equal(t, model.AuthBackup, list[1].Kind)

	// A backup signs in once; the password never counts as "used".
	ok, err := passwords.Use(ctx, list[1].ID)
	assert.NoError(t, err)
	assert.True(t, ok)
	ok, err = passwords.Use(ctx, list[1].ID)
	assert.NoError(t, err)
	assert.False(t, ok)
	ok, err = passwords.Use(ctx, list[0].ID)
	assert.NoError(t, err)
	assert.False(t, ok)

	// New backups replace the old, used or not.
	assert.NoError(t, passwords.SetBackups(ctx, user.ID, []string{"$2b3"}))
	list, _ = passwords.List(ctx, user.ID)
	assert.Len(t, list, 2)
	assert.True(t, list[1].Usable())

	assert.Error(t, passwords.SetPassword(ctx, "", "$2x"))
	assert.Error(t, passwords.SetPassword(ctx, user.ID, ""))

	assert.NoError(t, passwords.DeleteByUser(ctx, user.ID))
	list, _ = passwords.List(ctx, user.ID)
	assert.Len(t, list, 0)
}

// Every method hands back what the driver said. PasswordHash tells a
// user with no password from a pool that cannot answer: the first is no
// hash and no error, the second is the error.
func TestUserAuthPasswordStorage_Closed(t *testing.T) {
	t.Parallel()
	passwords := storage.NewUserAuthPasswordStorage(closedDB(t))
	ctx := context.Background()

	_, err := passwords.List(ctx, "u1")
	assert.ErrorContains(t, err, errClosed)
	_, err = passwords.PasswordHash(ctx, "u1")
	assert.ErrorContains(t, err, errClosed)
	assert.ErrorContains(t, passwords.SetPassword(ctx, "u1", "hash"), errClosed)
	assert.ErrorContains(t, passwords.SetBackups(ctx, "u1", []string{"hash"}), errClosed)
	_, err = passwords.Use(ctx, "p1")
	assert.ErrorContains(t, err, errClosed)
	assert.ErrorContains(t, passwords.DeleteByUser(ctx, "u1"), errClosed)
}

// A replace clears the user's passwords of the kind and then writes the
// hashes, and stops at the write that failed. A password set leaves the
// user's sessions, which is its own statement and its own error.
func TestUserAuthPasswordStorage_Writes(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	refused := newDB(t)
	refuse(t, refused, model.UserAuthPasswordTable, "INSERT")
	assert.ErrorContains(t, storage.NewUserAuthPasswordStorage(refused).SetBackups(ctx, "u1", []string{"hash"}), errRefused)

	sessions := newDB(t)
	refuse(t, sessions, model.UserSessionTable, "DELETE")
	users := storage.NewUserStorage(sessions)
	user := &model.User{Name: "alice", Role: model.RoleRunner}
	assert.NoError(t, users.Create(ctx, user))
	open := storage.NewUserSessionStorage(sessions)
	_, err := open.Create(ctx, user.ID, "cookiehash", time.Now().Add(time.Hour))
	assert.NoError(t, err)
	assert.ErrorContains(t, storage.NewUserAuthPasswordStorage(sessions).SetPassword(ctx, user.ID, "hash"), errRefused)
}

// A password needs a user, and a hash for every password it names.
func TestUserAuthPasswordStorage_Refuses(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	passwords := storage.NewUserAuthPasswordStorage(newDB(t))

	assert.ErrorContains(t, passwords.SetPassword(ctx, "", "hash"), "needs a user")
	assert.ErrorContains(t, passwords.SetBackups(ctx, "u1", []string{""}), "needs a hash")
}
