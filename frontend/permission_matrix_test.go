package frontend_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/permission"
)

// Settings, Permissions shows a matrix per role: the admin's fixed, the
// others to change with user/edit, read only with user/view alone.
func TestPermissionMatrix(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	body := f.page(t, asAdmin, "/settings/permissions")
	assert.Contains(t, body, "<h2>Permissions</h2>")
	for _, role := range []string{"admin", "runner", "viewer"} {
		assert.Contains(t, body, `id="role-`+role+`"`)
	}
	assert.NotContains(t, body, `data-api="/api/permissions/roles/admin"`)
	assert.Contains(t, body, `data-api="/api/permissions/roles/runner" data-method="PUT"`)
	assert.Contains(t, body, `data-api="/api/permissions/roles/viewer" data-method="PUT"`)
	assert.Contains(t, body, `<input type="hidden" name="run" value="allow" data-was="allow" disabled>`)
	assert.Contains(t, body, "<small>can do everything</small>")
	assert.Contains(t, f.page(t, asAdmin, "/settings"), `href="/settings/permissions"`)

	assert.Equal(t, http.StatusForbidden, f.get(t, asOps, "/settings/permissions").Status)
	assert.NoError(t, f.opts.Users.Policies().Put(context.Background(), permission.KindUser, asOps, "user/view", permission.Allow))
	read := f.page(t, asOps, "/settings/permissions")
	assert.NotContains(t, read, "data-matrix")
	assert.Contains(t, read, "read only")
}
