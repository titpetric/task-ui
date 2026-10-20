package api_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
)

// An admin saves an integration; the API shows its plain settings and of
// a secret only whether it has a value, and the secrets list leaves it out.
func TestIntegration_SaveListDelete(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.withKey(t)

	denied := f.call(t, asOps, http.MethodPut, "/integrations/claude", api.IntegrationRequest{Values: map[string]string{"token": "sk-1"}})
	assert.Equal(t, http.StatusForbidden, denied.Status)

	saved := f.call(t, asAdmin, http.MethodPut, "/integrations/claude", api.IntegrationRequest{Values: map[string]string{"token": "sk-1", "who": "admin"}})
	assert.Equal(t, http.StatusOK, saved.Status, string(saved.Raw))
	assert.False(t, strings.Contains(string(saved.Raw), "sk-1"), "the token never comes back")
	assert.Contains(t, string(saved.Raw), `"set":["token"]`)
	assert.Contains(t, string(saved.Raw), `"configured":true`)

	list := f.call(t, asAdmin, http.MethodGet, "/integrations", nil)
	assert.Equal(t, http.StatusOK, list.Status)
	assert.Len(t, list.list("integrations"), 2)
	assert.False(t, strings.Contains(string(list.Raw), "sk-1"))
	assert.Len(t, f.call(t, asAdmin, http.MethodGet, "/secrets", nil).list("secrets"), 0)

	unknown := f.call(t, asAdmin, http.MethodPut, "/integrations/nothing", api.IntegrationRequest{})
	assert.Equal(t, http.StatusBadRequest, unknown.Status)
	assert.Equal(t, http.StatusNoContent, f.call(t, asAdmin, http.MethodDelete, "/integrations/claude", nil).Status)
}

// The hosts come from the Beszel hub the integration names.
func TestIntegration_Hosts(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.withKey(t)
	off := f.call(t, asViewer, http.MethodGet, "/inventory", nil)
	assert.Equal(t, http.StatusNotFound, off.Status)

	hub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "t1" {
			http.Error(w, "{}", http.StatusUnauthorized)
			return
		}
		_, _ = w.Write([]byte(`{"items":[{"id":"a","name":"dev1","host":"dev1","status":"up","info":{"cpu":5,"mp":40,"dp":30}}]}`))
	}))
	t.Cleanup(hub.Close)
	f.call(t, asAdmin, http.MethodPut, "/integrations/beszel", api.IntegrationRequest{Values: map[string]string{"url": hub.URL, "token": "t1"}})

	hosts := f.call(t, asViewer, http.MethodGet, "/inventory", nil)
	assert.Equal(t, http.StatusOK, hosts.Status, string(hosts.Raw))
	assert.Len(t, hosts.list("hosts"), 1)

	// An admin runs a command on the hosts; nobody else may.
	denied := f.call(t, asOps, http.MethodPost, "/inventory/run", api.InventoryRequest{Command: "uptime"})
	assert.Equal(t, http.StatusForbidden, denied.Status)
	ran := f.call(t, asAdmin, http.MethodPost, "/inventory/run", api.InventoryRequest{Command: "echo hi from $(echo dev1)"})
	assert.Equal(t, http.StatusOK, ran.Status, string(ran.Raw))
	results := ran.list("results")
	assert.Len(t, results, 1)
	first := results[0].(map[string]any)
	assert.Equal(t, "debian@dev1", first["login"])
	assert.Equal(t, "hi from dev1\n", first["output"])
	empty := f.call(t, asAdmin, http.MethodPost, "/inventory/run", api.InventoryRequest{Command: " "})
	assert.Equal(t, http.StatusBadRequest, empty.Status)
	stranger := f.call(t, asAdmin, http.MethodPost, "/inventory/run", api.InventoryRequest{Command: "uptime", Hosts: []string{"elsewhere"}})
	assert.Equal(t, http.StatusBadRequest, stranger.Status)

	f.call(t, asAdmin, http.MethodPut, "/integrations/beszel", api.IntegrationRequest{Values: map[string]string{"url": hub.URL, "token": "expired"}})
	assert.Equal(t, http.StatusBadGateway, f.call(t, asViewer, http.MethodGet, "/inventory", nil).Status)
}
