package frontend

import (
	"net/http"
	"strings"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// folderPath serves everything under /f: the folder a path names, a
// job of it, or one of its env files. Folders nest, so the longest
// folder that starts the path is the one meant.
func (h *Handlers) folderPath(w http.ResponseWriter, r *http.Request) {
	path := param(r, "*")
	f, rest := folder.Resolve(h.opts.Catalogue, path)
	if f == nil || !access.UserFrom(r.Context()).CanSee(f.ID) {
		h.missing(w, r, "No folder named "+path+".")
		return
	}
	switch {
	case rest == "":
		h.folder(w, r, f)
	case strings.HasPrefix(rest, "env/"):
		h.envFile(w, r, f, strings.TrimPrefix(rest, "env/"))
	case !strings.Contains(rest, "/"):
		h.job(w, r, f, rest)
	default:
		h.missing(w, r, "There is nothing at /f/"+path+".")
	}
}

// folder shows a folder: its jobs, what it is, its env files.
func (h *Handlers) folder(w http.ResponseWriter, r *http.Request, f *folder.Folder) {
	p := h.page(r, f.Title)
	p.Folder = f
	p.Crumbs = crumbsOf(f.ID, "")
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
