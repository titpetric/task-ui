package frontend

import (
	"net/http"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/model"
)

// home is the dashboard: every live terminal.
func (h *Handlers) home(w http.ResponseWriter, r *http.Request) {
	p := h.page(r, "Dashboard")
	p.Runs = h.live(h.runSeer(r))
	h.render(w, http.StatusOK, "home", p)
}

// latestRuns returns the latest finished run of each job of the folders,
// by job key, for the Jobs table to colour its buttons by.
func (h *Handlers) latestRuns(r *http.Request, folders []*folder.Folder) map[string]view.Run {
	if len(folders) == 0 {
		return nil
	}
	ids := make([]string, 0, len(folders))
	for _, f := range folders {
		ids = append(ids, f.ID)
	}
	runs, err := h.opts.Runs.Latest(r.Context(), ids)
	if err != nil {
		return nil
	}
	last := make(map[string]view.Run, len(runs))
	for _, run := range runs {
		last[run.Key()] = h.runView(run)
	}
	return last
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
