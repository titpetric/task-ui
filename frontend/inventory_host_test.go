package frontend_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/integration"
	"github.com/titpetric/task-ui/internal/assert"
)

// hostHub is a Beszel hub of dev1, which runs db and web, and old, down.
func hostHub(t *testing.T) *httptest.Server {
	t.Helper()
	hub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/collections/system_stats/records":
			_, _ = w.Write([]byte(`{"items":[{"system":"a","created":"2026-10-16 19:00:00.000Z","stats":{"m":8,"mu":4,"d":200,"du":60,"efs":{"sdb":{"d":100,"du":10}}}}]}`))
		case "/api/collections/containers/records":
			_, _ = w.Write([]byte(`{"items":[{"system":"a","name":"web","image":"caddy:2","status":"Up 1 hour","cpu":1.5,"memory":64,"net":2048},{"system":"a","name":"db","cpu":4,"memory":2048}]}`))
		case "/api/collections/systems/records":
			_, _ = w.Write([]byte(`{"items":[{"id":"a","name":"dev1","host":"10.10.10.11","status":"up","info":{"cpu":20,"t":4,"u":259200,"v":"0.18"}},{"id":"b","name":"old","host":"10.10.10.12","status":"down"}]}`))
		default:
			_, _ = w.Write([]byte(`{"items":[]}`))
		}
	}))
	t.Cleanup(hub.Close)
	return hub
}

// A host's page, the expanded row of its line: the containers of both
// sources by CPU, the durable filesystems from df, and for a host
// without a login, what Beszel has and why the rest is missing. Takes
// hosts.
func TestInventoryHost(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.withKey(t)
	hub := hostHub(t)
	assert.NoError(t, f.opts.Integrations.Save(context.Background(), integration.Beszel, map[string]string{"url": hub.URL, "token": "hub-token-secret"}))

	list := f.page(t, asAdmin, "/inventory")
	assert.Contains(t, list, `<tr class="host-row" data-host-row="dev1" tabindex="0" aria-expanded="false">`)
	assert.NotContains(t, f.page(t, asViewer, "/inventory"), "data-host-row")

	body := f.page(t, asAdmin, "/inventory/hosts/dev1")
	assert.Contains(t, body, `data-host-detail="dev1"`)
	assert.Contains(t, body, `<span>ssh <code>debian@10.10.10.11</code></span>`)
	assert.Contains(t, body, `<span>up 3 days</span>`)
	assert.Contains(t, body, "2, measures from Beszel, sizes from docker ps")
	db := `<td data-v="db"><span class="key">db</span><span class="faint under">postgres:17</span></td>`
	web := `<td data-v="web"><span class="key">web</span><span class="faint under">caddy:2</span></td>`
	assert.Contains(t, body, db)
	assert.True(t, strings.Index(body, db) < strings.Index(body, web), "the busiest first")
	assert.Contains(t, body, `2.0 GB<span class="faint under">25.0% of RAM</span>`)
	assert.Contains(t, body, `1000 B<span class="faint under">of 1.9 MB</span>`)
	assert.Contains(t, body, `title="docker ps does not list the container">-</td>`)
	assert.Contains(t, body, `2.0 KB/s`)
	assert.Contains(t, body, "durable storage, from df over ssh")
	assert.Contains(t, body, `<code>/mnt/data</code>`)
	assert.NotContains(t, body, `<code>/run</code>`)
	assert.Contains(t, body, `class="meter high"`)
	assert.NotContains(t, body, "hub-token-secret")

	old := f.page(t, asAdmin, "/inventory/hosts/old")
	assert.Contains(t, old, "no ssh login")
	assert.Contains(t, old, "Neither Beszel nor docker ps lists containers on the host: task-ui has no ssh login to the host, for docker ps.")
	assert.Contains(t, old, "Neither df over ssh nor Beszel reports the host's filesystems: task-ui has no ssh login to it, and Beszel no measures.")

	assert.Equal(t, http.StatusNotFound, f.get(t, asAdmin, "/inventory/hosts/nope").Status)
	assert.Equal(t, http.StatusForbidden, f.get(t, asViewer, "/inventory/hosts/dev1").Status)
	assert.Equal(t, http.StatusForbidden, f.get(t, asOps, "/inventory/hosts/dev1").Status)
}
