package api_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
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

	res := f.call(t, asAdmin, http.MethodGet, "/users", nil)
	assert.Equal(t, http.StatusOK, res.Status)
	assert.Len(t, res.list("users"), 4)
}

func TestUser_Create(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	res := f.call(t, asAdmin, http.MethodPost, "/users", api.UserRequest{Name: " dev ", Password: "secret", Folders: []string{"infra"}})
	assert.Equal(t, http.StatusCreated, res.Status, string(res.Raw))
	assert.Equal(t, "dev", res.str("name"))
	assert.Equal(t, model.RoleViewer, res.str("role"))

	user, err := f.opts.Users.Get(context.Background(), res.str("id"))
	assert.NoError(t, err)
	assert.True(t, access.CheckPassword(user, "secret"))
	assert.Equal(t, []string{"infra"}, user.FolderList())

	dup := f.call(t, asAdmin, http.MethodPost, "/users", api.UserRequest{Name: "dev", Password: "x", Role: model.RoleRunner})
	assert.Equal(t, http.StatusBadRequest, dup.Status)

	blank := f.call(t, asAdmin, http.MethodPost, "/users", api.UserRequest{Name: "nopass"})
	assert.Equal(t, http.StatusBadRequest, blank.Status)

	role := f.call(t, asAdmin, http.MethodPost, "/users", api.UserRequest{Name: "boss", Password: "x", Role: "king"})
	assert.Equal(t, http.StatusBadRequest, role.Status)
}

func TestUser_Update(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	path := "/users/" + f.users[asViewer].ID

	res := f.call(t, asAdmin, http.MethodPut, path, api.UserRequest{Role: model.RoleRunner, Folders: []string{"private"}, Password: "changed"})
	assert.Equal(t, http.StatusOK, res.Status, string(res.Raw))
	assert.Equal(t, model.RoleRunner, res.str("role"))

	user, err := f.opts.Users.Get(context.Background(), f.users[asViewer].ID)
	assert.NoError(t, err)
	assert.True(t, access.CheckPassword(user, "changed"))
	assert.Equal(t, []string{"private"}, user.FolderList())

	// An empty request changes nothing.
	same := f.call(t, asAdmin, http.MethodPut, path, api.UserRequest{})
	assert.Equal(t, http.StatusOK, same.Status)
	assert.Equal(t, model.RoleRunner, same.str("role"))

	bad := f.call(t, asAdmin, http.MethodPut, path, api.UserRequest{Role: "king"})
	assert.Equal(t, http.StatusBadRequest, bad.Status)

	missing := f.call(t, asAdmin, http.MethodPut, "/users/"+model.NewID(), api.UserRequest{})
	assert.Equal(t, http.StatusNotFound, missing.Status)
}

func TestUser_Delete(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	self := f.call(t, asAdmin, http.MethodDelete, "/users/"+f.users[asAdmin].ID, nil)
	assert.Equal(t, http.StatusBadRequest, self.Status)

	missing := f.call(t, asAdmin, http.MethodDelete, "/users/"+model.NewID(), nil)
	assert.Equal(t, http.StatusNotFound, missing.Status)

	res := f.call(t, asAdmin, http.MethodDelete, "/users/"+f.users[asViewer].ID, nil)
	assert.Equal(t, http.StatusNoContent, res.Status)
	assert.Len(t, f.call(t, asAdmin, http.MethodGet, "/users", nil).list("users"), 3)
}
