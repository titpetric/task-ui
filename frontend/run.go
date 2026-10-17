package frontend

import (
	"net/http"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/storage"
)

// runs lists every run the user may see, filtered by the query.
func (h *Handlers) runs(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	p := h.page(r, "Runs")
	p.Crumbs = view.Crumbs("Runs")
	p.Form = map[string]string{"folder": q.Get("folder"), "job": q.Get("job"), "state": q.Get("state")}
	filter := storage.RunFilter{Folder: q.Get("folder"), Job: q.Get("job"), State: q.Get("state"), Limit: 200}
	p.Runs = h.visibleRuns(r.Context(), access.UserFrom(r.Context()), filter)
	h.render(w, http.StatusOK, "runs", p)
}

// run shows one run: its terminal while it is live, its details and its
// artifacts.
func (h *Handlers) run(w http.ResponseWriter, r *http.Request) {
	user := access.UserFrom(r.Context())
	row, err := h.opts.Runs.Get(r.Context(), param(r, "id"))
	if err != nil || !user.CanSee(row.Folder) {
		h.missing(w, r, "No run with that ID.")
		return
	}
	run := h.runView(*row)
	p := h.page(r, run.Folder+" / "+run.Job)
	p.Run = &run
	p.Crumbs = view.Crumbs("Runs", "/runs", run.Folder+" / "+run.Job)
	if f := h.opts.Catalogue.Find(run.Folder); f != nil {
		p.Folder, p.Job = f, f.Job(run.Job)
	}
	artifacts, _ := h.opts.Artifacts.List(r.Context(), storage.ArtifactFilter{RunID: run.ID, Limit: 200})
	p.Artifacts = artifacts
	h.render(w, http.StatusOK, "run", p)
}
