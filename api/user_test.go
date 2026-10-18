package api_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

func TestUser_AdminOnly(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	for _, c := range []struct{ method, path string }{
		{http.MethodGet, "/users"},
		{http.MethodPost, "/users"},
		{http.MethodPut, "/users/" + f.users[asOps].ID},
		{http.MethodDelete, "/users/" + f.users[asOps].ID},
	} {
		res := f.call(t, asOps, c.method, c.path, map[string]any{})
		assert.Equal(t, http.StatusForbidden, res.Status, c.method+" "+c.path)
	}
}

func TestUser_List(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	ctx := context.Background()
	admin := f.users[asAdmin]
	hash, err := access.HashPassword("secret")
	assert.NoError(t, err)
	assert.NoError(t, storage.NewUserAuthPasswordStorage(f.db).SetPassword(ctx, admin.ID, hash))
	assert.NoError(t, storage.NewUserEmailStorage(f.db).Add(ctx, admin.ID, "admin@example.com"))

	res := f.call(t, asAdmin, http.MethodGet, "/users", nil)
	assert.Equal(t, http.StatusOK, res.Status)
	assert.Len(t, res.list("users"), 4)
	// What a user signs in with is never in the answer.
	assert.NotContains(t, string(res.Raw), hash)
	assert.NotContains(t, string(res.Raw), "$2")
	assert.NotContains(t, string(res.Raw), "password")
	assert.NotContains(t, string(res.Raw), "admin@example.com")
}

// Users are made and removed in the password file, not here.
func TestUser_ManagedElsewhere(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	created := f.call(t, asAdmin, http.MethodPost, "/users", map[string]any{"name": "dev", "password": "x"})
	assert.Equal(t, http.StatusMethodNotAllowed, created.Status)
	assert.Contains(t, created.str("error"), "task-ui passwd")

	deleted := f.call(t, asAdmin, http.MethodDelete, "/users/"+f.users[asViewer].ID, nil)
	assert.Equal(t, http.StatusMethodNotAllowed, deleted.Status)
	assert.Len(t, f.call(t, asAdmin, http.MethodGet, "/users", nil).list("users"), 4)
}

func TestUser_Update(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	path := "/users/" + f.users[asViewer].ID

	res := f.call(t, asAdmin, http.MethodPut, path, api.UserRequest{Folders: []string{"private"}})
	assert.Equal(t, http.StatusOK, res.Status, string(res.Raw))
	user, err := f.opts.Users.Get(context.Background(), f.users[asViewer].ID)
	assert.NoError(t, err)
	assert.Equal(t, []string{"private"}, user.FolderList())
	// The role is not this endpoint's to change.
	assert.Equal(t, model.RoleViewer, user.Role)

	all := f.call(t, asAdmin, http.MethodPut, path, api.UserRequest{})
	assert.Equal(t, http.StatusOK, all.Status)
	user, _ = f.opts.Users.Get(context.Background(), f.users[asViewer].ID)
	assert.Len(t, user.FolderList(), 0)

	missing := f.call(t, asAdmin, http.MethodPut, "/users/"+model.NewID(), api.UserRequest{})
	assert.Equal(t, http.StatusNotFound, missing.Status)
	assert.Equal(t, http.StatusBadRequest, f.call(t, asAdmin, http.MethodPut, path, nil).Status)
}
