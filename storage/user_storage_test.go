package storage_test

import (
	"context"
	"testing"
	"time"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

func TestUserStorage_Create(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	users := storage.NewUserStorage(newDB(t))

	user := &model.User{Name: "ops", Role: model.RoleRunner}
	assert.NoError(t, users.Create(ctx, user))
	assert.True(t, model.ValidID(user.ID))
	assert.NotNil(t, user.CreatedAt)
	assert.Equal(t, "[]", user.Folders)

	assert.Error(t, users.Create(ctx, &model.User{Role: model.RoleAdmin}), "no name")
	assert.Error(t, users.Create(ctx, &model.User{Name: "x", Role: "root"}), "bad role")
	assert.Error(t, users.Create(ctx, &model.User{Name: "ops", Role: model.RoleAdmin}), "duplicate name")
}

func TestUserStorage_Get(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	users := storage.NewUserStorage(newDB(t))
	user := &model.User{Name: "admin", Role: model.RoleAdmin, Folders: `["web"]`}
	assert.NoError(t, users.Create(ctx, user))

	got, err := users.Get(ctx, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, "admin", got.Name)
	assert.Equal(t, []string{"web"}, got.FolderList())
	assert.NotNil(t, got.CreatedAt)

	_, err = users.Get(ctx, "missing")
	assert.ErrorIs(t, err, storage.ErrNotFound)
}

func TestUserStorage_GetByName(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	users := storage.NewUserStorage(newDB(t))
	assert.NoError(t, users.Create(ctx, &model.User{Name: "guest", Role: model.RoleViewer}))

	got, err := users.GetByName(ctx, "guest")
	assert.NoError(t, err)
	assert.Equal(t, model.RoleViewer, got.Role)
	_, err = users.GetByName(ctx, "nobody")
	assert.ErrorIs(t, err, storage.ErrNotFound)
}

func TestUserStorage_List(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	users := storage.NewUserStorage(newDB(t))
	for _, name := range []string{"zed", "amy"} {
		assert.NoError(t, users.Create(ctx, &model.User{Name: name, Role: model.RoleViewer}))
	}
	list, err := users.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, list, 2)
	assert.Equal(t, "amy", list[0].Name)

	count, err := users.Count(ctx)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestUserStorage_Update(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	users := storage.NewUserStorage(newDB(t))
	user := &model.User{Name: "ops", Role: model.RoleRunner}
	assert.NoError(t, users.Create(ctx, user))

	user.Role = model.RoleAdmin
	user.Folders = model.EncodeList([]string{"infra"})
	assert.NoError(t, users.Update(ctx, user))
	got, err := users.Get(ctx, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, model.RoleAdmin, got.Role)
	assert.Equal(t, []string{"infra"}, got.FolderList())

	user.Role = "root"
	assert.Error(t, users.Update(ctx, user))
}

// Deleting a user ends their sessions with them.
func TestUserStorage_Delete(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := newDB(t)
	users, sessions := storage.NewUserStorage(db), storage.NewUserSessionStorage(db)
	user := &model.User{Name: "ops", Role: model.RoleRunner}
	assert.NoError(t, users.Create(ctx, user))
	_, err := sessions.Create(ctx, user.ID, "h1", time.Now().Add(time.Hour))
	assert.NoError(t, err)
	assert.NoError(t, users.Delete(ctx, user.ID))
	_, err = users.Get(ctx, user.ID)
	assert.ErrorIs(t, err, storage.ErrNotFound)
	_, err = sessions.Find(ctx, "h1")
	assert.ErrorIs(t, err, storage.ErrNotFound)
	assert.NoError(t, users.Delete(ctx, user.ID), "deleting twice is fine")
}

func TestUserStorage_MoveFolder(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	users := storage.NewUserStorage(newDB(t))
	limited := &model.User{Name: "dev", Role: model.RoleRunner, Folders: model.EncodeList([]string{"web", "web/docs", "website"})}
	everyone := &model.User{Name: "ops", Role: model.RoleRunner}
	assert.NoError(t, users.Create(ctx, limited))
	assert.NoError(t, users.Create(ctx, everyone))

	changed, err := users.MoveFolder(ctx, "web", "acme/web")
	assert.NoError(t, err)
	assert.Equal(t, 1, changed)
	got, err := users.Get(ctx, limited.ID)
	assert.NoError(t, err)
	assert.Equal(t, []string{"acme/web", "acme/web/docs", "website"}, got.FolderList())
	got, err = users.Get(ctx, everyone.ID)
	assert.NoError(t, err)
	assert.Len(t, got.FolderList(), 0, "a user who sees every folder keeps seeing them")
}
