//go:build integration

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
	assert.NoError(t, f.opts.Workspaces.EnsureDefault(context.Background()), "a fresh instance has the default workspace")
	assert.NoError(t, f.opts.Secrets.Set(context.Background(), &model.Secret{Scope: "all", Name: "TOKEN", Content: "x"}))

	body := f.page(t, asAdmin, "/settings")
	assert.Contains(t, body, "<h2>Settings</h2>")
	assert.Contains(t, body, "2 folders")
	assert.Contains(t, body, "2 keys<", "the SSH key and the encryption key")
	assert.Contains(t, body, "1 variable<")
	assert.Contains(t, body, "4 users")
	// The key folder's path is the Keys page's to show.
	assert.NotContains(t, body, f.opts.Keys.Dir)
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
	assert.Contains(t, body, "/api/mcp")
	assert.NotContains(t, body, "TASK_UI_AGENT")
	assert.Contains(t, body, `<th>Variable</th>`)
	assert.Contains(t, body, `<td class="var" data-label="Variable"><code>TASK_UI_WORKSPACE</code></td>`)
	assert.Contains(t, body, `<span class="badge ok">on</span></td><td class="var" data-label="Variable"><span class="faint">none</span></td><td class="desc" data-label="About"><span>Hosts from the Beszel`)
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
		"/settings/keys":            {"Telemetry", "Secrets"},
		"/settings/keys/generate":   {"Telemetry", "Secrets"},
		"/settings/secrets":         {"Keys", "Integrations"},
		"/settings/integrations":    {"Secrets", "Users"},
		"/settings/users":           {"Integrations", "Permissions"},
		"/settings/permissions":     {"Users", "Workflows"},
		"/settings/workspaces":      {"Workflows", "Projects"},
		"/settings/workspaces/acme": {"Workflows", "Projects"},
		"/settings/projects/infra":  {"Workspaces", "Sidebar"},
		"/settings/sidebar":         {"Projects", ""},
	} {
		body := f.page(t, asAdmin, path)
		assert.Contains(t, body, `rel="prev"><svg class="ic back"><use href="#i-chev"/></svg>Previous: `+want[0]+`</a>`, path)
		if want[1] == "" {
			assert.NotContains(t, body, `rel="next">Next:`, path)
			continue
		}
		assert.Contains(t, body, `rel="next">Next: `+want[1]+`<svg`, path)
	}

	// A page outside the settings has no footer, the Workflows page, a
	// profile and a user's edit page among them.
	for _, path := range []string{"/runs", "/workflows", "/workflows/new", "/users/" + asOps, "/users/" + asOps + "/edit"} {
		assert.NotContains(t, f.page(t, asAdmin, path), `class="setup-trail"`, path)
	}
}
