package frontend

import (
	"net/http"

	"github.com/titpetric/task-ui/model"
)

// home is the dashboard: every live terminal.
func (h *Handlers) home(w http.ResponseWriter, r *http.Request) {
	p := h.page(r, "Dashboard")
	p.Runs = h.live(h.runSeer(r))
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
