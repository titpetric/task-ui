package frontend

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
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/runner"
)

// checkoutTree links a listing of a directory inside a folder, rel to
// it: a directory to its page, a file to /files/ for the preview window,
// each with its path. A sealed entry stays plain, as does a symlink
// that leads out of the checkout.
func (h *Handlers) checkoutTree(p *view.Page, f *folder.Folder, rel string, items []folder.Item) []folder.Item {
	for i := range items {
		item := &items[i]
		if item.Sealed {
			continue
		}
		item.Path = path.Join(rel, item.Name)
		if _, ok := f.Path(item.Path); !ok {
			// A symlink out of the checkout, or a dangling one.
			item.Path, item.Exec = "", false
			continue
		}
		if item.Dir {
			item.Href = p.Href(f.ID) + "/" + escapePath(item.Path)
		} else {
			item.Href = fileHref(f.ID, item.Path)
		}
	}
	return items
}

// fileHref is where a file of a folder's checkout opens on its own.
func fileHref(folderID, rel string) string {
	return "/files/" + escapePath(folderID+"/"+rel)
}

// checkoutFile serves a file of a folder's checkout, /files/<folder>/<path>,
// for whoever sees the folder: inline as its type says, text without an
// extension as text, ?download to save it. A path with .., a symlink out
// of the checkout, a sealed entry or a path through one is not found.
func (h *Handlers) checkoutFile(w http.ResponseWriter, r *http.Request) {
	f, rel := folder.Resolve(h.opts.Catalogue, param(r, "*"))
	if f == nil || rel == "" || !access.UserFrom(r.Context()).CanSee(f.ID) {
		h.missing(w, r, "No such file.")
		return
	}
	full, ok := f.Path(rel)
	if !ok {
		h.missing(w, r, "No such file.")
		return
	}
	file, err := os.Open(full)
	if err != nil {
		h.missing(w, r, "No such file.")
		return
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		h.missing(w, r, "No such file.")
		return
	}
	serveFile(w, r, rel, sniffType(rel, file), info.ModTime(), file)
}

// sniffType is a file's type by its extension, or plain text for a file
// the extension does not name whose start reads as text: a script, a
// Makefile, a LICENSE. sniffType rewinds the file after.
func sniffType(rel string, file io.ReadSeeker) string {
	kind := runner.ContentType(rel)
	if kind != "application/octet-stream" {
		return kind
	}
	head := make([]byte, 512)
	n, _ := io.ReadFull(file, head)
	_, _ = file.Seek(0, io.SeekStart)
	head = head[:n]
	for i := 0; i < utf8.UTFMax && n == len(head) && len(head) > 0 && !utf8.Valid(head); i++ {
		head = head[:len(head)-1]
	}
	if utf8.Valid(head) && bytes.IndexByte(head, 0) < 0 && !strings.HasPrefix(string(head), "\x7fELF") {
		return "text/plain; charset=utf-8"
	}
	return kind
}
