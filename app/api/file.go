package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/titpetric/task-ui/internal/access"
	"github.com/titpetric/task-ui/internal/folder"
	"github.com/titpetric/task-ui/internal/highlight"
	"github.com/titpetric/task-ui/internal/markdown"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/storage"
)

// viewFile returns a file of a folder's checkout, or its bytes with ?raw, for whoever sees the folder.
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
	if _, raw := r.URL.Query()["raw"]; raw {
		w.Header().Set("Content-Type", runner.ContentType(rel))
		http.ServeContent(w, r, path.Base(rel), info.ModTime(), file)
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
	case strings.HasPrefix(view.ContentType, "audio/"):
		view.Kind, view.Truncated = "audio", false
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

// maxFileWrite bounds a file written into a checkout from the editor.
const maxFileWrite = 16 << 20

// maxCommitMessage bounds the message an edit is committed with.
const maxCommitMessage = 200

// maxFileRequest bounds the body the editor sends: the file as JSON, whose
// characters can escape to more than the bytes they stand for. A control
// character escapes to six bytes, \u00XX, and the message escapes with it, so
// a file at maxFileWrite fits the bound; the rest is the JSON around them.
const maxFileRequest = maxFileWrite*6 + maxCommitMessage*6 + 4096

// fileWrite is what the editor's Save sends: the whole file and the message
// its change is committed with.
type fileWrite struct {
	Content *string `json:"content"`
	Message string  `json:"message"`
}

// fileSaved is what the editor's Save answers.
type fileSaved struct {
	Folder    string `json:"folder"`
	Path      string `json:"path"`
	Commit    string `json:"commit,omitempty"`
	Committed bool   `json:"committed"`
}

// writeFile writes a text file of a checkout and commits it with the
// reader's message, as the reader, answering the commit the file is in.
func (h *Handlers) writeFile(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "project/edit") {
		return
	}
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
	info, err := os.Stat(full)
	if err != nil || !info.Mode().IsRegular() {
		failErr(w, storage.ErrNotFound)
		return
	}
	defer r.Body.Close()
	body, err := io.ReadAll(io.LimitReader(r.Body, maxFileRequest+1))
	if err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	if len(body) > maxFileRequest {
		fail(w, http.StatusRequestEntityTooLarge, "a file written from the editor is at most 16 MiB")
		return
	}
	var req fileWrite
	if err := json.Unmarshal(body, &req); err != nil || req.Content == nil {
		fail(w, http.StatusBadRequest, "the file's content is needed")
		return
	}
	data := []byte(*req.Content)
	if len(data) > maxFileWrite {
		fail(w, http.StatusRequestEntityTooLarge, "a file written from the editor is at most 16 MiB")
		return
	}
	if !utf8.Valid(data) || bytes.IndexByte(data, 0) >= 0 {
		fail(w, http.StatusBadRequest, "a binary file is not editable")
		return
	}
	message := strings.TrimSpace(req.Message)
	if message == "" {
		fail(w, http.StatusBadRequest, "a commit message is needed to save a file")
		return
	}
	if utf8.RuneCountInString(message) > maxCommitMessage {
		fail(w, http.StatusBadRequest, "a commit message is at most 200 characters")
		return
	}
	if err := os.WriteFile(full, data, info.Mode().Perm()); err != nil {
		failErr(w, err)
		return
	}
	saved := fileSaved{Folder: f.ID, Path: rel}
	if h.opts.Git != nil {
		if checkout := h.opts.Git.Checkout(f.Dir, f.Depth()); checkout != "" {
			target, err := filepath.Rel(checkout, full)
			if err == nil {
				author := access.UserFrom(r.Context()).Name
				commit, made, err := h.opts.Git.CommitPath(r.Context(), checkout, filepath.ToSlash(target), message, author)
				if err != nil {
					writeJSON(w, http.StatusInternalServerError, map[string]any{
						"written": true,
						"folder":  f.ID,
						"path":    rel,
						"error":   "the file was written but the commit failed: " + err.Error(),
					})
					return
				}
				saved.Commit, saved.Committed = commit, made
			}
		}
	}
	writeJSON(w, http.StatusOK, saved)
}

// pushRequest names the folder whose checkout the Push action sends.
type pushRequest struct {
	Folder string `json:"folder"`
}

// pushCheckout pushes a folder's checkout, with the key the instance holds.
func (h *Handlers) pushCheckout(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "project/edit") {
		return
	}
	var req pushRequest
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&req); err != nil {
		fail(w, http.StatusBadRequest, "a folder is needed")
		return
	}
	f, _ := folder.Resolve(h.opts.Catalogue, req.Folder)
	if f == nil || !access.UserFrom(r.Context()).CanSee(f.ID) {
		failErr(w, storage.ErrNotFound)
		return
	}
	if h.opts.Git == nil {
		fail(w, http.StatusBadRequest, "this instance holds no git")
		return
	}
	checkout := h.opts.Git.Checkout(f.Dir, f.Depth())
	if checkout == "" {
		fail(w, http.StatusBadRequest, "the folder is not a git checkout")
		return
	}
	var out bytes.Buffer
	if err := h.opts.Git.Push(r.Context(), checkout, &out); err != nil {
		said := strings.TrimSpace(out.String())
		if said == "" {
			said = err.Error()
		}
		fail(w, http.StatusBadGateway, said)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"pushed": true, "output": strings.TrimSpace(out.String())})
}
