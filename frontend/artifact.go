package frontend

import (
	"net/http"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// artifacts lists the files the runs left, newest first.
func (h *Handlers) artifacts(w http.ResponseWriter, r *http.Request) {
	user := access.UserFrom(r.Context())
	q := r.URL.Query()
	p := h.page(r, "Artifacts")
	p.Crumbs = view.Crumbs("Artifacts")
	p.Form = map[string]string{"folder": q.Get("folder"), "job": q.Get("job")}
	artifacts, _ := h.opts.Artifacts.List(r.Context(), storage.ArtifactFilter{Folder: q.Get("folder"), Job: q.Get("job"), Limit: 200})
	p.Artifacts = []model.Artifact{}
	for _, artifact := range artifacts {
		if user.CanSee(artifact.Folder) {
			p.Artifacts = append(p.Artifacts, artifact)
		}
	}
	h.render(w, http.StatusOK, "artifacts", p)
}
