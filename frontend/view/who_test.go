package view_test

import (
	"strings"
	"testing"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/keystore"
	"github.com/titpetric/task-ui/model"
)

// The element takes the checked names as a list, a template's list, a
// set or nothing, counts the users by the role they act with, and keeps
// a checked name nobody holds as an extra row.
func TestWho(t *testing.T) {
	t.Parallel()
	all := []model.User{{Name: "admin", Role: model.RoleAdmin}, {Name: "ops", Role: model.RoleRunner, Admin: true}, {Name: "dev", Role: model.RoleRunner}, {Name: "guest", Role: model.RoleViewer}}
	var w view.Who
	w, err := view.NewWho("allow", "set", []string{"runner", "dev", "gone"}, []string{"runner", "dev", "gone"}, all, "every runner can")
	assert.NoError(t, err)
	assert.Equal(t, "allow", w.UserName)
	assert.True(t, w.RoleChecked(model.RoleRunner))
	assert.False(t, w.RoleChecked(model.RoleAdmin))
	assert.True(t, w.UserChecked("dev"))
	assert.False(t, w.UserChecked("ops"))
	assert.Equal(t, 2, w.RoleCount(model.RoleAdmin))
	assert.Equal(t, 1, w.RoleCount(model.RoleRunner))
	assert.Equal(t, []string{"gone"}, w.Extra())
	assert.Equal(t, "users", w.UsersNamed("users").UserName)
	assert.True(t, w.AsCompact().Compact)
	assert.False(t, w.Compact)

	set, err := view.NewWho("values.who", "joined", map[string]bool{"ops": true, "admin": true, "off": false}, map[string]bool{"ops": true, "admin": true}, all, "every run gets the token")
	assert.NoError(t, err)
	assert.Equal(t, []string{"admin", "ops"}, set.Roles)
	assert.True(t, set.RoleChecked(model.RoleAdmin))
	assert.Nil(t, set.Extra())

	listed, err := view.NewWho("policies", "set", nil, []any{"alice", nil, ""}, all, "")
	assert.NoError(t, err)
	assert.Equal(t, []string{"alice"}, listed.Users)
	assert.Equal(t, []string{"alice"}, listed.Extra())

	_, err = view.NewWho("x", "set", 3, nil, all, "")
	assert.Error(t, err)
	_, err = view.NewWho("x", "set", nil, []any{3}, all, "")
	assert.Error(t, err)
}

// The pages draw whofield: a key's edit form with policies and users as
// sets and no sentence, since no run gets an instance key with nothing
// checked; a project's Who can run as the allow set, with the compact
// form per job; the roles carry their counts, the users their effective
// role and a state the script fills, and the summary stands under the list.
func TestWho_Render(t *testing.T) {
	t.Parallel()
	p := admin(t)
	p.Users = []model.User{{Name: "admin", Role: model.RoleAdmin}, {Name: "ops", Role: model.RoleRunner, Admin: true}, {Name: "dev", Role: model.RoleRunner}, {Name: "guest", Role: model.RoleViewer}}
	p.Key = &keystore.Key{Name: "deploy_ed25519", Purpose: keystore.PurposeSSH, Algo: "ed25519", Fingerprint: "SHA256:d", Roles: []string{"admin"}, Users: []string{"dev", "gone"}}
	body := render(t, "key_edit", p)
	assert.Contains(t, body, `<div class="whofield" data-who-empty="">`)
	assert.Contains(t, body, `<label class="check"><input type="checkbox" name="policies" value="admin" data-type="set" checked> admin <small class="count">· 2</small></label>`)
	assert.Contains(t, body, `<label class="check"><input type="checkbox" name="policies" value="runner" data-type="set"> runner <small class="count">· 1</small></label>`)
	assert.Contains(t, body, `<label class="check who-user" data-role="admin"><input type="checkbox" name="users" value="ops" data-type="set"> ops <small class="role">admin</small> <small class="who-state"></small></label>`)
	assert.Contains(t, body, `<label class="check who-user" data-role="runner"><input type="checkbox" name="users" value="dev" data-type="set" checked> dev <small class="role">runner</small> <small class="who-state"></small></label>`)
	assert.Contains(t, body, `<label class="check who-user" data-role=""><input type="checkbox" name="users" value="gone" data-type="set" checked> gone <small class="role">no such user</small> <small class="who-state"></small></label>`)
	assert.Contains(t, body, `<p class="dim who-summary"></p>`)
	assert.Equal(t, 1, strings.Count(body, `class="whofield`))

	p.Allow = []string{"runner"}
	p.JobAllow = map[string][]string{"plan": {"ops"}}
	body = render(t, "project_edit", p)
	assert.Contains(t, body, `<div class="whofield" data-who-empty="everyone who may run can">`)
	assert.Contains(t, body, `<input type="checkbox" name="allow" value="runner" data-type="set" checked> runner <small class="count">· 1</small>`)
	assert.Contains(t, body, `<label class="check who-user" data-role="viewer"><input type="checkbox" name="allow" value="guest" data-type="set"> guest <small class="role">viewer</small>`)
	assert.Contains(t, body, `<div class="whofield compact" data-who-empty="the folder&#39;s Who can run applies">`)
	assert.Contains(t, body, `<input type="checkbox" name="job_allow.plan" value="ops" data-type="set" checked>`)
	assert.Equal(t, 1, strings.Count(body, "who-summary"), "the compact form has none")
}
