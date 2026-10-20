package api

import (
	"errors"
	"net/http"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/storage"
)

// listFolders returns the folders the user may see with their jobs, each
// with whether the user may run it and its latest run. ?folder= keeps
// one folder.
func (h *Handlers) listFolders(w http.ResponseWriter, r *http.Request) {
	user := access.UserFrom(r.Context())
	only := r.URL.Query().Get("folder")
	views := []FolderView{}
	for _, f := range h.opts.Catalogue.List() {
		if !user.CanSee(f.ID) || only != "" && f.ID != only {
			continue
		}
		view := FolderView{ID: f.ID, Title: f.Title, Description: f.Description, Error: f.Error, Jobs: []JobView{}}
		for _, job := range f.Jobs {
			last, err := h.opts.Runs.Last(r.Context(), storage.RunFilter{Folder: f.ID, Job: job.ID})
			if err != nil && !errors.Is(err, storage.ErrNotFound) {
				failErr(w, err)
				return
			}
			view.Jobs = append(view.Jobs, JobView{Job: job, CanRun: canRun(user, f, job), LastRun: last})
		}
		views = append(views, view)
	}
	writeJSON(w, http.StatusOK, map[string]any{"folders": views})
}
