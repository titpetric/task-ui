package api

import (
	"io"
	"net/http"
	"os"
	"path"
	"strconv"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// listArtifacts returns the artifacts the user may see, newest first.
func (h *Handlers) listArtifacts(w http.ResponseWriter, r *http.Request) {
	user := access.UserFrom(r.Context())
	q := r.URL.Query()
	filter := storage.ArtifactFilter{RunID: q.Get("run"), Folder: q.Get("folder"), Job: q.Get("job"), Limit: queryInt(r, "limit", 100)}
	artifacts, err := h.opts.Artifacts.List(r.Context(), filter)
	if err != nil {
		failErr(w, err)
		return
	}
	visible := []model.Artifact{}
	for _, artifact := range artifacts {
		if user.CanSee(artifact.Folder) {
			visible = append(visible, artifact)
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"artifacts": visible})
}

// artifactOf returns the artifact a route names, when the user may see it.
func (h *Handlers) artifactOf(r *http.Request) (*model.Artifact, error) {
	artifact, err := h.opts.Artifacts.Get(r.Context(), param(r, "id"))
	if err != nil {
		return nil, err
	}
	if !access.UserFrom(r.Context()).CanSee(artifact.Folder) {
		return nil, storage.ErrNotFound
	}
	return artifact, nil
}

// downloadArtifact serves an artifact's file.
func (h *Handlers) downloadArtifact(w http.ResponseWriter, r *http.Request) {
	artifact, err := h.artifactOf(r)
	if err != nil {
		failErr(w, err)
		return
	}
	file, err := h.opts.Runner.OpenArtifact(artifact)
	if err != nil {
		fail(w, http.StatusNotFound, "the file is gone")
		return
	}
	defer file.Close()
	w.Header().Set("Content-Type", artifact.ContentType)
	w.Header().Set("Content-Disposition", `attachment; filename="`+path.Base(artifact.Name)+`"`)
	w.Header().Set("Content-Length", strconv.FormatInt(artifact.ByteSize, 10))
	_, _ = io.Copy(w, file)
}

// deleteArtifact removes an artifact and its file. Admins only.
func (h *Handlers) deleteArtifact(w http.ResponseWriter, r *http.Request) {
	if !access.UserFrom(r.Context()).CanAdmin() {
		fail(w, http.StatusForbidden, "admin role required")
		return
	}
	artifact, err := h.artifactOf(r)
	if err != nil {
		failErr(w, err)
		return
	}
	_ = os.Remove(h.opts.Runner.ArtifactPath(artifact))
	if err := h.opts.Artifacts.Delete(r.Context(), artifact.ID); err != nil {
		failErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
