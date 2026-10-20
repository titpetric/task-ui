package frontend_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/titpetric/task-ui/integration"
	"github.com/titpetric/task-ui/internal/assert"
)

// Settings, Integrations shows a form per integration: a secret only as
// set, never its value. Admins only.
func TestIntegrations(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.withKey(t)
	assert.NoError(t, f.opts.Integrations.Save(context.Background(), integration.Claude, map[string]string{"token": "sk-secret", "who": "admin"}))

	body := f.page(t, asAdmin, "/settings/integrations")
	assert.Contains(t, body, "Claude Code")
	assert.Contains(t, body, `name="values.who" value="admin"`)
	assert.Contains(t, body, `<span class="badge ok">set</span>`)
	assert.NotContains(t, body, "sk-secret")
	assert.Contains(t, body, `data-api="/api/integrations/beszel"`)

	assert.Equal(t, http.StatusForbidden, f.get(t, asOps, "/settings/integrations").Status)
	assert.Contains(t, f.page(t, asAdmin, "/settings"), `href="/settings/integrations"`)
	assert.NotContains(t, f.page(t, asAdmin, "/settings/secrets"), "integration:claude")
}

// Inventory shows the Beszel hosts, or how to set them up, and the run
// panel for an admin.
func TestInventory(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.withKey(t)
	assert.Contains(t, f.page(t, asAdmin, "/inventory"), "The Beszel integration is not set up")

	hub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/collections/system_stats/records" {
			_, _ = w.Write([]byte(`{"items":[{"system":"a","stats":{"m":15.6,"mu":12.4,"d":200,"du":60}}]}`))
			return
		}
		_, _ = w.Write([]byte(`{"items":[{"id":"a","name":"dev1","host":"10.10.10.11","status":"up","info":{"cpu":95,"mp":80,"dp":30,"u":7200,"t":4,"m":"EPYC","k":"6.12","v":"0.12"}}]}`))
	}))
	t.Cleanup(hub.Close)
	assert.NoError(t, f.opts.Integrations.Save(context.Background(), integration.Beszel, map[string]string{"url": hub.URL, "token": "t1"}))

	admin := f.page(t, asAdmin, "/inventory")
	assert.Contains(t, admin, `data-inventory-run`)
	assert.Contains(t, admin, `name="hosts" value="10.10.10.11" checked> <code>debian@10.10.10.11</code>`)
	assert.Contains(t, admin, `data-command="uptime"`)
	assert.NotContains(t, admin, `name="summarize"`)
	assert.NoError(t, f.opts.Integrations.Save(context.Background(), integration.Claude, map[string]string{"token": "sk-1", "who": "admin"}))
	assert.Contains(t, f.page(t, asAdmin, "/inventory"), `name="summarize" value="1"`)
	body := f.page(t, asViewer, "/inventory")
	assert.NotContains(t, body, `data-inventory-run`)
	assert.Contains(t, body, "dev1")
	assert.Contains(t, body, `class="meter high"`)
	assert.Contains(t, body, `class="meter warm"`)
	assert.Contains(t, body, `<td class="center">4</td>`)
	assert.Contains(t, body, `<code>debian@10.10.10.11</code>`)
	assert.Contains(t, body, `data-copy="ssh debian@10.10.10.11"`)
	assert.Contains(t, body, `<td class="key nowrap" title="10.10.10.11"><svg class="ic"><use href="#i-server"/></svg> dev1</td>`)
	assert.Contains(t, body, "12.4 / 15.6 GB")
	assert.Contains(t, body, "60 / 200 GB")
	assert.NotContains(t, body, "EPYC")
	assert.Contains(t, body, `href="/inventory" class="active"`)
	assert.Equal(t, http.StatusNotFound, f.get(t, asViewer, "/hosts").Status)
}
