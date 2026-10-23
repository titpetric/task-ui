package view_test

import (
	"bytes"
	"testing"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/keystore"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/runner"
)

// The Instance page shows the configuration and the footer of its step.
func TestInstance(t *testing.T) {
	t.Parallel()
	v, err := view.New()
	assert.NoError(t, err)

	p := admin(t)
	p.Step = view.StepInstance
	p.Instance = &view.Instance{Root: "/srv/root", DataDir: "/srv/data", Beszel: true, TimeZone: "Europe/Ljubljana, CEST",
		Backlog: &runner.BacklogState{Parallel: 4, Waiting: 2, Auto: 80, Length: 80}}
	var buf bytes.Buffer
	assert.NoError(t, v.Render(&buf, "instance", p))
	body := buf.String()
	assert.Contains(t, body, "/srv/root")
	assert.Contains(t, body, `<a href="/settings/users">Users</a>`)
	assert.NotContains(t, body, "TASK_UI_ADMINS")
	assert.Contains(t, body, `class="setup-trail"`)
	assert.Contains(t, body, `rel="next">Next: Data retention<svg`)
	assert.NotContains(t, body, `rel="prev"`)
	for _, id := range []string{"general", "runs", "paths", "keys", "integrations"} {
		assert.Contains(t, body, `id="`+id+`"`, id)
	}
	assert.Contains(t, body, "Europe/Ljubljana, CEST")
	assert.Contains(t, body, `<td class="value mono ro" data-label="Value">80</td>`, "the automatic backlog is read only")
	assert.Contains(t, body, `data-api="/api/instance" data-method="PUT"`, "an admin saves the backlog length in place")
	assert.Contains(t, body, `name="backlog_length" value="" placeholder="80"`)
	assert.Contains(t, body, "in effect: 80")
}

// A user who is no admin reads the backlog length and saves none.
func TestInstance_Viewer(t *testing.T) {
	t.Parallel()
	v, err := view.New()
	assert.NoError(t, err)

	p := admin(t)
	p.User = &model.User{Name: "ann", Role: model.RoleViewer}
	p.Instance = &view.Instance{Backlog: &runner.BacklogState{Parallel: 2, Auto: 40, Configured: 12, Length: 12}}
	var buf bytes.Buffer
	assert.NoError(t, v.Render(&buf, "instance", p))
	body := buf.String()
	assert.NotContains(t, body, `data-api="/api/instance"`)
	assert.Contains(t, body, `<span class="mono">12</span>`)
	assert.Contains(t, body, "in effect: 12")
}

// The keys of the Instance page: what each is, and its actions behind a
// menu, the public half only.
func TestInstance_Keys(t *testing.T) {
	t.Parallel()
	v, err := view.New()
	assert.NoError(t, err)

	p := admin(t)
	p.Instance = &view.Instance{}
	enc := keystore.Key{Name: "task_encryption_ed25519", Purpose: keystore.PurposeEncrypt, Algo: "ed25519", Fingerprint: "SHA256:enc", PublicKey: "ssh-ed25519 AAAAenc"}
	git := keystore.Key{Name: "deploy_rsa4096", Purpose: keystore.PurposeSSH, Algo: "rsa", Bits: 4096, Fingerprint: "SHA256:git", PublicKey: "ssh-rsa AAAAgit", Git: true}
	p.InstanceKey, p.Keys = &enc, []keystore.Key{enc, git}
	var buf bytes.Buffer
	assert.NoError(t, v.Render(&buf, "instance", p))
	body := buf.String()
	for _, want := range []string{
		"<th>Name</th><th>Algorithm</th><th>Size (bits)</th><th>Fingerprint</th><th>Purpose</th>",
		`<span class="badge">encryption</span>`, `<span class="badge">git &#43; ssh</span>`,
		`data-label="Size">4096</td>`, `data-label="Size">256</td>`,
		`data-copy="ssh-rsa AAAAgit"`, `data-download="ssh-ed25519 AAAAenc" data-filename="task_encryption_ed25519.pub"`,
		`<details class="more-menu row-more">`, "Rotate Key",
	} {
		assert.Contains(t, body, want)
	}
	assert.NotContains(t, body, "Regenerate")
	assert.NotContains(t, body, "PRIVATE KEY")
}
