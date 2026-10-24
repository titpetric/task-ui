package view_test

import (
	"bytes"
	"testing"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/keystore"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/permission"
)

func TestKeyWorkspace(t *testing.T) {
	t.Parallel()
	row := view.KeyWorkspace{
		Workspace: model.Workspace{Folder: "acme"}, Projects: []string{"acme/api"}, Secrets: 2,
		SSH: []keystore.Key{{Name: "workspace_acme_deploy_ed25519"}, {Name: "workspace_acme_ssh_rsa4096"}},
	}
	assert.Equal(t, []string{"workspace_acme_deploy_ed25519", "workspace_acme_ssh_rsa4096"}, row.Names())
	assert.Equal(t, []string{}, view.KeyWorkspace{}.Names())
}

// The keys page names one encryption key, the instance's, for every
// workspace, and says git + ssh for an SSH key, the git key too.
func TestKeyWorkspace_Page(t *testing.T) {
	t.Parallel()
	v, err := view.New()
	assert.NoError(t, err)
	p := admin(t)
	enc := keystore.Key{Name: "task_encryption_ed25519", Purpose: keystore.PurposeEncrypt, Algo: "ed25519", Fingerprint: "SHA256:enc"}
	git := keystore.Key{Name: "deploy_ed25519", Purpose: keystore.PurposeSSH, Algo: "ed25519", Fingerprint: "SHA256:git", Git: true}
	p.InstanceKey, p.Keys = &enc, []keystore.Key{enc, git}
	p.KeyWorkspaces = []view.KeyWorkspace{{Workspace: model.Workspace{Folder: "acme"}}}
	var buf bytes.Buffer
	assert.NoError(t, v.Render(&buf, "keys", p))
	body := buf.String()
	assert.Contains(t, body, `<span class="faint">the instance's</span>`)
	assert.Contains(t, body, "git + ssh</span>")
	assert.NotContains(t, body, `>git</span>`)
	assert.NotContains(t, body, `name="encrypt"`)
}

// keysPage is the SSH keys page of malta with one key, as user sees it.
func keysPage(t *testing.T, user *model.User, role permission.Policy) *view.Page {
	t.Helper()
	p := admin(t)
	signIn(p, user, role)
	p.Workspace = &model.Workspace{ID: "w1", Folder: "malta"}
	p.KeysDir = "/keys"
	p.Keys = []keystore.Key{{Name: "workspace_malta_deploy_ed25519", Purpose: keystore.PurposeSSH, Scope: keystore.ScopeWorkspace("w1"), Algo: "ed25519", Fingerprint: "SHA256:abc", PublicKey: "ssh-ed25519 AAAA deploy"}}
	p.KeyPolicies = map[string][]string{"workspace_malta_deploy_ed25519": {model.RoleAdmin}}
	p.KeyUsers = map[string][]string{}
	return p
}

// The SSH keys page of a workspace lists its keys and edits them in
// place: Generate, Upload and Edit links open dialogs that return to it.
func TestWorkspaceKeys_Page(t *testing.T) {
	t.Parallel()

	body := render(t, "workspace_keys", keysPage(t, &model.User{Name: "admin", Role: model.RoleAdmin}, nil))
	assert.Contains(t, body, "<h2>SSH keys of malta</h2>")
	assert.Contains(t, body, `<tr id="key-workspace_malta_deploy_ed25519">`)
	assert.Contains(t, body, `data-dialog-open="key-generate-dialog"`)
	assert.Contains(t, body, `data-dialog-open="key-upload-dialog"`)
	assert.Contains(t, body, `data-dialog-open="key-links-workspace_malta_deploy_ed25519"`)
	assert.Contains(t, body, `data-api="/api/workspaces/malta/keys/generate" data-method="POST" data-next="/settings/workspaces/malta/keys?notice=Key+generated"`)
	assert.Contains(t, body, `data-api="/api/workspaces/malta/keys/upload" data-method="POST" data-next="/settings/workspaces/malta/keys?notice=Key+uploaded"`)
	assert.Contains(t, body, `data-api="/api/workspaces/malta/keys/workspace_malta_deploy_ed25519" data-method="PUT" data-next="/settings/workspaces/malta/keys?notice=Links+saved"`)
	assert.Contains(t, body, `name="policies" value="admin" data-type="set" checked>`)
	assert.Contains(t, body, `data-delete="/api/workspaces/malta/keys/workspace_malta_deploy_ed25519"`)
	assert.NotContains(t, body, `href="/settings/workspaces/malta/keys/workspace_malta_deploy_ed25519"`)

	body = render(t, "workspace_keys", keysPage(t, &model.User{Name: "ops", Role: model.RoleRunner}, permission.Policy{"workspace/view": permission.Allow}))
	assert.Contains(t, body, `<tr id="key-workspace_malta_deploy_ed25519">`)
	assert.NotContains(t, body, `id="key-generate-dialog"`)
	assert.NotContains(t, body, "data-delete")

	p := keysPage(t, &model.User{Name: "admin", Role: model.RoleAdmin}, nil)
	p.Keys = nil
	assert.Contains(t, render(t, "workspace_keys", p), "No keys added.")
}
