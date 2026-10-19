package api

import (
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/highlight"
	"github.com/titpetric/task-ui/storage"
)

// SourceView is a file a folder's pipeline reads, as it is and coloured.
type SourceView struct {
	File    string        `json:"file"`
	Content string        `json:"content"`
	HTML    template.HTML `json:"html"`
}

// maxSource is the largest file the source view serves.
const maxSource = 1 << 20

// getSource returns a file of a folder's sources, ?file=atkins.yml, or a
// file they include, for whoever sees the folder. Only the files the
// folder lists are served, from inside the root.
func (h *Handlers) getSource(w http.ResponseWriter, r *http.Request) {
	f, rest := folder.Resolve(h.opts.Catalogue, param(r, "*"))
	if f == nil || rest != "" || !access.UserFrom(r.Context()).CanSee(f.ID) {
		failErr(w, storage.ErrNotFound)
		return
	}
	name := r.URL.Query().Get("file")
	path, ok := h.sourcePath(f, name)
	if !ok {
		failErr(w, storage.ErrNotFound)
		return
	}
	info, err := os.Stat(path)
	if err != nil || info.IsDir() || info.Size() > maxSource {
		failErr(w, storage.ErrNotFound)
		return
	}
	data, err := os.ReadFile(path)
	if err != nil {
		failErr(w, storage.ErrNotFound)
		return
	}
	writeJSON(w, http.StatusOK, SourceView{File: name, Content: string(data), HTML: highlight.HTML(name, string(data))})
}

// sourcePath returns where a file the folder lists is: one of its
// sources' files, or a file they include, inside the root.
func (h *Handlers) sourcePath(f *folder.Folder, name string) (string, bool) {
	if name == "" {
		return "", false
	}
	listed := false
	for _, source := range f.Sources {
		listed = listed || source.File == name
	}
	for _, include := range f.Includes {
		listed = listed || include == name
	}
	if !listed || filepath.IsAbs(name) {
		return "", false
	}
	path := filepath.Join(f.Dir, filepath.FromSlash(name))
	root := h.opts.Root
	if root == "" {
		root = f.Dir
	}
	rootAbs, err1 := filepath.Abs(root)
	pathAbs, err2 := filepath.Abs(path)
	if err1 != nil || err2 != nil || !strings.HasPrefix(pathAbs, rootAbs+string(filepath.Separator)) {
		return "", false
	}
	return pathAbs, true
}
