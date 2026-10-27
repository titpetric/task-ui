package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/beszel"
	"github.com/titpetric/task-ui/claude"
	"github.com/titpetric/task-ui/hostrun"
	"github.com/titpetric/task-ui/integration"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/runner"
)

// integrationView reads an integration's settings for the user: the plain
// ones as they are, the secret ones as set or not.
func (h *Handlers) integrationView(r *http.Request, kind integration.Kind) (IntegrationView, error) {
	values, err := h.opts.Integrations.Values(r.Context(), kind.ID)
	if err != nil {
		return IntegrationView{}, err
	}
	view := IntegrationView{Kind: kind, Configured: integration.Configured(kind, values), Values: map[string]string{}, Set: []string{}, Switches: map[string]bool{}}
	for _, field := range kind.Fields {
		switch {
		case field.Secret && values[field.Name] != "":
			view.Set = append(view.Set, field.Name)
		case !field.Secret:
			view.Values[field.Name] = values[field.Name]
		}
		if field.Toggle {
			view.Switches[field.Name] = integration.Enabled(field, values)
		}
	}
	return view, nil
}

// listIntegrations returns every integration with its settings, secrets
// withheld. Takes integration/view.
func (h *Handlers) listIntegrations(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "integration/view") {
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

// saveIntegration writes an integration's settings, with
// integration/edit. Beszel's are the inventory's source, so they take
// inventory/edit too.
func (h *Handlers) saveIntegration(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "integration/edit") {
		return
	}
	if param(r, "id") == integration.Beszel && !h.require(w, r, "inventory/edit") {
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

// deleteIntegration removes an integration's settings, with
// integration/delete; Beszel's, the inventory's source, with
// inventory/delete too.
func (h *Handlers) deleteIntegration(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "integration/delete") {
		return
	}
	if param(r, "id") == integration.Beszel && !h.require(w, r, "inventory/delete") {
		return
	}
	if err := h.opts.Integrations.Delete(r.Context(), param(r, "id")); err != nil {
		failErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// listHosts returns the inventory: the hosts of the Beszel integration,
// largest first.
func (h *Handlers) listHosts(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "inventory/view") {
		return
	}
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

// runOnHosts runs a command on hosts of the inventory, all at once, a
// run of @host:<host> each over the ssh login task-ui has to the host, under
// one batch, and answers the batch's ID and its runs. With summarize,
// Claude summarizes the batch once every run ended, in a thread about
// its first run. Takes hosts: the logins reach the hosts as debian or
// root.
func (h *Handlers) runOnHosts(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "hosts") {
		return
	}
	var req InventoryRequest
	if err := readJSON(r, &req); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.Summarize && !h.summarizes(w, r, nil) {
		return
	}
	batch, sessions, ok := h.startHosts(w, r, hostrun.Request{Targets: req.Hosts, Command: req.Command, Batch: true}, req.Summarize)
	if !ok {
		return
	}
	views := make([]RunView, 0, len(sessions))
	for _, session := range sessions {
		views = append(views, h.view(r, session.Snapshot()))
	}
	writeJSON(w, http.StatusCreated, map[string]any{"batch": batch, "runs": views})
}

// startHosts starts the runs of @host:<host> a request names, as the user,
// and with summarize asks Claude about them once they ended: about the
// batch when there are more than one, about the run when one. It answers the
// request and returns false when nothing started.
func (h *Handlers) startHosts(w http.ResponseWriter, r *http.Request, req hostrun.Request, summarize bool) (string, []*runner.Session, bool) {
	if h.opts.HostRuns == nil {
		fail(w, http.StatusNotFound, beszel.ErrOff.Error())
		return "", nil, false
	}
	user := access.UserFrom(r.Context())
	batch, sessions, err := h.opts.HostRuns.Start(r.Context(), user, req)
	if err != nil {
		failErr(w, err)
		return "", nil, false
	}
	if summarize && len(sessions) > 0 {
		h.summarizing.Store(sessions[0].ID(), true)
		go h.summarizeHosts(user, sessions)
	}
	return batch, sessions, true
}

// summarizeHosts waits for the host runs to end, then asks Claude for
// their summary as the user who started them: a thread about the first
// run, of the batch when there are more than one.
func (h *Handlers) summarizeHosts(user *model.User, sessions []*runner.Session) {
	for _, session := range sessions {
		<-session.Done()
	}
	first := sessions[0].ID()
	defer h.summarizing.Delete(first)
	ask := h.opts.Board.SummarizeRun
	if len(sessions) > 1 {
		ask = h.opts.Board.SummarizeBatch
	}
	if _, err := ask(context.Background(), user, first); err != nil {
		slog.Warn("host runs summary not asked", "run", first, "err", err)
	}
}

// summarize sends the report to Claude, signed in with the token the
// Claude integration holds for the user, with task-ui's context and
// what the user may do in its system prompt.
func (h *Handlers) summarize(r *http.Request, prompt, report string) (string, error) {
	if h.opts.Summarizer == nil || h.opts.Integrations == nil {
		return "", claude.ErrNoToken
	}
	user := access.UserFrom(r.Context())
	env, err := h.opts.Integrations.Env(r.Context(), user.Name, user.EffectiveRole())
	if err != nil {
		return "", err
	}
	system := claude.Context(user.Name, user.EffectiveRole(), h.permissions(r).Summary())
	return h.opts.Summarizer.Summarize(r.Context(), env, system, prompt, report)
}
