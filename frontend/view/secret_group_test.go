package view_test

import (
	"testing"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/keystore"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/permission"
)

func TestSecretGroup(t *testing.T) {
	t.Parallel()
	group := view.SecretGroup{Workspace: "acme", Title: "Acme", Secrets: []model.Secret{{Scope: "workspace:acme", Name: "TOKEN"}}, Folders: []string{"acme/api"}}
	assert.Equal(t, "workspace:acme", group.Shared())
	assert.Len(t, group.Secrets, 1)
	assert.Equal(t, []string{"acme/api"}, group.Folders)
	assert.Len(t, group.SharedSecrets(), 1)
	assert.Len(t, group.Projects(), 0)

	group.Secrets = append(group.Secrets, model.Secret{Scope: "acme/web", Name: "B"}, model.Secret{Scope: "acme/api", Name: "A"}, model.Secret{Scope: "acme/web", Name: "C"})
	projects := group.Projects()
	assert.Len(t, projects, 2)
	assert.Equal(t, "acme/api", projects[0].Folder)
	assert.Equal(t, "acme/web", projects[1].Folder)
	assert.Len(t, projects[1].Secrets, 2)
}

// secretsPage is the secrets page of malta with a secret every project
// gets and one of malta/api, as user sees it.
func secretsPage(t *testing.T, user *model.User, role permission.Policy) *view.Page {
	t.Helper()
	p := admin(t)
	signIn(p, user, role)
	p.Workspace = &model.Workspace{ID: "w1", Folder: "malta"}
	p.SecretGroups = []view.SecretGroup{{
		Workspace: "malta", Title: "Malta", Key: &keystore.Key{Name: "instance_encryption"},
		Secrets: []model.Secret{{ID: "s1", Scope: "workspace:malta", Name: "TOKEN"}, {ID: "s2", Scope: "malta/api", Name: "DB_PASS", Content: "hunter2"}},
		Folders: []string{"malta/api"},
	}}
	p.Secret = &model.Secret{Scope: "workspace:malta"}
	return p
}

// The secrets page of a workspace lists its secrets and edits them in
// place: Add secret and Replace the value open dialogs that return to
// it, and no value shows.
func TestWorkspaceSecrets_Page(t *testing.T) {
	t.Parallel()

	body := render(t, "workspace_secrets", secretsPage(t, &model.User{Name: "admin", Role: model.RoleAdmin}, nil))
	assert.Contains(t, body, "<h2>Secrets of malta</h2>")
	assert.Contains(t, body, `<td class="key">TOKEN</td>`)
	assert.Contains(t, body, `<td class="key">DB_PASS</td>`)
	assert.NotContains(t, body, "hunter2")
	assert.Contains(t, body, `id="secret-add-dialog"`)
	assert.Contains(t, body, `<option value="workspace:malta" selected>every project in malta</option>`)
	assert.Contains(t, body, `<option value="malta/api">malta/api</option>`)
	assert.Contains(t, body, `data-dialog-open="secret-replace-dialog" data-fill-scope="malta/api" data-fill-name="DB_PASS"`)
	assert.Contains(t, body, `data-next="/settings/workspaces/malta/secrets?notice=Secret+saved"`)
	assert.Contains(t, body, `data-delete="/api/secrets/s2"`)
	assert.NotContains(t, body, "/settings/secrets/edit")

	body = render(t, "workspace_secrets", secretsPage(t, &model.User{Name: "ops", Role: model.RoleRunner}, permission.Policy{"secret/view": permission.Allow}))
	assert.Contains(t, body, `<td class="key">TOKEN</td>`)
	assert.NotContains(t, body, `id="secret-add-dialog"`)
	assert.NotContains(t, body, "Replace the value")
	assert.NotContains(t, body, "data-delete")
}
