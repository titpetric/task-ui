package view_test

import (
	"bytes"
	"testing"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/keystore"
	"github.com/titpetric/task-ui/model"
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
