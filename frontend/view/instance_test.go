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

// An admin picks the parent environment and sees the server's
// variables by name, each with whether a run gets it and a flag on a
// sensitive one; a value never shows. Anyone else reads the mode.
func TestInstance_ParentEnv(t *testing.T) {
	t.Parallel()
	v, err := view.New()
	assert.NoError(t, err)

	env := &runner.Environment{Environ: func() []string {
		return []string{"PATH=/usr/bin", "WORKSPACE_KEY=hunter2", "GOTOOLCHAIN=auto"}
	}}
	p := admin(t)
	p.Instance = &view.Instance{ParentEnv: runner.EnvAuto, Env: env.Names()}
	var buf bytes.Buffer
	assert.NoError(t, v.Render(&buf, "instance", p))
	body := buf.String()
	assert.Contains(t, body, `id="environment"`)
	assert.Contains(t, body, `name="parent_env" value="auto" checked`)
	assert.Contains(t, body, `<td class="mono">WORKSPACE_KEY</td>`)
	assert.Contains(t, body, `data-env-none="false" data-env-auto="true"><td class="mono">GOTOOLCHAIN</td>`)
	assert.Contains(t, body, ">sensitive</span>")
	assert.Contains(t, body, "<code>LC_*</code>")
	assert.NotContains(t, body, "hunter2")
	assert.NotContains(t, body, "/usr/bin")

	p.User = &model.User{Name: "ann", Role: model.RoleViewer}
	p.Instance = &view.Instance{ParentEnv: runner.EnvAll}
	buf.Reset()
	assert.NoError(t, v.Render(&buf, "instance", p))
	body = buf.String()
	assert.NotContains(t, body, `name="parent_env"`)
	assert.Contains(t, body, "In effect: <strong>all</strong>")
}

// The base variables and the sensitive words the page lists are the
// runner's.
func TestInstance_EnvBase(t *testing.T) {
	t.Parallel()
	i := &view.Instance{}
	assert.Equal(t, runner.BaseEnv, i.EnvBase())
	assert.Equal(t, runner.SensitiveWords, i.EnvSensitive())
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
		"<th>Name</th><th>Algorithm</th><th>Size</th><th>Public key</th><th>Purpose</th>",
		`<span class="badge signal">encryption</span>`, `title="git pulls and clones with it">git + ssh</span>`,
		`data-label="Size">4096 bits</td>`, `data-label="Size">256 bits</td>`,
		`<span class="copyable"><code>ssh-rsa AAAAgit</code>`,
		`<dt>Fingerprint</dt><dd class="mono">SHA256:git</dd>`, `aria-haspopup="menu"`,
		`data-copy="ssh-rsa AAAAgit"`, `data-download="ssh-ed25519 AAAAenc" data-filename="task_encryption_ed25519.pub"`,
		`<details class="more-menu">`, "Rotate Key",
	} {
		assert.Contains(t, body, want)
	}
	assert.NotContains(t, body, "Regenerate")
	assert.NotContains(t, body, "PRIVATE KEY")
}

// A rotated key shows its backup, rotates no more and lists Delete
// backup.
func TestInstance_KeyBackup(t *testing.T) {
	t.Parallel()
	v, err := view.New()
	assert.NoError(t, err)

	p := admin(t)
	p.Instance = &view.Instance{}
	git := keystore.Key{Name: "deploy_ed25519", Purpose: keystore.PurposeSSH, Algo: "ed25519", Fingerprint: "SHA256:new", PublicKey: "ssh-ed25519 AAAAnew", Git: true,
		Backup: "backup/deploy_ed25519", BackupFingerprint: "SHA256:old"}
	p.Keys = []keystore.Key{git}
	var buf bytes.Buffer
	assert.NoError(t, v.Render(&buf, "instance", p))
	body := buf.String()
	assert.Contains(t, body, "BACKUP ACTIVE: backup/deploy_ed25519")
	assert.Contains(t, body, `role="menuitem" disabled><svg class="ic"><use href="#i-rotate"/></svg><span>Rotate Key</span><small class="faint">Delete the backup first</small>`)
	assert.Contains(t, body, `data-delete="/api/keys/deploy_ed25519/backup"`)
	assert.Contains(t, body, "SHA256:old")
	assert.NotContains(t, body, `data-api="/api/keys/deploy_ed25519/rotate"`)

	git.Backup = ""
	p.Keys = []keystore.Key{git}
	buf.Reset()
	assert.NoError(t, v.Render(&buf, "instance", p))
	body = buf.String()
	assert.Contains(t, body, `data-api="/api/keys/deploy_ed25519/rotate"`)
	assert.Contains(t, body, "backup/deploy_ed25519, which runs and git keep using")
	assert.Contains(t, body, `<hr class="more-sep">`)
	assert.NotContains(t, body, "BACKUP ACTIVE")
}
