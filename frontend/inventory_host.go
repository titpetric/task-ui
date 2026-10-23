package frontend

import (
	"net/http"
	"time"

	"github.com/titpetric/task-ui/beszel"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/hostinfo"
)

// inventoryHost shows a host of the inventory in full, the expanded row
// of its line on the Inventory page, which fetches this page: its
// containers, its durable filesystems and its charts. Takes hosts: the
// facts come over the host's ssh login.
func (h *Handlers) inventoryHost(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "hosts") {
		return
	}
	if h.opts.Hosts == nil {
		h.missing(w, r, beszel.ErrOff.Error())
		return
	}
	systems, err := h.opts.Hosts.Systems(r.Context())
	if err != nil {
		h.missing(w, r, err.Error())
		return
	}
	name := param(r, "name")
	var system *beszel.System
	for i := range systems {
		if systems[i].Name == name {
			system = &systems[i]
		}
	}
	if system == nil {
		h.missing(w, r, "Beszel monitors no host "+name+".")
		return
	}
	p := h.page(r, name)
	p.Crumbs = view.Crumbs("Inventory", "/inventory", name)
	detail := &view.HostDetail{System: *system, Unsized: "task-ui has no ssh login to the host, for docker ps"}
	if h.opts.SSH != nil {
		detail.Login = h.opts.SSH.Logins(r.Context(), []string{system.Host}, 4*time.Second)[system.Host]
	}
	var memory float64
	if system.Stats != nil {
		memory = system.Stats.Memory
	}
	var measured []beszel.Container
	if hub, err := h.opts.Hosts.Hub(r.Context()); err == nil {
		containers, _ := hub.Containers(r.Context())
		measured, detail.Measured = containers[system.ID], len(containers[system.ID]) > 0
	}
	var sized []hostinfo.DockerContainer
	facts := h.facts(r, p, detail.Login, 15*time.Second)
	switch {
	case detail.Login == "":
	case facts == nil:
		detail.Unsized = "the host has not answered over ssh yet"
	case facts.Err != nil:
		detail.Unsized = "ssh: " + facts.Err.Error()
	case !facts.Docker:
		detail.Unsized = "docker ps failed for " + detail.Login + ", and through sudo"
	default:
		sized, detail.Sized = facts.Containers, true
	}
	if facts != nil {
		detail.Read = facts.At
	}
	detail.Containers = hostinfo.Merge(measured, sized, memory)
	if facts != nil && len(facts.Filesystems) > 0 {
		detail.Filesystems, detail.FromDF = facts.Filesystems, true
	} else {
		detail.Filesystems = hostinfo.BeszelFilesystems(system.Stats)
	}
	p.HostDetail = detail
	h.render(w, http.StatusOK, "inventory-host", p)
}
