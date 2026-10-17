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
	assert.Contains(t, body, "task-ui passwd")
	assert.Contains(t, body, "TASK_UI_ADMINS")
	for _, name := range []string{asAdmin, asOps, asViewer, asLimited} {
		assert.Contains(t, body, `href="/settings/users/`+f.users[name].ID+`"`)
	}
	// Nobody is made or removed here.
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
	assert.Contains(t, edit, `value="`+model.RoleRunner+`" readonly`)
	assert.Contains(t, edit, `name="folders" data-type="list" value="infra"`)
	assert.NotContains(t, edit, `name="password"`)
	assert.NotContains(t, edit, `name="role"`)

	assert.Equal(t, http.StatusNotFound, f.get(t, asAdmin, "/settings/users/"+model.NewID()).Status)
	assert.Equal(t, http.StatusNotFound, f.get(t, asAdmin, "/settings/users/new").Status)
}
