package frontend

import (
	"errors"
	"net/http"
	"time"

	"github.com/titpetric/task-ui/beszel"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/integration"
)

// integrations shows Settings, Integrations: a form for each integration.
// Admins only.
func (h *Handlers) integrations(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	p := h.page(r, "Integrations")
	p.Crumbs = view.Crumbs("Settings", "/settings", "Integrations")
	p.Encryption = h.opts.Secrets.Vault.Recipients()
	for _, kind := range integration.Kinds() {
		values, err := h.opts.Integrations.Values(r.Context(), kind.ID)
		if err != nil {
			p.Error = err.Error()
		}
		p.Integrations = append(p.Integrations, view.NewIntegration(kind, values))
	}
	h.render(w, http.StatusOK, "integrations", p)
}

// hosts shows the host inventory from the Beszel integration.
func (h *Handlers) hosts(w http.ResponseWriter, r *http.Request) {
	p := h.page(r, "Hosts")
	p.Crumbs = view.Crumbs("Hosts")
	if h.opts.Hosts == nil {
		p.HostsOff = true
		h.render(w, http.StatusOK, "hosts", p)
		return
	}
	systems, err := h.opts.Hosts.Systems(r.Context())
	switch {
	case errors.Is(err, beszel.ErrOff):
		p.HostsOff = true
	case err != nil:
		p.Error = err.Error()
	}
	p.Hosts = beszel.ByCapacity(systems)
	if h.opts.SSH != nil && len(systems) > 0 {
		hosts := make([]string, len(systems))
		for i, system := range systems {
			hosts[i] = system.Host
		}
		p.Logins = h.opts.SSH.Logins(r.Context(), hosts, 4*time.Second)
	}
	h.render(w, http.StatusOK, "hosts", p)
}
