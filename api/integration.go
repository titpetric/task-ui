package api

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/beszel"
	"github.com/titpetric/task-ui/claude"
	"github.com/titpetric/task-ui/integration"
	"github.com/titpetric/task-ui/sshprobe"
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

// runTimeout bounds a command on one host.
const runTimeout = 2 * time.Minute

// runOnHosts runs a command on hosts of the inventory, all at once, over
// the ssh login task-ui has to each, and returns each host's output and
// exit code. Takes hosts: the logins reach the hosts as debian or root.
func (h *Handlers) runOnHosts(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "hosts") {
		return
	}
	if h.opts.Hosts == nil || h.opts.SSH == nil {
		fail(w, http.StatusNotFound, beszel.ErrOff.Error())
		return
	}
	var req InventoryRequest
	if err := readJSON(r, &req); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	if strings.TrimSpace(req.Command) == "" {
		fail(w, http.StatusBadRequest, "enter a command")
		return
	}
	systems, err := h.opts.Hosts.Systems(r.Context())
	if err != nil {
		fail(w, http.StatusBadGateway, err.Error())
		return
	}
	known := map[string]bool{}
	var all []string
	for _, system := range beszel.ByCapacity(systems) {
		known[system.Host] = true
		all = append(all, system.Host)
	}
	hosts := req.Hosts
	if len(hosts) == 0 {
		hosts = all
	}
	for _, host := range hosts {
		if !known[host] {
			fail(w, http.StatusBadRequest, "no host "+host+" in the inventory")
			return
		}
	}
	logins := h.opts.SSH.Logins(r.Context(), hosts, 10*time.Second)
	if len(req.Hosts) == 0 {
		var reachable []string
		for _, host := range hosts {
			if logins[host] != "" {
				reachable = append(reachable, host)
			}
		}
		hosts = reachable
	}
	results := h.opts.SSH.Run(r.Context(), hosts, logins, req.Command, runTimeout)
	body := map[string]any{"results": results}
	if req.Summarize {
		report := sshprobe.Report(req.Command, results)
		body["report"] = report
		summary, err := h.summarize(r, summaryPrompt, report)
		if err != nil {
			body["summary_error"] = err.Error()
		} else {
			body["summary"] = summary
		}
	}
	writeJSON(w, http.StatusOK, body)
}

// summaryPrompt is what Claude reads the report with.
const summaryPrompt = "The input is a markdown report of one shell command run on several hosts over ssh: a section a host, with its exit code and output. " +
	"Summarise it for the operator in one or two short paragraphs: lead with what needs attention (failures, hosts that differ, warnings, anything near a limit), then a line on what is fine, naming the hosts. " +
	"Then, only if it helps, a markdown table with a row a host. No preamble, no headings, don't explain the obvious, and say only what the report shows."

// summarize sends the report to Claude, signed in with the token the
// Claude integration holds for the user, with task-ui's context and
// what the user may do in its system prompt.
func (h *Handlers) summarize(r *http.Request, prompt, report string) (string, error) {
	if h.opts.Summarizer == nil || h.opts.Integrations == nil {
		return "", claude.ErrNoToken
	}
	user := access.UserFrom(r.Context())
	env, err := h.opts.Integrations.Env(r.Context(), user.Name, user.Role)
	if err != nil {
		return "", err
	}
	system := claude.Context(user.Name, user.Role, h.permissions(r).Summary())
	return h.opts.Summarizer.Summarize(r.Context(), env["CLAUDE_CODE_OAUTH_TOKEN"], system, prompt, report)
}
