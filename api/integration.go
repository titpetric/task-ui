package api

import (
	"errors"
	"net/http"

	"github.com/titpetric/task-ui/beszel"
	"github.com/titpetric/task-ui/integration"
)

// integrationView reads an integration's settings for the user: the plain
// ones as they are, the secret ones as set or not.
func (h *Handlers) integrationView(r *http.Request, kind integration.Kind) (IntegrationView, error) {
	values, err := h.opts.Integrations.Values(r.Context(), kind.ID)
	if err != nil {
		return IntegrationView{}, err
	}
	view := IntegrationView{Kind: kind, Configured: integration.Configured(kind, values), Values: map[string]string{}, Set: []string{}}
	for _, field := range kind.Fields {
		switch {
		case field.Secret && values[field.Name] != "":
			view.Set = append(view.Set, field.Name)
		case !field.Secret:
			view.Values[field.Name] = values[field.Name]
		}
	}
	return view, nil
}

// listIntegrations returns every integration with its settings, secrets
// withheld. Admins only.
func (h *Handlers) listIntegrations(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	views := []IntegrationView{}
	for _, kind := range integration.Kinds() {
		view, err := h.integrationView(r, kind)
		if err != nil {
			failErr(w, err)
			return
		}
		views = append(views, view)
	}
	writeJSON(w, http.StatusOK, map[string]any{"integrations": views})
}

// saveIntegration writes an integration's settings. Admins only.
func (h *Handlers) saveIntegration(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	var req IntegrationRequest
	if err := readJSON(r, &req); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	id := param(r, "id")
	if err := h.opts.Integrations.Save(r.Context(), id, req.Values); err != nil {
		failErr(w, err)
		return
	}
	view, err := h.integrationView(r, *integration.KindOf(id))
	if err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, view)
}

// deleteIntegration removes an integration's settings. Admins only.
func (h *Handlers) deleteIntegration(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	if err := h.opts.Integrations.Delete(r.Context(), param(r, "id")); err != nil {
		failErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// listHosts returns the host inventory from the Beszel integration.
func (h *Handlers) listHosts(w http.ResponseWriter, r *http.Request) {
	if h.opts.Hosts == nil {
		fail(w, http.StatusNotFound, beszel.ErrOff.Error())
		return
	}
	systems, err := h.opts.Hosts.Systems(r.Context())
	switch {
	case errors.Is(err, beszel.ErrOff):
		fail(w, http.StatusNotFound, err.Error())
		return
	case err != nil:
		fail(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"hosts": beszel.ByCapacity(systems)})
}
