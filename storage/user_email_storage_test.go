package storage_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

func TestUserEmailStorage(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := newDB(t)
	users, emails := storage.NewUserStorage(db), storage.NewUserEmailStorage(db)
	ops := &model.User{Name: "ops", Role: model.RoleRunner}
	dev := &model.User{Name: "dev", Role: model.RoleRunner}
	assert.NoError(t, users.Create(ctx, ops))
	assert.NoError(t, users.Create(ctx, dev))

	assert.NoError(t, emails.Add(ctx, ops.ID, " Ops@Example.com "))
	assert.NoError(t, emails.Add(ctx, ops.ID, "oncall@example.com"))
	assert.Error(t, emails.Add(ctx, dev.ID, "ops@example.com"), "one user per address")
	for _, bad := range []string{"", "ops", "@example.com", "ops@", "o ps@example.com"} {
		assert.Error(t, emails.Add(ctx, ops.ID, bad), bad)
	}
	assert.Error(t, emails.Add(ctx, "", "x@example.com"))

	list, err := emails.List(ctx, ops.ID)
	assert.NoError(t, err)
	assert.Len(t, list, 2)
	assert.Equal(t, "ops@example.com", list[0].Email)

	id, err := emails.UserID(ctx, "OPS@example.com")
	assert.NoError(t, err)
	assert.Equal(t, ops.ID, id)
	_, err = emails.UserID(ctx, "nobody@example.com")
	assert.ErrorIs(t, err, storage.ErrNotFound)

	assert.NoError(t, emails.Delete(ctx, ops.ID, "OnCall@example.com"))
	list, _ = emails.List(ctx, ops.ID)
	assert.Len(t, list, 1)

	assert.NoError(t, emails.DeleteByUser(ctx, ops.ID))
	list, _ = emails.List(ctx, ops.ID)
	assert.Len(t, list, 0)
}

// Removing a user removes what it signs in with.
func TestUserStorage_DeleteCredentials(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := newDB(t)
	users, emails, passwords := storage.NewUserStorage(db), storage.NewUserEmailStorage(db), storage.NewUserAuthPasswordStorage(db)
	ops := &model.User{Name: "ops", Role: model.RoleRunner}
	assert.NoError(t, users.Create(ctx, ops))
	assert.NoError(t, emails.Add(ctx, ops.ID, "ops@example.com"))
	assert.NoError(t, passwords.SetPassword(ctx, ops.ID, "$2x"))

	assert.NoError(t, users.Delete(ctx, ops.ID))
	_, err := emails.UserID(ctx, "ops@example.com")
	assert.ErrorIs(t, err, storage.ErrNotFound)
	list, _ := passwords.List(ctx, ops.ID)
	assert.Len(t, list, 0)
}
