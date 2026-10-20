package storage_test

import (
	"context"
	"testing"

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
	assert.NoError(t, passwords.SetPassword(ctx, user.ID, "$2second"))
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
