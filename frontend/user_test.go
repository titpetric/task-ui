package frontend_test

import (
	"net/http"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

func TestUsers(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	body := f.page(t, asAdmin, "/settings/users")
	assert.Contains(t, body, "<h2>Users</h2>")
	for _, name := range []string{asAdmin, asOps, asViewer, asLimited} {
		assert.Contains(t, body, `href="/settings/users/`+f.users[name].ID+`"`)
	}
	// Nobody deletes themselves.
	assert.NotContains(t, body, `data-delete="/api/users/`+f.users[asAdmin].ID+`"`)
	assert.Contains(t, body, `data-delete="/api/users/`+f.users[asOps].ID+`"`)
	assert.Contains(t, body, "infra</td>")
}

func TestUserForms(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	add := f.page(t, asAdmin, "/settings/users/new")
	assert.Contains(t, add, "<h2>Add user</h2>")
	assert.Contains(t, add, `data-api="/api/users" data-method="POST"`)
	assert.Contains(t, add, `<option value="`+model.RoleViewer+`">`)

	edit := f.page(t, asAdmin, "/settings/users/"+f.users[asLimited].ID)
	assert.Contains(t, edit, "<h2>Edit limited</h2>")
	assert.Contains(t, edit, `data-api="/api/users/`+f.users[asLimited].ID+`" data-method="PUT"`)
	assert.Contains(t, edit, `<option value="`+model.RoleRunner+`" selected>`)
	assert.Contains(t, edit, `value="infra"`)
	assert.Contains(t, edit, "readonly")

	assert.Equal(t, http.StatusNotFound, f.get(t, asAdmin, "/settings/users/"+model.NewID()).Status)
}
