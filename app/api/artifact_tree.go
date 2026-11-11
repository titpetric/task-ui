package api

import (
	"net/http"
	"path"
	"strings"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/artifactstore"
	"github.com/titpetric/task-ui/diskusage"
	"github.com/titpetric/task-ui/model"
)

// ArtifactTreeView is a directory of the artifact tree, the latest
// versions: its directories, then its files, those the user may see.
type ArtifactTreeView struct {
	Path    string          `json:"path"`
	Entries []ArtifactEntry `json:"entries"`
}

// artifactTree lists a directory of the artifact tree, the root without
// a path, as the Artifacts page does: each file with its size, time and
// the run or the user that wrote its latest version. A user sees the
// files whose latest version a folder they see wrote, or a user wrote.
func (h *Handlers) artifactTree(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "artifact/view") {
		return
	}
	store := h.opts.ArtifactStore
	if store == nil {
		fail(w, http.StatusNotFound, artifactstore.ErrNotFound.Error())
		return
	}
	user := access.UserFrom(r.Context())
	dir := strings.Trim(path.Clean("/"+folderParam(r)), "/")
	latest, err := h.opts.Artifacts.Latest(r.Context())
	if err != nil {
		failErr(w, err)
		return
	}
	writers := map[string]model.Artifact{}
	for _, artifact := range latest {
		if user.CanAdmin() || user.SeesArtifact(&artifact) {
			writers[artifact.Path] = artifact
		}
	}
	view := ArtifactTreeView{Path: dir, Entries: []ArtifactEntry{}}
	for _, item := range store.List(dir) {
		full := path.Join(dir, item.Name)
		entry := ArtifactEntry{Name: item.Name, Path: full, Dir: item.Dir, ByteSize: item.Size, Modified: item.ModTime}
		if item.Dir {
			if !user.CanAdmin() && !holds(writers, full) {
				continue
			}
			if p, err := store.Path(full); err == nil {
				entry.ByteSize = diskusage.Measure(p).Files
			}
		} else {
			writer, ok := writers[full]
			if !ok && !user.CanAdmin() {
				continue
			}
			entry.RunID, entry.Folder, entry.Job, entry.UserName = writer.RunID, writer.Folder, writer.Job, writer.UserName
		}
		view.Entries = append(view.Entries, entry)
	}
	if dir != "" && len(view.Entries) == 0 && !holds(writers, dir) && !user.CanAdmin() {
		fail(w, http.StatusNotFound, artifactstore.ErrNotFound.Error())
		return
	}
	writeJSON(w, http.StatusOK, view)
}

// holds reports whether a directory of the tree holds a file of the set.
func holds(files map[string]model.Artifact, dir string) bool {
	for p := range files {
		if strings.HasPrefix(p, dir+"/") {
			return true
		}
	}
	return false
}
