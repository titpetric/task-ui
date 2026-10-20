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
	assert.NoError(t, f.opts.Integrations.Save(context.Background(), integration.Claude, map[string]string{"token": "sk-secret", "folders": "infra"}))

	body := f.page(t, asAdmin, "/settings/integrations")
	assert.Contains(t, body, "Claude Code")
	assert.Contains(t, body, `name="values.folders" value="infra"`)
	assert.Contains(t, body, `<span class="badge ok">set</span>`)
	assert.NotContains(t, body, "sk-secret")
	assert.Contains(t, body, `data-api="/api/integrations/beszel"`)

	assert.Equal(t, http.StatusForbidden, f.get(t, asOps, "/settings/integrations").Status)
	assert.Contains(t, f.page(t, asAdmin, "/settings"), `href="/settings/integrations"`)
	assert.NotContains(t, f.page(t, asAdmin, "/settings/secrets"), "integration:claude")
}

// Hosts shows the Beszel inventory, or how to set it up.
func TestHosts(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.withKey(t)
	assert.Contains(t, f.page(t, asAdmin, "/hosts"), "The Beszel integration is not set up")

	hub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"items":[{"id":"a","name":"dev1","host":"10.10.10.11","status":"up","info":{"cpu":95,"mp":80,"dp":30,"u":7200,"c":4,"m":"EPYC","k":"6.12","v":"0.12"}}]}`))
	}))
	t.Cleanup(hub.Close)
	assert.NoError(t, f.opts.Integrations.Save(context.Background(), integration.Beszel, map[string]string{"url": hub.URL, "token": "t1"}))

	body := f.page(t, asViewer, "/hosts")
	assert.Contains(t, body, "dev1")
	assert.Contains(t, body, `class="meter high"`)
	assert.Contains(t, body, `class="meter warm"`)
	assert.Contains(t, body, "EPYC")
	assert.Contains(t, body, `href="/hosts" class="active"`)
}
