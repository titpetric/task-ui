package frontend

import (
	"net/http"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/storage"
)

// job shows a job: what it is, its history and its schedules.
func (h *Handlers) job(w http.ResponseWriter, r *http.Request) {
	f := h.visibleFolder(r)
	if f == nil {
		h.missing(w, r, "No folder named "+param(r, "folder")+".")
		return
	}
	job := f.Job(param(r, "job"))
	if job == nil {
		h.missing(w, r, "No job named "+param(r, "job")+" in "+f.ID+".")
		return
	}
	p := h.page(r, f.ID+" / "+job.Name)
	p.Folder, p.Job = f, job
	p.Crumbs = view.Crumbs(f.ID, "/f/"+f.ID, job.Name)
	p.Runs = h.visibleRuns(r.Context(), access.UserFrom(r.Context()), storage.RunFilter{Folder: f.ID, Job: job.ID, Limit: 100})
	p.Schedules = h.visibleSchedules(r, f.ID, job.ID)
	h.render(w, http.StatusOK, "job", p)
}
