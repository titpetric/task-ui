package frontend

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/diskusage"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/storage"
)

// artifacts shows the artifact tree: the latest version of every file the
// runs left, a directory at a time (?path=), the space the files and their
// history take, and the record of every version, newest first.
func (h *Handlers) artifacts(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "artifact/view") {
		return
	}
	user := access.UserFrom(r.Context())
	q := r.URL.Query()
	p := h.page(r, "Artifacts")
	dir := strings.Trim(path.Clean("/"+q.Get("path")), "/")
	p.Crumbs = artifactCrumbs(dir)
	p.Form = map[string]string{"folder": q.Get("folder"), "job": q.Get("job")}
	if dir != "" {
		p.Browse = dir + "/"
	}
	p.Archive = archiveHref(dir)

	if store := h.opts.ArtifactStore; store != nil {
		visible := h.visibleArtifacts(r)
		for _, item := range store.List(dir) {
			full := path.Join(dir, item.Name)
			if !user.CanAdmin() && !visible.covers(full, item.Dir) {
				continue
			}
			item.Path = full
			if item.Dir {
				item.Href = "/artifacts?path=" + url.QueryEscape(full)
				item.Archive = archiveHref(full)
				if dir, err := store.Path(full); err == nil {
					item.Size = diskusage.Measure(dir).Files
				}
			} else {
				item.Href = "/artifacts/files/" + escapePath(full)
			}
			p.Tree = append(p.Tree, item)
		}
		if h.opts.Disk != nil {
			usage := h.opts.Disk.Measure(store.Dir)
			p.Space = &usage
		}
	}

	artifacts, _ := h.opts.Artifacts.List(r.Context(), storage.ArtifactFilter{Folder: q.Get("folder"), Job: q.Get("job"), Limit: 200})
	p.Artifacts = []model.Artifact{}
	for _, artifact := range artifacts {
		if user.CanSee(artifact.Folder) {
			p.Artifacts = append(p.Artifacts, artifact)
		}
	}
	h.render(w, http.StatusOK, "artifacts", p)
}

// artifactFile serves the latest version of a file in the artifact tree,
// or the one ?version= names, inline: an image or a page opens in the
// browser, ?download saves it. A directory goes to its listing.
func (h *Handlers) artifactFile(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "artifact/view") {
		return
	}
	store := h.opts.ArtifactStore
	rel := strings.Trim(param(r, "*"), "/")
	if store == nil || rel == "" {
		h.missing(w, r, "No such artifact.")
		return
	}
	user := access.UserFrom(r.Context())
	if version := r.URL.Query().Get("version"); version != "" {
		h.artifactVersion(w, r, rel, version)
		return
	}
	if !user.CanAdmin() && !h.visibleArtifacts(r).covers(rel, false) {
		h.missing(w, r, "No such artifact.")
		return
	}
	file, err := store.Path(rel)
	if err != nil {
		h.missing(w, r, "No such artifact.")
		return
	}
	info, err := os.Stat(file)
	if err != nil {
		h.missing(w, r, "No such artifact.")
		return
	}
	if info.IsDir() {
		http.Redirect(w, r, "/artifacts?path="+url.QueryEscape(rel), http.StatusSeeOther)
		return
	}
	f, err := os.Open(file)
	if err != nil {
		h.missing(w, r, "No such artifact.")
		return
	}
	defer f.Close()
	serveArtifact(w, r, rel, info.ModTime(), f)
}

// artifactVersion serves a version of a file of the artifact tree, from
// its history, when a run of a folder the user sees wrote it.
func (h *Handlers) artifactVersion(w http.ResponseWriter, r *http.Request, rel, version string) {
	user := access.UserFrom(r.Context())
	records, _ := h.opts.Artifacts.List(r.Context(), storage.ArtifactFilter{Path: rel, Limit: 200})
	var when time.Time
	found := false
	for _, record := range records {
		if record.CommitID == version && (user.CanAdmin() || user.CanSee(record.Folder)) {
			found = true
			if record.CreatedAt != nil {
				when = *record.CreatedAt
			}
		}
	}
	if !found && !user.CanAdmin() {
		h.missing(w, r, "No such artifact.")
		return
	}
	data, err := h.opts.ArtifactStore.Show(r.Context(), version, rel)
	if err != nil {
		h.missing(w, r, "No such artifact.")
		return
	}
	serveArtifact(w, r, rel, when, bytes.NewReader(data))
}

// serveArtifact writes a file of the artifact tree, inline unless the
// query carries ?download.
func serveArtifact(w http.ResponseWriter, r *http.Request, rel string, modified time.Time, content io.ReadSeeker) {
	serveFile(w, r, rel, runner.ContentType(rel), modified, content)
}

// serveFile writes a file a run or a checkout holds as its type says,
// inline unless the query carries ?download.
func serveFile(w http.ResponseWriter, r *http.Request, rel, contentType string, modified time.Time, content io.ReadSeeker) {
	disposition := "inline"
	if _, ok := r.URL.Query()["download"]; ok {
		disposition = "attachment"
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", disposition+`; filename="`+strings.ReplaceAll(path.Base(rel), `"`, "")+`"`)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	// A run or a commit wrote the file: an svg or an html page opens
	// without its scripts, in a sandbox of its own.
	w.Header().Set("Content-Security-Policy", "default-src 'none'; img-src 'self' data:; style-src 'unsafe-inline'; sandbox")
	w.Header().Set("Cache-Control", "no-cache")
	http.ServeContent(w, r, "", modified, content)
}

// visibleArtifacts returns the paths of the tree the user may see: the
// files whose latest version a folder they see wrote.
func (h *Handlers) visibleArtifacts(r *http.Request) artifactPaths {
	user := access.UserFrom(r.Context())
	latest, _ := h.opts.Artifacts.Latest(r.Context())
	paths := artifactPaths{}
	for _, artifact := range latest {
		if user.CanSee(artifact.Folder) {
			paths[artifact.Path] = true
		}
	}
	return paths
}

// artifactPaths is a set of file paths in the artifact tree.
type artifactPaths map[string]bool

// covers reports whether a file is in the set, or a directory holds one.
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

// artifactCrumbs are the crumbs of a directory of the artifact tree.
func artifactCrumbs(dir string) []view.Crumb {
	if dir == "" {
		return view.Crumbs("Artifacts")
	}
	pairs := []string{"Artifacts", "/artifacts"}
	parts := strings.Split(dir, "/")
	for i, part := range parts[:len(parts)-1] {
		pairs = append(pairs, part, "/artifacts?path="+url.QueryEscape(strings.Join(parts[:i+1], "/")))
	}
	pairs = append(pairs, parts[len(parts)-1])
	return view.Crumbs(pairs...)
}

// escapePath escapes each element of a slash path for a URL.
func escapePath(p string) string {
	parts := strings.Split(p, "/")
	for i, part := range parts {
		parts[i] = url.PathEscape(part)
	}
	return strings.Join(parts, "/")
}

// artifactArchive serves a directory of the artifact tree as a gzipped
// tar of the latest files the user may see: /artifacts/<dir>.tgz, or
// /artifacts.tgz for the whole tree.
func (h *Handlers) artifactArchive(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "artifact/view") {
		return
	}
	store := h.opts.ArtifactStore
	wanted := param(r, "*")
	if r.URL.Path != "/artifacts.tgz" && !strings.HasSuffix(wanted, ".tgz") {
		h.missing(w, r, "No such artifact.")
		return
	}
	rel := strings.Trim(path.Clean("/"+strings.TrimSuffix(wanted, ".tgz")), "/")
	if store == nil {
		h.missing(w, r, "No such artifact.")
		return
	}
	user := access.UserFrom(r.Context())
	visible := h.visibleArtifacts(r)
	if rel != "" && !user.CanAdmin() && !visible.covers(rel, true) {
		h.missing(w, r, "No such artifact.")
		return
	}
	dir, err := store.Path(rel)
	if err != nil {
		h.missing(w, r, "No such artifact.")
		return
	}
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		h.missing(w, r, "No such artifact.")
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

// archiveHref is where a directory of the artifact tree downloads as a
// .tgz: /artifacts/<dir>.tgz, or /artifacts.tgz for the whole tree.
func archiveHref(dir string) string {
	dir = strings.Trim(dir, "/")
	if dir == "" {
		return "/artifacts.tgz"
	}
	return "/artifacts/" + escapePath(dir) + ".tgz"
}
