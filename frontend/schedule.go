package frontend

import (
	"net/http"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/model"
)

// schedules lists what runs on a schedule, and takes a new one.
func (h *Handlers) schedules(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "schedule/view") {
		return
	}
	user := access.UserFrom(r.Context())
	p := h.page(r, "Scheduled")
	p.Crumbs = view.Crumbs("Scheduled")
	p.Form = map[string]string{"folder": r.URL.Query().Get("folder")}
	schedules, _ := h.opts.Schedules.List(r.Context())
	p.Schedules = []model.Schedule{}
	for _, schedule := range schedules {
		if user.CanSee(schedule.Folder) {
			p.Schedules = append(p.Schedules, schedule)
		}
	}
	h.runAsUsers(r, p)
	h.render(w, http.StatusOK, "schedules", p)
}

// runAsUsers fills the users a schedule may run as, for an admin, who
// picks one: the Run as picker shows only with them.
func (h *Handlers) runAsUsers(r *http.Request, p *view.Page) {
	if p.User.CanAdmin() {
		p.Users, _ = h.opts.Users.List(r.Context())
	}
}
