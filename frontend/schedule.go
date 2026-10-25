package frontend

import (
	"net/http"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/model"
)

// schedules lists what runs on a schedule, the jobs and the workflows,
// and takes a new one of a job; a workflow's comes from its edit page.
func (h *Handlers) schedules(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "schedule/view") {
		return
	}
	user := access.UserFrom(r.Context())
	p := h.page(r, "Scheduled")
	p.Crumbs = view.Crumbs("Scheduled")
	p.Form = map[string]string{"folder": r.URL.Query().Get("folder")}
	p.WorkflowNames = map[string]string{}
	for _, wf := range p.SideWorkflows {
		p.WorkflowNames[wf.ID] = wf.Name
	}
	schedules, _ := h.opts.Schedules.List(r.Context())
	p.Schedules = []model.Schedule{}
	for _, schedule := range schedules {
		_, sees := p.WorkflowNames[schedule.WorkflowID]
		if schedule.IsWorkflow() && sees || !schedule.IsWorkflow() && user.CanSee(schedule.Folder) {
			p.Schedules = append(p.Schedules, schedule)
		}
	}
	h.runAsUsers(r, p)
	h.hostJobs(r, p)
	h.render(w, http.StatusOK, "schedules", p)
}

// runAsUsers fills the users a schedule may run as, for an admin, who
// picks one: the Runs as picker shows only with them, each user with the
// instance SSH keys their runs get, read from the key folder once.
func (h *Handlers) runAsUsers(r *http.Request, p *view.Page) {
	if !p.User.CanAdmin() {
		return
	}
	p.Users, _ = h.opts.Users.List(r.Context())
	if p.UserKeys == nil {
		p.UserKeys = h.runAsKeys(r, p.Users)
	}
}

// runAsKeys returns the instance SSH keys each user's runs get, by name;
// nil without a key folder.
func (h *Handlers) runAsKeys(r *http.Request, users []model.User) map[string][]string {
	if h.opts.Keys == nil {
		return nil
	}
	keys, _ := h.opts.Keys.List(r.Context())
	return view.UserKeys(keys, users)
}
