package frontend

import (
	"net/http"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/storage"
)

// job shows a job: its details, its history and its schedules. f is the
// folder of the directory a job of a directory inside a folder is from.
func (h *Handlers) job(w http.ResponseWriter, r *http.Request, f *folder.Folder, id string) {
	job := f.Job(id)
	if job == nil {
		h.missing(w, r, "No job named "+id+" in "+f.ID+".")
		return
	}
	where := f.ID
	if job.Dir != "" {
		where += "/" + job.Dir
	}
	p := h.page(r, where+" / "+job.Name)
	p.Folder, p.Job = f, job
	p.Crumbs = crumbsOf(where, job.Name, p.Workspaces)
	p.Runs = h.visibleRuns(r.Context(), h.runSeer(r), storage.RunFilter{Folder: f.ID, Dir: job.Dir, Job: job.ID, Limit: 100})
	p.Activity = h.activity(r.Context(), h.runSeer(r), storage.RunFilter{Folder: f.ID, Dir: job.Dir, Job: job.ID, Limit: activityLimit}, p.Now, true)
	// A schedule names a folder's job: a directory's job has none.
	if job.Dir == "" {
		p.Schedules = h.visibleSchedules(r, f.ID, job.ID)
	}
	h.render(w, http.StatusOK, "job", p)
}
