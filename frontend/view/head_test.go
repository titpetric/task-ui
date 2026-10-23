package view_test

import (
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/keystore"
	"github.com/titpetric/task-ui/model"
)

// A head takes its empty line as a string or a false value, keeps its
// buttons in order and leaves out the false ones.
func TestNewHead(t *testing.T) {
	t.Parallel()

	upload := must(view.NewMenuItem("upload", "Upload key", "href", "/u"))
	generate := must(view.NewPrimaryItem("key", "Generate key", "href", "/g"))
	h, err := view.NewHead("Keys", "SSH keys", "No keys added.", upload, false, nil, generate)
	assert.NoError(t, err)
	assert.Equal(t, "No keys added.", h.Empty)
	assert.Equal(t, []view.MenuItem{upload, generate}, h.Buttons)

	h, err = view.NewHead("Keys", "", false)
	assert.NoError(t, err)
	assert.Equal(t, "", h.Empty)
	assert.Equal(t, 0, len(h.Buttons))

	_, err = view.NewHead("Keys", "", true)
	assert.Error(t, err)
	_, err = view.NewHead("Keys", "", "", "Upload key")
	assert.Error(t, err)
}

// keyRow returns the row of the key named in a page, its data-next left
// out: the one attribute that names the page a change returns to.
func keyRow(t *testing.T, body, name string) string {
	t.Helper()
	at := strings.Index(body, `<tr id="key-`+name+`">`)
	assert.True(t, at >= 0, "no row of "+name)
	row := body[at:]
	row = row[:strings.Index(row, "</tr>")]
	return regexp.MustCompile(` data-next="[^"]*"`).ReplaceAllString(row, "")
}

// An SSH key renders the same row on every page that lists keys: the
// Keys page, Your SSH keys and a profile, with its public key to copy
// inline, in code with the copy button right after it.
func TestHead_SharedKeyRows(t *testing.T) {
	t.Parallel()

	key := keystore.Key{Name: "deploy_ed25519", Purpose: keystore.PurposeSSH, Algo: "ed25519", Fingerprint: "SHA256:d",
		PublicKey: "ssh-ed25519 AAAAdeploy task-ui", CreatedAt: time.Now(), Users: []string{"ops"}}
	p := admin(t)
	p.Keys = []keystore.Key{key}
	keys := keyRow(t, render(t, "keys", p), key.Name)
	assert.Equal(t, keys, keyRow(t, render(t, "my_keys", p), key.Name))
	assert.Equal(t, keys, keyRow(t, render(t, "profile", p), key.Name))
	assert.Contains(t, keys, `<span class="copyable"><code title="ssh-ed25519 AAAAdeploy task-ui">`)
	assert.Contains(t, keys, `</code><button type="button" class="btn-icon" data-copy="ssh-ed25519 AAAAdeploy task-ui" title="Copy the public key of deploy_ed25519"`)
}

// projectTable returns the table of projects of a page.
func projectTable(t *testing.T, body string) string {
	t.Helper()
	at := strings.Index(body, `<table class="table projects">`)
	assert.True(t, at >= 0, "no projects table")
	return body[at : at+strings.Index(body[at:], "</table>")]
}

// A workspace's page lists its projects with the table of the Projects
// page, Add project top right beside Export, and Delete in the head's
// [...]; an empty list draws one line and no table: no project,
// and no key beside Upload key and Generate key.
func TestHead_WorkspacePage(t *testing.T) {
	t.Parallel()

	p := admin(t)
	p.Workspace = &model.Workspace{Folder: "acme", SourceURL: "git@example.com:acme.git"}
	p.Nested = map[string][]string{"acme": {"acme/api"}}
	p.KeysDir = "/keys"
	p.Entries = []folder.Entry{{ID: "acme/api", Sources: []string{"atkins.yml"}, Jobs: 2}}
	body := render(t, "workspace_edit", p)
	assert.Equal(t, projectTable(t, render(t, "projects", p)), projectTable(t, body))
	assert.Contains(t, body, `href="/settings/projects/acme/api"><svg class="ic"><use href="#i-settings"/></svg><span>Project settings</span>`)

	head := body[strings.Index(body, `<div class="page-head">`):strings.Index(body, `<section class="block">`)]
	add, export, del := strings.Index(head, `href="/settings/projects/new?workspace=acme"`), strings.Index(head, `data-dialog-open="export-dialog"`), strings.Index(head, `data-delete="/api/workspaces/acme"`)
	assert.True(t, add > 0 && export > add && del > export, "Add project, Export, then Delete in the [...]")
	assert.Contains(t, head, `class="more-item danger"`)
	assert.Contains(t, head, `data-confirm="Delete the workspace acme with its checkout, keys and secrets, and the projects inside it: acme/api?`)

	keys := body[strings.Index(body, `id="keys"`):]
	assert.Contains(t, keys, `<p class="empty-line">No keys added.</p>`)
	assert.Contains(t, keys, `<a class="btn btn-sm" href="/settings/workspaces/acme/keys/upload">`)
	assert.Contains(t, keys, `<a class="btn btn-sm" href="/settings/workspaces/acme/keys/generate">`)
	assert.NotContains(t, keys, `<table`)

	p.Entries = nil
	body = render(t, "workspace_edit", p)
	assert.Contains(t, body, `<p class="empty-line">No projects in acme yet.</p>`)
	assert.NotContains(t, body, `<table class="table projects">`)
}

// Every list page holds its create action top right, in the page head,
// worded Add <thing> with the plus icon.
func TestHead_CreateActions(t *testing.T) {
	t.Parallel()

	p := admin(t)
	p.CanAddUsers = true
	for page, add := range map[string]string{
		"projects":   `href="/settings/projects/new"><svg class="ic"><use href="#i-plus"/></svg>Add project</a>`,
		"workspaces": `href="/settings/workspaces/new"><svg class="ic"><use href="#i-plus"/></svg>Add workspace</a>`,
		"users":      `href="/settings/users/new"><svg class="ic"><use href="#i-plus"/></svg>Add user</a>`,
		"workflows":  `href="/workflows/new"><svg class="ic"><use href="#i-plus"/></svg>Add workflow</a>`,
		"keys":       `href="/settings/keys/generate"><svg class="ic"><use href="#i-key"/></svg>Generate key</a>`,
	} {
		body := render(t, page, p)
		head := body[strings.Index(body, `<div class="page-head">`):]
		head = head[:strings.Index(head, "\n</div>")]
		assert.Contains(t, head, add, page)
	}
}
