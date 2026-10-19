package frontend

import (
	"net/http"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/model"
)

// home is the dashboard: every live terminal, and the control centre,
// a button for every job of every folder.
func (h *Handlers) home(w http.ResponseWriter, r *http.Request) {
	user := access.UserFrom(r.Context())
	p := h.page(r, "Dashboard")
	p.Runs = h.live(user)
	p.Control = view.NewControl(p.Folders, view.ControlRepeats)
	schedules, _ := h.opts.Schedules.List(r.Context())
	for _, schedule := range schedules {
		if user.CanSee(schedule.Folder) {
			p.Schedules = append(p.Schedules, schedule)
		}
	}
	h.render(w, http.StatusOK, "home", p)
}

// visibleSchedules returns the schedules of a folder, or of a job of it.
func (h *Handlers) visibleSchedules(r *http.Request, folderID, job string) []model.Schedule {
	schedules, err := h.opts.Schedules.ListByFolder(r.Context(), folderID)
	if err != nil {
		return nil
	}
	if job == "" {
		return schedules
	}
	var result []model.Schedule
	for _, schedule := range schedules {
		if schedule.Job == job {
			result = append(result, schedule)
		}
	}
	return result
}
