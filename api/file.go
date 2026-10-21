package api

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"unicode/utf8"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/highlight"
	"github.com/titpetric/task-ui/markdown"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/storage"
)

// viewFile returns a file of a folder's checkout for the preview window,
// /api/files/<folder>/<path>, for whoever sees the folder. Only a file
// inside the checkout opens: a path with .., a symlink out of it, a
// sealed entry or a path through one answer 404, as a directory does.
func (h *Handlers) viewFile(w http.ResponseWriter, r *http.Request) {
	f, rel := folder.Resolve(h.opts.Catalogue, param(r, "*"))
	if f == nil || rel == "" || !access.UserFrom(r.Context()).CanSee(f.ID) {
		failErr(w, storage.ErrNotFound)
		return
	}
	full, ok := f.Path(rel)
	if !ok {
		failErr(w, storage.ErrNotFound)
		return
	}
	file, err := os.Open(full)
	if err != nil {
		failErr(w, storage.ErrNotFound)
		return
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		failErr(w, storage.ErrNotFound)
		return
	}
	data, err := io.ReadAll(io.LimitReader(file, maxPreview))
	if err != nil {
		failErr(w, err)
		return
	}
	view := FileView{Folder: f.ID, Path: rel, ContentType: runner.ContentType(rel), ByteSize: info.Size()}
	view.Truncated = info.Size() > int64(len(data))
	if view.Truncated {
		// The cut may fall inside a character; the text ends before it.
		for i := 0; i < utf8.UTFMax && len(data) > 0 && !utf8.Valid(data); i++ {
			data = data[:len(data)-1]
		}
	}
	ext := strings.ToLower(path.Ext(rel))
	switch {
	case imageTypes[ext]:
		view.Kind, view.Truncated = "image", false
	case !utf8.Valid(data) || bytes.IndexByte(data, 0) >= 0:
		view.Kind, view.Truncated = "file", false
	case (ext == ".md" || ext == ".markdown") && !view.Truncated:
		view.Kind = "markdown"
		view.HTML = markdown.HTML(data, "/files/"+f.ID+"/", path.Dir(rel))
	default:
		view.Kind = "text"
		view.Text = string(data)
		view.HTML = string(highlight.HTML(rel, view.Text))
	}
	writeJSON(w, http.StatusOK, view)
}
