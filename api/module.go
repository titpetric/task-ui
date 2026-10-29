package api

import (
	"net/http"

	"github.com/titpetric/task-ui/access"
)

// ModuleView is a module as the API answers it: what its modinfo.php
// registered, and where it opens.
//
// The root is left out. It is a path on the server's disk, and a module's
// page is reached by its link.
type ModuleView struct {
	Name        string `json:"name"`
	Vendor      string `json:"vendor"`
	Plugin      string `json:"plugin"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Icon        string `json:"icon,omitempty"`
	Version     string `json:"version,omitempty"`
	Href        string `json:"href"`
}

// listModules answers the modules the instance found, by name.
func (h *Handlers) listModules(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "module") {
		return
	}
	found := h.opts.Modules.Modules()
	out := make([]ModuleView, 0, len(found))
	for _, m := range found {
		out = append(out, ModuleView{
			Name: m.Name(), Vendor: m.Vendor, Plugin: m.Plugin,
			Title: m.Label(), Description: m.Description, Icon: m.Icon,
			Version: m.Version, Href: m.Href(),
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"modules": out})
}

// reloadModules reads every root again and runs each modinfo.php, for
// admins alone: a module added to a checkout appears without a restart,
// and one whose directory is gone loses its rows.
//
// A module that will not register is reported in the log and left out,
// the way a start reports it, so this answers what did register rather
// than the first failure.
func (h *Handlers) reloadModules(w http.ResponseWriter, r *http.Request) {
	if !access.UserFrom(r.Context()).CanAdmin() {
		fail(w, http.StatusForbidden, "only an admin reloads the modules")
		return
	}
	if h.opts.Modules == nil {
		fail(w, http.StatusNotFound, "this instance serves no modules")
		return
	}
	if err := h.opts.Modules.Reload(r.Context()); err != nil {
		failErr(w, err)
		return
	}
	h.listModules(w, r)
}
