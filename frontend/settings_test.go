package frontend_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/integration"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// The Settings page lists the steps in setup order, numbered, each with
// how much it holds.
func TestSettings(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.withKey(t)
	assert.NoError(t, f.opts.Secrets.Set(context.Background(), &model.Secret{Scope: "all", Name: "TOKEN", Content: "x"}))

	body := f.page(t, asAdmin, "/settings")
	assert.Contains(t, body, "<h2>Settings</h2>")
	assert.Contains(t, body, "2 folders")
	assert.Contains(t, body, "1 key<")
	assert.Contains(t, body, "1 secret<")
	assert.Contains(t, body, "4 users")
	assert.Contains(t, body, f.opts.Keys.Dir)
	assert.Contains(t, body, `<td class="num dim">1</td>`)
	order := []string{"/settings/instance", "/settings/retention", "/settings/keys", "/settings/secrets", "/settings/integrations", "/settings/users", "/settings/permissions", "/workflows", "/settings/workspaces", "/settings/projects"}
	at := -1
	for _, href := range order {
		i := strings.Index(body, `<td class="key nowrap"><a href="`+href+`"`)
		assert.True(t, i > at, href)
		at = i
	}
}

// The Instance page shows what the server runs with, and no secret.
func TestSettings_Instance(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.withKey(t)
	assert.NoError(t, f.opts.Integrations.Save(context.Background(), integration.Beszel, map[string]string{"url": "https://hub.example", "token": "beszel-secret-1"}))
	assert.NoError(t, f.opts.Integrations.Save(context.Background(), integration.Claude, map[string]string{"token": "sk-claude-1"}))

	body := f.page(t, asAdmin, "/settings/instance")
	assert.Contains(t, body, "<h2>Instance</h2>")
	assert.Contains(t, body, f.root)
	assert.Contains(t, body, "/srv/task-ui/data")
	assert.Contains(t, body, "@claude-note")
	assert.Contains(t, body, "/api/mcp")
	assert.Contains(t, body, `<th>Variable</th>`)
	assert.Contains(t, body, `<td class="nowrap" data-label="Variable"><code>TASK_UI_ROOT</code></td>`)
	assert.Contains(t, body, `<span class="badge ok">on</span></td><td class="nowrap" data-label="Variable"><span class="faint">none</span></td><td class="desc dim">Hosts from the Beszel`)
	assert.NotContains(t, body, "beszel-secret-1")
	assert.NotContains(t, body, "sk-claude-1")
	assert.Contains(t, body, `rel="next">Next: Data retention<svg`)

	assert.Equal(t, http.StatusForbidden, f.get(t, asViewer, "/settings/instance").Status)
}

// Each step ends with the step before and the step after; a sub-page
// carries its step's footer.
func TestSettings_Footer(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.withKey(t)
	ctx := context.Background()
	assert.NoError(t, f.opts.Workspaces.Create(ctx, &model.Workspace{Folder: "acme", Title: "Acme"}))

	for path, want := range map[string][2]string{
		"/settings/keys":            {"Data retention", "Secrets"},
		"/settings/keys/generate":   {"Data retention", "Secrets"},
		"/settings/secrets":         {"Keys", "Integrations"},
		"/settings/integrations":    {"Secrets", "Users"},
		"/settings/users":           {"Integrations", "Permissions"},
		"/users/" + asOps:           {"Integrations", "Permissions"},
		"/settings/permissions":     {"Users", "Workflows"},
		"/workflows":                {"Permissions", "Workspaces"},
		"/settings/workspaces":      {"Workflows", "Projects"},
		"/settings/workspaces/acme": {"Workflows", "Projects"},
		"/settings/projects/infra":  {"Workspaces", ""},
	} {
		body := f.page(t, asAdmin, path)
		assert.Contains(t, body, `rel="prev"><svg class="ic back"><use href="#i-chev"/></svg>Previous: `+want[0]+`</a>`, path)
		if want[1] == "" {
			assert.NotContains(t, body, `rel="next">Next:`, path)
			continue
		}
		assert.Contains(t, body, `rel="next">Next: `+want[1]+`<svg`, path)
	}

	// A page outside the steps has no footer.
	assert.NotContains(t, f.page(t, asAdmin, "/runs"), `class="setup-trail"`)
}
