package frontend

import (
	"net/http"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/note"
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
	p.Summaries = h.summaries(r, p)
	p.Crumbs = crumbsOf(where, job.Name, p.Workspaces)
	p.Runs = h.visibleRuns(r.Context(), h.runSeer(r), storage.RunFilter{Folder: f.ID, Dir: job.Dir, Job: job.ID, Limit: 100})
	p.RunReports = h.runReports(r, p)
	p.Activity = h.activity(r.Context(), h.runSeer(r), storage.RunFilter{Folder: f.ID, Dir: job.Dir, Job: job.ID, Limit: activityLimit}, p.Now, true)
	// A schedule names a folder's job: a directory's job has none.
	if job.Dir == "" {
		p.Schedules = h.visibleSchedules(r, f.ID, job.ID)
		h.runAsUsers(r, p)
	}
	h.render(w, http.StatusOK, "job", p)
}

// runReports returns the summaries of the runs a page lists, by run ID,
// for a user with prompt/view.
func (h *Handlers) runReports(r *http.Request, p *view.Page) map[string]*note.Report {
	if h.opts.Board == nil || !p.Can("prompt/view") {
		return nil
	}
	ids := make([]string, 0, len(p.Runs))
	for _, run := range p.Runs {
		ids = append(ids, run.ID)
	}
	return h.opts.Board.Summaries(r.Context(), ids)
}
