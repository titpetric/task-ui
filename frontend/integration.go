package frontend

import (
	"errors"
	"net/http"
	"time"

	"github.com/titpetric/task-ui/beszel"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/integration"
	"github.com/titpetric/task-ui/model"
)

// integrations shows Settings, Integrations: a form for each integration.
// Takes integration/view.
func (h *Handlers) integrations(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "integration/view") {
		return
	}
	p := h.page(r, "Integrations")
	p.Crumbs = view.Crumbs("Settings", "/settings", "Integrations")
	h.instanceKey(p)
	p.Users, _ = h.opts.Users.List(r.Context())
	for _, kind := range integration.Kinds() {
		values, err := h.opts.Integrations.Values(r.Context(), kind.ID)
		if err != nil {
			p.Error = err.Error()
		}
		p.Integrations = append(p.Integrations, view.NewIntegration(kind, values))
	}
	h.render(w, http.StatusOK, "integrations", p)
}

// inventory shows the hosts of the Beszel integration, who task-ui signs
// in to each as, and with the hosts permission the form that runs a
// command on them.
func (h *Handlers) inventory(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "inventory/view") {
		return
	}
	p := h.page(r, "Inventory")
	p.Crumbs = view.Crumbs("Inventory")
	if h.opts.Hosts == nil {
		p.HostsOff = true
		h.render(w, http.StatusOK, "inventory", p)
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
	p.Summarize = p.Can("hosts") && h.claudeFor(r, p.User)
	h.render(w, http.StatusOK, "inventory", p)
}

// claudeFor reports whether the Claude Code integration has a token for
// the user, so the pages show what Claude summarizes.
func (h *Handlers) claudeFor(r *http.Request, user *model.User) bool {
	if h.opts.Integrations == nil || user == nil {
		return false
	}
	env, _ := h.opts.Integrations.Env(r.Context(), user.Name, user.Role)
	return env["CLAUDE_CODE_OAUTH_TOKEN"] != ""
}
