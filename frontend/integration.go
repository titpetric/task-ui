package frontend

import (
	"errors"
	"net/http"
	"time"

	"github.com/titpetric/task-ui/beszel"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/hostinfo"
	"github.com/titpetric/task-ui/integration"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// integrations shows Settings, Integrations: a form for each integration.
// Takes integration/view.
func (h *Handlers) integrations(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "integration/view") {
		return
	}
	p := h.page(r, "Integrations")
	p.Crumbs = view.Crumbs("Settings", "/settings", "Integrations")
	p.Step = view.StepIntegrations
	h.instanceKey(r, p)
	p.Users, _ = h.opts.Users.List(r.Context())
	for _, kind := range integration.Kinds() {
		values, err := h.opts.Integrations.Values(r.Context(), kind.ID)
		if err != nil {
			p.Error = err.Error()
		}
		p.Integrations = append(p.Integrations, view.NewIntegration(kind, values))
	}
	p.Instance = h.opts.Instance
	h.lastSync(r, p)
	h.render(w, http.StatusOK, "integrations", p)
}

// lastSync puts the latest run of @forgejo-sync on the page, for the
// Forgejo form's line on it.
func (h *Handlers) lastSync(r *http.Request, p *view.Page) {
	runs, err := h.opts.Runs.List(r.Context(), storage.RunFilter{Folder: model.Default, Job: folder.ForgejoSyncID, Limit: 1})
	if err != nil || len(runs) == 0 {
		return
	}
	p.Last = map[string]view.Run{model.RunKey(runs[0].Folder, "", runs[0].Job): h.runView(runs[0])}
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
	if len(systems) > 0 {
		h.cluster(r, p)
	}
	p.Summarize = p.Can("hosts") && h.claudeFor(r, p.User)
	if p.Can("hosts") {
		h.lastBatch(r, p)
	}
	h.render(w, http.StatusOK, "inventory", p)
}

// cluster sums the page's hosts into pools: their running containers
// from the hub, and their durable filesystems from what each host last
// told over ssh. With the hosts permission, the reader reads a host
// whose answer is stale again in the background, for the next view.
func (h *Handlers) cluster(r *http.Request, p *view.Page) {
	var containers map[string][]beszel.Container
	var history map[string][]beszel.Sample
	if hub, err := h.opts.Hosts.Hub(r.Context()); err == nil {
		containers, _ = hub.Containers(r.Context())
		history, _ = hub.History(r.Context())
	}
	disks := map[string][]hostinfo.Filesystem{}
	for _, system := range p.Hosts {
		if facts := h.facts(r, p, p.Logins[system.Host], 0); facts != nil {
			disks[system.ID] = facts.Filesystems
		}
	}
	cluster := hostinfo.NewCluster(p.Hosts, containers, disks)
	p.Cluster = &cluster
	if len(history) > 0 {
		p.ClusterCharts = view.ClusterCharts(p.Cluster, p.Hosts, history, p.Now)
	}
}

// facts returns what a login's host last told over ssh, waiting at most
// wait for a fresh answer; only a user with the hosts permission makes
// task-ui read a host. nil without a login or an answer.
func (h *Handlers) facts(r *http.Request, p *view.Page, login string, wait time.Duration) *hostinfo.Facts {
	if h.opts.Facts == nil || login == "" {
		return nil
	}
	if p.Can("hosts") {
		return h.opts.Facts.Facts(r.Context(), login, wait)
	}
	return h.opts.Facts.Cached(login)
}

// lastBatch puts the user's last Run on the inventory on the page: its
// runs, one a host, and the prompt box about them, where Run and
// summarize's thread shows.
func (h *Handlers) lastBatch(r *http.Request, p *view.Page) {
	last, err := h.opts.Runs.List(r.Context(), storage.RunFilter{Folder: model.Default, Job: model.HostAll, UserName: p.User.Name, Limit: 1})
	if err != nil || len(last) == 0 {
		return
	}
	runs := h.batch(r.Context(), last[0].BatchID)
	if len(runs) == 0 {
		runs = last
	}
	for _, run := range runs {
		p.Batch = append(p.Batch, h.framed(h.runView(run)))
	}
	p.Prompt = h.runPrompt(r, p, &p.Batch[0])
}

// claudeFor reports whether the Claude Code integration has a token for
// the user, so the pages show what Claude summarizes.
func (h *Handlers) claudeFor(r *http.Request, user *model.User) bool {
	if h.opts.Integrations == nil || user == nil {
		return false
	}
	env, _ := h.opts.Integrations.Env(r.Context(), user.Name, user.EffectiveRole())
	return env["CLAUDE_CODE_OAUTH_TOKEN"] != ""
}

// hostJobs puts the jobs that run a command on hosts on a page with a
// job picker, for a user with the hosts permission: @host:* and
// @host:<host> for each host task-ui has an ssh login to.
func (h *Handlers) hostJobs(r *http.Request, p *view.Page) {
	if !p.Can("hosts") || h.opts.Hosts == nil || h.opts.SSH == nil {
		return
	}
	systems, err := h.opts.Hosts.Systems(r.Context())
	if err != nil || len(systems) == 0 {
		return
	}
	systems = beszel.ByCapacity(systems)
	hosts := make([]string, len(systems))
	for i, system := range systems {
		hosts[i] = system.Host
	}
	logins := h.opts.SSH.Logins(r.Context(), hosts, 4*time.Second)
	root := &folder.Folder{ID: model.Default}
	for _, host := range hosts {
		if job := root.HostJob(host); job != nil && logins[host] != "" {
			p.HostJobs = append(p.HostJobs, job)
		}
	}
	if len(p.HostJobs) > 0 {
		p.HostJobs = append([]*folder.Job{root.HostJob("*")}, p.HostJobs...)
	}
}
