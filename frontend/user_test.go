package frontend_test

import (
	"net/http"
	"testing"

	"github.com/titpetric/task-ui/frontend"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

func TestUsers(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	body := f.page(t, asAdmin, "/settings/users")
	assert.Contains(t, body, "<h2>Users</h2>")
	assert.Contains(t, body, "task-ui passwd")
	assert.Contains(t, body, "TASK_UI_ADMINS")
	for _, name := range []string{asAdmin, asOps, asViewer, asLimited} {
		assert.Contains(t, body, `href="/settings/users/`+f.users[name].ID+`"`)
	}
	// The page removes no user, and makes one only with a password file
	// to write.
	assert.NotContains(t, body, "data-delete")
	assert.NotContains(t, body, "/settings/users/new")
	assert.Contains(t, body, "infra</td>")
}

func TestUserForm(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	edit := f.page(t, asAdmin, "/settings/users/"+f.users[asLimited].ID)
	assert.Contains(t, edit, "<h2>Folders of limited</h2>")
	assert.Contains(t, edit, `data-api="/api/users/`+f.users[asLimited].ID+`" data-method="PUT"`)
	// The name and the role are text, not fields; the folders are Custom,
	// infra checked.
	assert.NotContains(t, edit, `id="u-name"`)
	assert.Contains(t, edit, `<span class="badge">`+model.RoleRunner+`</span>`)
	assert.Contains(t, edit, `<input type="radio" name="scope" value="custom" checked>`)
	assert.Contains(t, edit, `name="folders" value="infra" data-type="set" checked>`)
	assert.NotContains(t, edit, `name="password"`)
	assert.NotContains(t, edit, `name="role"`)
	// All projects: the boxes go off, none sent.
	all := f.page(t, asAdmin, "/settings/users/"+f.users[asOps].ID)
	assert.Contains(t, all, `<input type="radio" name="scope" value="all" checked>`)
	assert.Contains(t, all, `name="folders" value="infra" data-type="set" disabled>`)

	assert.Equal(t, http.StatusNotFound, f.get(t, asAdmin, "/settings/users/"+model.NewID()).Status)
	assert.Equal(t, http.StatusNotFound, f.get(t, asAdmin, "/settings/users/new").Status)

	// With a password file to write, Create user makes one.
	f.opts.AddUsers = true
	h, err := frontend.New(f.opts)
	assert.NoError(t, err)
	f.server.Config.Handler = h.Routes()
	assert.Contains(t, f.page(t, asAdmin, "/settings/users"), `href="/settings/users/new">Create user</a>`)
	created := f.page(t, asAdmin, "/settings/users/new")
	assert.Contains(t, created, `data-api="/api/users" data-method="POST"`)
	assert.Contains(t, created, `name="password" required minlength="8"`)
}
