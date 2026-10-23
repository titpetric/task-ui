package view_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/titpetric/task-ui/beszel"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/keystore"
	"github.com/titpetric/task-ui/model"
)

// must returns a value its constructor made without an error.
func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

// A menu keeps its groups in order, leaves out the empty parts a
// template passes for an item it skips, and puts the danger items last.
func TestNewMenu(t *testing.T) {
	t.Parallel()

	edit := must(view.NewMenuItem("edit", "Edit", "href", "/e"))
	del := must(view.NewDangerItem("trash", "Delete", "data-delete", "/api/x"))
	copyItem := must(view.NewMenuItem("copy", "Copy", "data-copy", "x"))
	sep := view.MenuItem{Sep: true}
	fp := must(view.NewMenuDetail("Fingerprint", "SHA256:x"))

	var none *time.Time
	m, err := view.NewMenu("Actions on x", del, edit, false, "", nil, none, map[string]int{}, 0, sep, sep, copyItem, fp)
	assert.NoError(t, err)
	assert.False(t, m.Empty())
	assert.Equal(t, []view.MenuDetail{fp}, m.Details)
	assert.Equal(t, [][]view.MenuItem{{edit}, {copyItem}, {del}}, m.Groups())

	empty, err := view.NewMenu("Actions on y", false, sep)
	assert.NoError(t, err)
	assert.True(t, empty.Empty())

	_, err = view.NewMenu("Actions on z", "Delete")
	assert.Error(t, err)
	_, err = view.NewMenu("Actions on z", true)
	assert.Error(t, err)
}

func render(t *testing.T, name string, p *view.Page) string {
	t.Helper()
	v, err := view.New()
	assert.NoError(t, err)
	var buf bytes.Buffer
	assert.NoError(t, v.Render(&buf, name, p))
	return buf.String()
}

// The rows of the pages show their actions behind the [...] menu: an
// icon and a label each, the details on top, the danger group last,
// with the data-* and confirm attributes the loose buttons had.
func TestMenu_Pages(t *testing.T) {
	t.Parallel()

	p := admin(t)
	p.Users = []model.User{{ID: "u2", Name: "ann", Role: model.RoleViewer}}
	p.CanAddUsers = true
	body := render(t, "users", p)
	assert.Contains(t, body, `<summary class="btn-icon" aria-haspopup="menu" aria-expanded="false" title="Actions on ann" aria-label="Actions on ann">`)
	assert.Contains(t, body, `<div class="more-items" role="menu" aria-label="Actions on ann">`)
	assert.Contains(t, body, `<a class="more-item" role="menuitem" href="/users/ann"><svg class="ic"><use href="#i-users"/></svg><span>Profile</span></a>`)
	assert.Contains(t, body, `<a class="more-item" role="menuitem" href="/runs?user=ann">`)
	assert.Contains(t, body, `<hr class="more-sep">`)
	assert.Contains(t, body, `<button type="button" class="more-item danger" role="menuitem" data-delete="/api/users/u2" data-confirm="Delete the user ann?`)
	assert.Contains(t, body, `data-confirm-ok="Delete" data-next="/settings/users?notice=User+deleted"><svg class="ic"><use href="#i-trash"/></svg><span>Delete</span></button>`)
	assert.True(t, strings.Index(body, "<span>Permissions</span>") < strings.Index(body, "<span>Delete</span>"), "the danger group comes last")
	assert.NotContains(t, body, `class="btn-icon danger"`)

	p = admin(t)
	next := time.Now()
	p.Schedules = []model.Schedule{{ID: "s1", Folder: "infra", Job: "plan", Expression: "@daily", Enabled: 1, NextRunAt: &next}}
	body = render(t, "schedules", p)
	assert.Contains(t, body, `data-api="/api/schedules/s1" data-method="PUT" data-body="{&#34;enabled&#34;:false}" data-next="reload"><svg class="ic"><use href="#i-stop"/></svg><span>Disable</span>`)
	assert.Contains(t, body, `class="more-item danger" role="menuitem" data-delete="/api/schedules/s1" data-confirm="Delete the schedule @daily for infra / plan?"`)

	p = admin(t)
	body = render(t, "runs", p)
	run := p.Runs[0]
	assert.Contains(t, body, `<a class="btn btn-sm" href="/runs/`+run.ID+`" data-attach="`+run.ID+`">Attach</a>`, "Attach stays a visible button")
	assert.Contains(t, body, `<dt>Run</dt><dd class="mono">`+run.ID+`</dd>`)
	assert.Contains(t, body, `role="menuitem" data-stop="`+run.ID+`"><svg class="ic"><use href="#i-stop"/></svg><span>Stop</span>`)
	assert.Contains(t, body, `class="more-item danger" role="menuitem" data-kill="`+run.ID+`">`)

	p = admin(t)
	rotated := time.Now()
	ws := keystore.Key{Name: "workspace_acme_git_ed25519", Purpose: keystore.PurposeSSH, Scope: "workspace:acme", Algo: "ed25519", Fingerprint: "SHA256:ws",
		PublicKey: "ssh-ed25519 AAAAws", CreatedAt: time.Now(), RotatedAt: &rotated, Roles: []string{"admin"}, Users: []string{"ann"}}
	p.Keys = []keystore.Key{ws}
	p.KeyReads = map[string]int{ws.Name: 3}
	body = render(t, "keys", p)
	assert.Contains(t, body, `<dt>Links</dt><dd>admin, ann</dd>`)
	assert.Contains(t, body, `<dt>Reads</dt><dd>3</dd>`)
	assert.Contains(t, body, `<dt>Scope</dt><dd>workspace acme</dd>`)
	assert.Contains(t, body, `data-delete="/api/keys/workspace_acme_git_ed25519"`)

	p.Workspace, p.KeysDir = &model.Workspace{Folder: "acme"}, "/keys"
	body = render(t, "workspace_edit", p)
	assert.Contains(t, body, `href="/settings/workspaces/acme/keys/workspace_acme_git_ed25519"><svg class="ic"><use href="#i-edit"/></svg><span>Edit links</span>`)
	assert.Contains(t, body, `data-delete="/api/workspaces/acme/keys/workspace_acme_git_ed25519" data-confirm="Remove workspace_acme_git_ed25519? The workspace&#39;s runs lose it.`)

	p = admin(t)
	p.Hosts = []beszel.System{{Name: "pve", Host: "10.0.0.2", Status: "up"}, {Name: "docker1", Host: "10.0.0.12", Status: "down"}}
	p.Logins = map[string]string{"10.0.0.2": "debian@10.0.0.2"}
	body = render(t, "inventory", p)
	assert.Contains(t, body, `role="menuitem" data-copy="ssh debian@10.0.0.2"><svg class="ic"><use href="#i-copy"/></svg><span>Copy ssh debian@10.0.0.2</span>`)
	assert.Contains(t, body, `aria-label="Actions on docker1"`)
	assert.Contains(t, body, `<dt>Status</dt><dd>down</dd>`)
}
