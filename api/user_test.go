package api_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
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
	// The listing never holds what a user signs in with.
	assert.NotContains(t, string(res.Raw), hash)
	assert.NotContains(t, string(res.Raw), "$2")
	assert.NotContains(t, string(res.Raw), "password")
	assert.NotContains(t, string(res.Raw), "admin@example.com")
}

// An admin makes a user, with their folders; without a password file
// to write, nobody makes one here.
func TestUser_Create(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	created := f.call(t, asAdmin, http.MethodPost, "/users", map[string]any{"name": "dev", "password": "correct horse", "folders": []string{"infra"}})
	assert.Equal(t, http.StatusCreated, created.Status, string(created.Raw))
	assert.Equal(t, "dev", created.str("name"))
	assert.Equal(t, `["infra"]`, created.str("folders"))
	assert.NotContains(t, string(created.Raw), "correct horse")
	assert.Equal(t, http.StatusConflict, f.call(t, asAdmin, http.MethodPost, "/users", map[string]any{"name": "dev", "password": "correct horse"}).Status)
	assert.Equal(t, http.StatusBadRequest, f.call(t, asAdmin, http.MethodPost, "/users", map[string]any{"name": "eve", "password": "x"}).Status)
	assert.Len(t, f.call(t, asAdmin, http.MethodGet, "/users", nil).list("users"), 5)

	opts := f.opts
	opts.AddUser = nil
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"name":"dev2","password":"correct horse"}`))
	req = req.WithContext(access.WithUser(req.Context(), f.users[asAdmin]))
	rec := httptest.NewRecorder()
	api.New(opts).Routes().ServeHTTP(rec, req)
	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
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

// User/delete removes a user who is no admin and not the caller; without
// a password file to write, removing stays with the file.
func TestUser_Delete(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	limited := f.users[asLimited]
	tokens := storage.NewUserAuthTokenStorage(f.db)
	assert.NoError(t, tokens.Set(ctx, limited.ID, "cli", "h-limited"))

	assert.Equal(t, http.StatusForbidden, f.call(t, asOps, http.MethodDelete, "/users/"+limited.ID, nil).Status)
	admin := f.call(t, asAdmin, http.MethodDelete, "/users/"+f.users[asAdmin].ID, nil)
	assert.Equal(t, http.StatusForbidden, admin.Status)
	assert.Equal(t, "you can't delete yourself", admin.str("error"))
	assert.Equal(t, http.StatusNotFound, f.call(t, asAdmin, http.MethodDelete, "/users/"+model.NewID(), nil).Status)

	// Another admin stays too.
	other := &model.User{Name: "root", Role: model.RoleAdmin}
	assert.NoError(t, f.opts.Users.Create(ctx, other))
	refused := f.call(t, asAdmin, http.MethodDelete, "/users/"+other.ID, nil)
	assert.Equal(t, http.StatusForbidden, refused.Status)
	assert.Equal(t, "you can't delete an admin here", refused.str("error"))

	deleted := f.call(t, asAdmin, http.MethodDelete, "/users/"+limited.ID, nil)
	assert.Equal(t, http.StatusNoContent, deleted.Status, string(deleted.Raw))
	_, err := f.opts.Users.Get(ctx, limited.ID)
	assert.ErrorIs(t, err, storage.ErrNotFound)
	_, err = tokens.Find(ctx, "h-limited")
	assert.ErrorIs(t, err, storage.ErrNotFound)

	opts := f.opts
	opts.RemoveUser = nil
	req := httptest.NewRequest(http.MethodDelete, "/users/"+f.users[asViewer].ID, nil)
	req = req.WithContext(access.WithUser(req.Context(), f.users[asAdmin]))
	rec := httptest.NewRecorder()
	api.New(opts).Routes().ServeHTTP(rec, req)
	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	assert.Contains(t, rec.Body.String(), "task-ui passwd")
}
