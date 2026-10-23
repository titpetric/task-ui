//go:build integration

package api_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/permission"
)

// policyOf decodes a policy answer.
func policyOf(t *testing.T, res response) api.PolicyView {
	t.Helper()
	assert.Equal(t, http.StatusOK, res.Status, string(res.Raw))
	var view api.PolicyView
	assert.NoError(t, json.Unmarshal(res.Raw, &view))
	return view
}

func TestPermissionsView(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	assert.Equal(t, http.StatusForbidden, f.call(t, asOps, http.MethodGet, "/permissions", nil).Status)
	res := f.call(t, asAdmin, http.MethodGet, "/permissions", nil)
	assert.Equal(t, http.StatusOK, res.Status)
	var view api.PermissionsView
	assert.NoError(t, json.Unmarshal(res.Raw, &view))
	assert.Equal(t, "all", view.Rules[0].Key)
	assert.Equal(t, "Notes", view.Sections[2].Title)
	assert.Equal(t, "run", view.Permissions[0].Key)
	assert.Contains(t, view.Paths, "note/all")
	assert.Equal(t, []string{"admin", "runner", "viewer"}, view.Roles)
}

// A role's policy reads and changes by path; the admin role's stays fixed,
// and a bad entry changes nothing.
func TestPolicyView_Role(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	runner := policyOf(t, f.call(t, asAdmin, http.MethodGet, "/permissions/roles/runner", nil))
	assert.Equal(t, permission.Allow, runner.Policy["run"])
	assert.True(t, runner.Effective["run"])
	assert.False(t, runner.Effective["secret/view"])
	assert.False(t, runner.Fixed)

	admin := policyOf(t, f.call(t, asAdmin, http.MethodGet, "/permissions/roles/admin", nil))
	assert.True(t, admin.Fixed)
	assert.True(t, admin.Effective["secret/delete"])
	assert.Equal(t, http.StatusConflict, f.call(t, asAdmin, http.MethodPut, "/permissions/roles/admin", map[string]string{"run": "deny"}).Status)
	assert.Equal(t, http.StatusNotFound, f.call(t, asAdmin, http.MethodGet, "/permissions/roles/root", nil).Status)

	changed := policyOf(t, f.call(t, asAdmin, http.MethodPut, "/permissions/roles/runner", map[string]string{"run": "deny", "secret/all": "allow"}))
	assert.False(t, changed.Effective["run"])
	assert.True(t, changed.Effective["secret/edit"])
	ops := policyOf(t, f.call(t, asAdmin, http.MethodGet, "/permissions/users/ops", nil))
	assert.False(t, ops.Effective["run"], "the write reaches the cache")

	for _, bad := range []map[string]string{{"run": "maybe"}, {"secret/nothing": "allow", "run": ""}} {
		res := f.call(t, asAdmin, http.MethodPut, "/permissions/roles/runner", bad)
		assert.Equal(t, http.StatusBadRequest, res.Status, string(res.Raw))
	}
	assert.False(t, policyOf(t, f.call(t, asAdmin, http.MethodGet, "/permissions/roles/runner", nil)).Effective["run"], "a bad entry changes nothing")
	cleared := policyOf(t, f.call(t, asAdmin, http.MethodPut, "/permissions/roles/runner", map[string]string{"run": ""}))
	_, set := cleared.Policy["run"]
	assert.False(t, set, "an empty effect inherits")
	assert.Equal(t, http.StatusForbidden, f.call(t, asOps, http.MethodPut, "/permissions/roles/viewer", map[string]string{"run": "allow"}).Status)
}

// A user's policy needs user/edit; your own, and your own role's, take
// it through the role, so a user policy grants itself nothing.
func TestPolicyView_User(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	assert.Equal(t, http.StatusForbidden, f.call(t, asOps, http.MethodGet, "/permissions/users/ops", nil).Status)
	assert.Equal(t, http.StatusNotFound, f.call(t, asAdmin, http.MethodGet, "/permissions/users/nobody", nil).Status)
	assert.Equal(t, http.StatusConflict, f.call(t, asAdmin, http.MethodPut, "/permissions/users/admin", map[string]string{"run": "deny"}).Status)
	assert.True(t, policyOf(t, f.call(t, asAdmin, http.MethodGet, "/permissions/users/admin", nil)).Fixed)

	ops := policyOf(t, f.call(t, asAdmin, http.MethodPut, "/permissions/users/ops", map[string]string{"user/view": "allow", "user/edit": "allow", "note/delete": "deny"}))
	assert.Equal(t, "runner", ops.Role)
	assert.Equal(t, permission.Allow, ops.RolePolicy["note/delete"])
	assert.False(t, ops.Effective["note/delete"], "the user's deny beats the role's allow")
	assert.True(t, ops.Effective["user/edit"])

	assert.Equal(t, http.StatusForbidden, f.call(t, asOps, http.MethodPut, "/permissions/users/ops", map[string]string{"secret/view": "allow"}).Status)
	assert.Equal(t, http.StatusForbidden, f.call(t, asOps, http.MethodPut, "/permissions/roles/runner", map[string]string{"secret/view": "allow"}).Status)
	assert.False(t, policyOf(t, f.call(t, asOps, http.MethodGet, "/permissions/users/ops", nil)).Effective["secret/view"])

	viewer := policyOf(t, f.call(t, asOps, http.MethodPut, "/permissions/users/viewer", map[string]string{"note/edit": "deny"}))
	assert.False(t, viewer.Effective["note/edit"])

	// user/edit through the role reaches your own policy.
	policyOf(t, f.call(t, asAdmin, http.MethodPut, "/permissions/roles/viewer", map[string]string{"user/view": "allow", "user/edit": "allow"}))
	viewer = policyOf(t, f.call(t, asViewer, http.MethodPut, "/permissions/users/viewer", map[string]string{"note/edit": ""}))
	assert.True(t, viewer.Effective["note/edit"])
}
