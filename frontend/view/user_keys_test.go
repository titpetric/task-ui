package view_test

import (
	"strings"
	"testing"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/keystore"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/permission"
)

// UserKeys names the instance SSH keys each user gets, by their name
// and by their effective role, sorted; an encryption key, a workspace's
// key and a key linked to nobody they know count for no one, and a user
// with no key has no entry.
func TestUserKeys(t *testing.T) {
	t.Parallel()
	keys := []keystore.Key{
		{Name: "forge_ed25519", Purpose: keystore.PurposeSSH, Scope: "instance", Roles: []string{"runner"}, Users: []string{"admin"}},
		{Name: "deploy_rsa4096", Purpose: keystore.PurposeSSH, Scope: "instance", Roles: []string{"admin"}},
		{Name: "user_alice_ed25519", Purpose: keystore.PurposeSSH, Scope: "instance", Users: []string{"alice"}},
		{Name: "spare_ed25519", Purpose: keystore.PurposeSSH, Scope: "instance"},
		{Name: "task_encryption_ed25519", Purpose: keystore.PurposeEncrypt, Scope: "instance", Roles: []string{"admin"}},
		{Name: "workspace_malta_ssh_ed25519", Purpose: keystore.PurposeSSH, Scope: keystore.ScopeWorkspace("w1"), Roles: []string{"admin"}},
	}
	users := []model.User{
		{Name: "admin", Role: model.RoleAdmin},
		{Name: "alice", Role: model.RoleRunner},
		{Name: "ops", Role: model.RoleViewer, Admin: true},
		{Name: "guest", Role: model.RoleViewer},
	}
	got := view.UserKeys(keys, users)
	assert.Equal(t, []string{"deploy_rsa4096", "forge_ed25519"}, got["admin"], "by role and by name, sorted")
	assert.Equal(t, []string{"forge_ed25519", "user_alice_ed25519"}, got["alice"])
	assert.Equal(t, []string{"deploy_rsa4096"}, got["ops"], "the effective role, admin, decides")
	_, has := got["guest"]
	assert.False(t, has, "no entry without a key")
	assert.Len(t, view.UserKeys(nil, users), 0)
}

// keysBlock returns the Keys block of a rendered page, up to the next block.
func keysBlock(t *testing.T, body string) string {
	t.Helper()
	at := strings.Index(body, `<section class="block" id="keys">`)
	assert.True(t, at >= 0, "a Keys block")
	rest := body[at:]
	return rest[:strings.Index(rest, "</section>")]
}

// The user's Keys block, on the profile and on the edit page, adds a key
// linked to them: Generate key and Upload key, for key/edit.
func TestUserKeys_AddButtons(t *testing.T) {
	t.Parallel()

	for _, page := range []string{"profile", "user_edit"} {
		p := admin(t)
		p.Profile.Own = true
		block := keysBlock(t, render(t, page, p))
		assert.Contains(t, block, `href="/settings/keys/generate?for=ops"`, page)
		assert.Contains(t, block, `href="/settings/keys/upload?for=ops"`, page)
		assert.Contains(t, block, "No key linked to ops", page)

		signIn(p, &model.User{Name: "ops", Role: model.RoleRunner}, permission.Policy{"key/view": permission.Allow})
		block = keysBlock(t, render(t, page, p))
		assert.NotContains(t, block, "/settings/keys/generate", page+": key/view alone adds nothing")
		assert.NotContains(t, block, "/settings/keys/upload", page)
	}
}

// The profile lists the keys linked to the user, with the row every key
// page shares, for their own page or with key/view.
func TestUserKeys_ProfileRows(t *testing.T) {
	t.Parallel()

	key := keystore.Key{Name: "deploy_ed25519", Purpose: keystore.PurposeSSH, Algo: "ed25519", Fingerprint: "SHA256:abc",
		PublicKey: "ssh-ed25519 AAAAdeploy task-ui", Users: []string{"ops"}}
	p := admin(t)
	p.Keys = []keystore.Key{key}
	p.Profile.Own = true
	body := render(t, "profile", p)
	assert.Equal(t, keyRow(t, render(t, "keys", p), key.Name), keyRow(t, body, key.Name))
	assert.Contains(t, body, `id="keys"`)

	signIn(p, &model.User{Name: "guest", Role: model.RoleViewer}, nil)
	p.Profile.Own = false
	assert.NotContains(t, render(t, "profile", p), `id="keys"`, "another user's keys take key/view")
}

// Your SSH keys, in the avatar menu, adds a key for yourself when you
// may; otherwise it says who can.
func TestMyKeys_AddButtons(t *testing.T) {
	t.Parallel()

	p := admin(t)
	body := render(t, "my_keys", p)
	assert.Contains(t, body, `href="/settings/keys/generate?for=admin"`)
	assert.Contains(t, body, `href="/settings/keys/upload?for=admin"`)
	assert.Contains(t, body, "Generate or upload one for yourself.")

	signIn(p, &model.User{Name: "guest", Role: model.RoleViewer}, nil)
	body = render(t, "my_keys", p)
	assert.NotContains(t, body, "/settings/keys/generate")
	assert.Contains(t, body, "Someone with Allow Edits on Instance Keys can grant one")
}
