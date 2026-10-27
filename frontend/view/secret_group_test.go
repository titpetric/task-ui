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

// secretsPage is the secrets page with an env entry and a secret, as
// user sees it.
func secretsPage(t *testing.T, user *model.User, role permission.Policy) *view.Page {
	t.Helper()
	p := admin(t)
	signIn(p, user, role)
	p.SecretGroups = []view.SecretGroup{{
		Workspace: "malta", Title: "Malta", Key: &keystore.Key{Name: "instance_encryption"},
	}}
	p.SecretScopeTitles = map[string]string{"all": "All workspaces", "workspace:malta": "Malta"}
	p.Secrets = []model.Secret{
		{ID: "s1", Scope: "all", Name: "PUBLIC_URL", Content: "https://example.com", IsSecret: 0},
		{ID: "s2", Scope: "workspace:malta", Name: "DB_PASS", Content: "hunter2", IsSecret: 1},
	}
	p.IntegrationEnv = []view.EnvVar{{Name: "ANTHROPIC_AUTH_TOKEN", Redacted: true}, {Name: "ANTHROPIC_BASE_URL", Value: "https://openrouter.ai/api"}}
	return p
}

// The secrets page lists secrets and env entries in one table: an env
// value shows, a secret's never does, and the integrations' computed
// environment is read only below them.
func TestSecrets_Page(t *testing.T) {
	t.Parallel()

	body := render(t, "secrets", secretsPage(t, &model.User{Name: "admin", Role: model.RoleAdmin}, nil))
	assert.Contains(t, body, "<h2>Secrets and env</h2>")
	assert.Contains(t, body, `<td class="key">PUBLIC_URL`)
	assert.Contains(t, body, `<code>https://example.com</code>`)
	assert.Contains(t, body, `<td class="key">DB_PASS`)
	assert.NotContains(t, body, "hunter2")
	assert.Contains(t, body, `href="/settings/secrets/edit?id=s2"`)
	assert.Contains(t, body, `data-delete="/api/secrets/s2"`)
	assert.Contains(t, body, `All workspaces`)
	assert.Contains(t, body, `>Malta</td>`)
	assert.Contains(t, body, `id="integration-env"`)
	assert.Contains(t, body, `<td class="key">ANTHROPIC_AUTH_TOKEN</td><td><span class="faint mono">REDACTED</span>`)
	assert.Contains(t, body, `<code>https://openrouter.ai/api</code>`)

	body = render(t, "secrets", secretsPage(t, &model.User{Name: "ops", Role: model.RoleRunner}, permission.Policy{"secret/view": permission.Allow}))
	assert.Contains(t, body, `<td class="key">PUBLIC_URL`)
	assert.NotContains(t, body, "Edit")
	assert.NotContains(t, body, "data-delete")
}
