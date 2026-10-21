package frontend

import (
	"net/http"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/storage"
)

// job shows a job: its details, its history and its schedules.
func (h *Handlers) job(w http.ResponseWriter, r *http.Request, f *folder.Folder, id string) {
	job := f.Job(id)
	if job == nil {
		h.missing(w, r, "No job named "+id+" in "+f.ID+".")
		return
	}
	p := h.page(r, f.ID+" / "+job.Name)
	p.Folder, p.Job = f, job
	p.Crumbs = crumbsOf(f.ID, job.Name, p.Workspaces)
	p.Runs = h.visibleRuns(r.Context(), access.UserFrom(r.Context()), storage.RunFilter{Folder: f.ID, Job: job.ID, Limit: 100})
	p.Activity = h.activity(r.Context(), access.UserFrom(r.Context()), storage.RunFilter{Folder: f.ID, Job: job.ID, Limit: activityLimit}, p.Now, true)
	p.Schedules = h.visibleSchedules(r, f.ID, job.ID)
	h.render(w, http.StatusOK, "job", p)
}
