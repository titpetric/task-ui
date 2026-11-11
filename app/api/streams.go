package api

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	chi "github.com/go-chi/chi/v5"

	"github.com/titpetric/task-ui/internal/access"
	"github.com/titpetric/task-ui/internal/folder"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/storage"
)

// Streams are the byte streams the front end links to at the root: a file of a checkout, a file of the artifact tree and a directory of it as a tar.
func (h *Handlers) Streams() http.Handler {
	r := chi.NewRouter()
	r.Get("/files/*", h.streamCheckoutFile)
	r.Get("/artifacts/files/*", h.streamArtifactFile)
	r.Get("/artifacts.tgz", h.streamArtifactArchive)
	r.Get("/artifacts/*", h.streamArtifactArchive)
	return r
}

// streamCheckoutFile writes a file of a folder's checkout for whoever sees the folder; a path out of it is not found.
func (h *Handlers) streamCheckoutFile(w http.ResponseWriter, r *http.Request) {
	f, rel := folder.Resolve(h.opts.Catalogue, param(r, "*"))
	if f == nil || rel == "" || !access.UserFrom(r.Context()).CanSee(f.ID) {
		http.NotFound(w, r)
		return
	}
	full, ok := f.Path(rel)
	if !ok {
		http.NotFound(w, r)
		return
	}
	file, err := os.Open(full)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		http.NotFound(w, r)
		return
	}
	streamFile(w, r, rel, sniffType(rel, file), info.ModTime(), file)
}

// streamArtifactFile writes the latest version of a file of the artifact tree, or the one ?version= names.
func (h *Handlers) streamArtifactFile(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "artifact/view") {
		return
	}
	store := h.opts.ArtifactStore
	rel := strings.Trim(param(r, "*"), "/")
	if store == nil || rel == "" {
		http.NotFound(w, r)
		return
	}
	if version := r.URL.Query().Get("version"); version != "" {
		h.streamArtifactVersion(w, r, rel, version)
		return
	}
	user := access.UserFrom(r.Context())
	if !user.CanAdmin() && !h.visibleArtifacts(r).covers(rel, false) {
		http.NotFound(w, r)
		return
	}
	file, err := store.Path(rel)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	info, err := os.Stat(file)
	if err != nil || info.IsDir() {
		http.NotFound(w, r)
		return
	}
	open, err := os.Open(file)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer open.Close()
	streamFile(w, r, rel, runner.ContentType(rel), info.ModTime(), open)
}

// streamArtifactVersion writes a version from the history, when a run of a folder the user sees wrote it.
func (h *Handlers) streamArtifactVersion(w http.ResponseWriter, r *http.Request, rel, version string) {
	user := access.UserFrom(r.Context())
	records, _ := h.opts.Artifacts.List(r.Context(), storage.ArtifactFilter{Path: rel, Limit: 200})
	var when time.Time
	found := false
	for _, record := range records {
		if record.CommitID == version && (user.CanAdmin() || user.SeesArtifact(&record)) {
			found = true
			if record.CreatedAt != nil {
				when = *record.CreatedAt
			}
		}
	}
	if !found && !user.CanAdmin() {
		http.NotFound(w, r)
		return
	}
	data, err := h.opts.ArtifactStore.Show(r.Context(), version, rel)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	streamFile(w, r, rel, runner.ContentType(rel), when, bytes.NewReader(data))
}

// streamArtifactArchive writes a directory of the artifact tree as a tar of the latest files the user may see.
func (h *Handlers) streamArtifactArchive(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "artifact/view") {
		return
	}
	store := h.opts.ArtifactStore
	wanted := param(r, "*")
	if store == nil || (r.URL.Path != "/artifacts.tgz" && !strings.HasSuffix(wanted, ".tgz")) {
		http.NotFound(w, r)
		return
	}
	rel := strings.Trim(path.Clean("/"+strings.TrimSuffix(wanted, ".tgz")), "/")
	user := access.UserFrom(r.Context())
	visible := h.visibleArtifacts(r)
	if rel != "" && !user.CanAdmin() && !visible.covers(rel, true) {
		http.NotFound(w, r)
		return
	}
	dir, err := store.Path(rel)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		http.NotFound(w, r)
		return
	}
	name := "artifacts"
	if rel != "" {
		name = path.Base(rel)
	}

	w.Header().Set("Content-Type", "application/gzip")
	w.Header().Set("Content-Disposition", `attachment; filename="`+strings.ReplaceAll(name, `"`, "")+`.tgz"`)
	zw := gzip.NewWriter(w)
	tw := tar.NewWriter(zw)
	_ = filepath.WalkDir(dir, func(p string, entry fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if entry.Name() == ".git" && entry.IsDir() {
			return filepath.SkipDir
		}
		if entry.IsDir() || !entry.Type().IsRegular() || strings.HasPrefix(entry.Name(), ".") {
			return nil
		}
		inner, _ := filepath.Rel(dir, p)
		full := path.Join(rel, filepath.ToSlash(inner))
		if !user.CanAdmin() && !visible[full] {
			return nil
		}
		return addToTar(tw, p, path.Join(name, filepath.ToSlash(inner)))
	})
	_ = tw.Close()
	_ = zw.Close()
}

// streamFile writes a file as its type says, inline unless ?download.
func streamFile(w http.ResponseWriter, r *http.Request, rel, contentType string, modified time.Time, content io.ReadSeeker) {
	disposition := "inline"
	if _, ok := r.URL.Query()["download"]; ok {
		disposition = "attachment"
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", disposition+`; filename="`+strings.ReplaceAll(path.Base(rel), `"`, "")+`"`)
	runner.SandboxFile(w)
	http.ServeContent(w, r, "", modified, content)
}

// sniffType is a file's type by its extension, text where the extension names none and the start reads as text.
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

// visibleArtifacts are the paths whose latest version a folder the user sees wrote.
func (h *Handlers) visibleArtifacts(r *http.Request) artifactPaths {
	user := access.UserFrom(r.Context())
	latest, _ := h.opts.Artifacts.Latest(r.Context())
	paths := artifactPaths{}
	for _, artifact := range latest {
		if user.SeesArtifact(&artifact) {
			paths[artifact.Path] = true
		}
	}
	return paths
}

// artifactPaths is a set of file paths in the artifact tree.
type artifactPaths map[string]bool

// covers reports a file in the set, or a directory holding one.
func (a artifactPaths) covers(rel string, dir bool) bool {
	if !dir {
		return a[rel]
	}
	for p := range a {
		if strings.HasPrefix(p, rel+"/") {
			return true
		}
	}
	return false
}

// addToTar writes a file into a tar under a name.
func addToTar(tw *tar.Writer, file, name string) error {
	f, err := os.Open(file)
	if err != nil {
		return nil
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil
	}
	if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0o644, Size: info.Size(), ModTime: info.ModTime(), Typeflag: tar.TypeReg}); err != nil {
		return err
	}
	_, err = io.Copy(tw, f)
	return err
}
