package frontend

import (
	"net/http"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// visibleFolder returns the folder a route names when the user may see
// it, or nil.
func (h *Handlers) visibleFolder(r *http.Request) *folder.Folder {
	f := h.opts.Catalogue.Find(param(r, "folder"))
	if f == nil || !access.UserFrom(r.Context()).CanSee(f.ID) {
		return nil
	}
	return f
}

// folder shows a folder: its jobs, what it is, its env files.
func (h *Handlers) folder(w http.ResponseWriter, r *http.Request) {
	f := h.visibleFolder(r)
	if f == nil {
		h.missing(w, r, "No folder named "+param(r, "folder")+".")
		return
	}
	p := h.page(r, f.Title)
	p.Folder = f
	p.Crumbs = view.Crumbs(f.ID)
	if h.opts.Git != nil && h.opts.Git.IsRepo(f.Dir) {
		p.Checkout = h.opts.Git.Info(f.Dir)
	}
	if h.opts.Env != nil {
		p.EnvFiles = h.opts.Env.Files(r.Context(), f)
	}
	p.Schedules = h.visibleSchedules(r, f.ID, "")
	p.Last = h.lastRuns(r, f)
	h.render(w, http.StatusOK, "folder", p)
}

// lastRuns returns the latest run of each job of a folder.
func (h *Handlers) lastRuns(r *http.Request, f *folder.Folder) map[string]view.Run {
	runs, err := h.opts.Runs.List(r.Context(), storage.RunFilter{Folder: f.ID, Limit: 20 * len(f.Jobs)})
	if err != nil {
		return nil
	}
	last := map[string]view.Run{}
	for _, run := range runs {
		if _, seen := last[run.Key()]; !seen && run.State != model.StateRunning {
			last[run.Key()] = h.runView(run)
		}
	}
	return last
}
