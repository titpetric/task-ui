package api

import (
	"errors"
	"net/http"
	"os"
	"path"
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
	for _, record := range records {
		if record.CommitID != "" && user.CanSee(record.Folder) {
			view.Versions = append(view.Versions, ArtifactVersion{Commit: record.CommitID, CreatedAt: record.CreatedAt, Folder: record.Folder, Job: record.Job, RunID: record.RunID, ByteSize: record.ByteSize})
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

// deleteArtifactFile takes a file or a directory out of the artifact
// tree; the history keeps its versions. Admins only.
func (h *Handlers) deleteArtifactFile(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
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

// hasVersion reports whether the versions hold the commit.
func hasVersion(versions []ArtifactVersion, commit string) bool {
	for _, version := range versions {
		if version.Commit == commit {
			return true
		}
	}
	return false
}
