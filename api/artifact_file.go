package api

import (
	"errors"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/artifactstore"
	"github.com/titpetric/task-ui/markdown"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/storage"
)

// maxPreview bounds the text and the markdown the artifact window shows.
const maxPreview = 1 << 20

// imageTypes are the extensions the artifact window shows as an image.
var imageTypes = map[string]bool{".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".webp": true, ".svg": true}

// viewArtifactFile returns a file of the artifact tree for the artifact
// window: the version ?version= names, the latest without, and the
// versions the user may see.
func (h *Handlers) viewArtifactFile(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "artifact/view") {
		return
	}
	store := h.opts.ArtifactStore
	rel := folderParam(r)
	if store == nil || rel == "" {
		fail(w, http.StatusNotFound, artifactstore.ErrNotFound.Error())
		return
	}
	user := access.UserFrom(r.Context())
	records, err := h.opts.Artifacts.List(r.Context(), storage.ArtifactFilter{Path: rel, Limit: 200})
	if err != nil {
		failErr(w, err)
		return
	}
	view := ArtifactFileView{Path: rel, ContentType: runner.ContentType(rel), Versions: []ArtifactVersion{}}
	// A run that writes the same content again leaves a record on the
	// commit before; the newest record of a commit stands for it.
	seen := map[string]bool{}
	for _, record := range records {
		if record.CommitID != "" && !seen[record.CommitID] && user.SeesArtifact(&record) {
			seen[record.CommitID] = true
			view.Versions = append(view.Versions, ArtifactVersion{Commit: record.CommitID, CreatedAt: record.CreatedAt, Folder: record.Folder, Job: record.Job, RunID: record.RunID, UserName: record.UserName, ByteSize: record.ByteSize})
		}
	}
	if !user.CanAdmin() && len(view.Versions) == 0 {
		fail(w, http.StatusNotFound, artifactstore.ErrNotFound.Error())
		return
	}

	var data []byte
	view.Version = r.URL.Query().Get("version")
	if view.Version != "" {
		if !user.CanAdmin() && !hasVersion(view.Versions, view.Version) {
			fail(w, http.StatusNotFound, artifactstore.ErrNotFound.Error())
			return
		}
		data, err = store.Show(r.Context(), view.Version, rel)
	} else {
		var file string
		if file, err = store.Path(rel); err == nil {
			data, err = os.ReadFile(file)
		}
		if len(view.Versions) > 0 {
			view.Version = view.Versions[0].Commit
			if err != nil {
				// Gone from the tree: the latest version is in the history.
				data, err = store.Show(r.Context(), view.Version, rel)
			}
		}
	}
	if err != nil {
		fail(w, http.StatusNotFound, artifactstore.ErrNotFound.Error())
		return
	}
	view.ByteSize = int64(len(data))
	if _, raw := r.URL.Query()["raw"]; raw {
		// The file itself, for a client that reads it, as MCP's
		// read_artifact does.
		w.Header().Set("Content-Type", view.ContentType)
		w.Header().Set("X-Artifact-Version", view.Version)
		w.Header().Set("Content-Length", strconv.Itoa(len(data)))
		_, _ = w.Write(data)
		return
	}
	ext := strings.ToLower(path.Ext(rel))
	switch {
	case imageTypes[ext]:
		view.Kind = "image"
	case len(data) > maxPreview || !utf8.Valid(data):
		view.Kind = "file"
	case ext == ".md" || ext == ".markdown":
		view.Kind = "markdown"
		view.HTML = markdown.HTML(data, "/artifacts/files/", path.Dir(rel))
	default:
		view.Kind = "text"
		view.Text = string(data)
	}
	writeJSON(w, http.StatusOK, view)
}

// maxArtifactWrite bounds a file written into the artifact tree.
const maxArtifactWrite = 16 << 20

// writeArtifactFile puts the request's body into the artifact tree at
// the path, as the user with no run, over the version there; the history
// keeps that one. Takes artifact/edit.
func (h *Handlers) writeArtifactFile(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "artifact/edit") {
		return
	}
	if h.opts.ArtifactStore == nil {
		fail(w, http.StatusNotFound, artifactstore.ErrNotFound.Error())
		return
	}
	defer r.Body.Close()
	data, err := io.ReadAll(io.LimitReader(r.Body, maxArtifactWrite+1))
	if err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	if len(data) > maxArtifactWrite {
		fail(w, http.StatusRequestEntityTooLarge, "a file written into the artifacts is at most 16 MiB")
		return
	}
	artifact, err := h.opts.Runner.WriteArtifact(r.Context(), param(r, "*"), data, access.UserFrom(r.Context()).Name)
	switch {
	case errors.Is(err, artifactstore.ErrPath):
		fail(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, artifactstore.ErrInTheWay):
		fail(w, http.StatusConflict, err.Error())
	case err != nil:
		failErr(w, err)
	default:
		writeJSON(w, http.StatusCreated, artifact)
	}
}

// deleteArtifactFile takes a file or a directory out of the artifact
// tree; the history keeps its versions. Takes artifact/edit.
func (h *Handlers) deleteArtifactFile(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "artifact/edit") {
		return
	}
	if h.opts.ArtifactStore == nil {
		fail(w, http.StatusNotFound, artifactstore.ErrNotFound.Error())
		return
	}
	_, err := h.opts.ArtifactStore.Remove(r.Context(), folderParam(r), access.UserFrom(r.Context()).Name)
	if errors.Is(err, artifactstore.ErrNotFound) {
		fail(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		failErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// hasVersion reports whether the versions hold the commit, named in
// full or by its first seven characters or more.
func hasVersion(versions []ArtifactVersion, commit string) bool {
	for _, version := range versions {
		if version.Commit == commit || len(commit) >= 7 && strings.HasPrefix(version.Commit, commit) {
			return true
		}
	}
	return false
}
