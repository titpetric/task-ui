package frontend

import (
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/storage"
)

// artifacts shows the artifact tree: the latest version of every file the
// runs left, a directory at a time (?path=), the space the files and their
// history take, and the record of every version, newest first.
func (h *Handlers) artifacts(w http.ResponseWriter, r *http.Request) {
	user := access.UserFrom(r.Context())
	q := r.URL.Query()
	p := h.page(r, "Artifacts")
	dir := strings.Trim(path.Clean("/"+q.Get("path")), "/")
	p.Crumbs = artifactCrumbs(dir)
	p.Form = map[string]string{"folder": q.Get("folder"), "job": q.Get("job")}
	if dir != "" {
		p.Browse = dir + "/"
	}

	if store := h.opts.ArtifactStore; store != nil {
		visible := h.visibleArtifacts(r)
		for _, item := range store.List(dir) {
			full := path.Join(dir, item.Name)
			if !user.CanAdmin() && !visible.covers(full, item.Dir) {
				continue
			}
			if item.Dir {
				item.Href = "/artifacts?path=" + url.QueryEscape(full)
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
// inline: an image or a page opens in the browser, ?download saves it. A
// directory goes to its listing.
func (h *Handlers) artifactFile(w http.ResponseWriter, r *http.Request) {
	store := h.opts.ArtifactStore
	rel := strings.Trim(param(r, "*"), "/")
	if store == nil || rel == "" {
		h.missing(w, r, "No such artifact.")
		return
	}
	user := access.UserFrom(r.Context())
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
	disposition := "inline"
	if _, ok := r.URL.Query()["download"]; ok {
		disposition = "attachment"
	}
	w.Header().Set("Content-Type", runner.ContentType(rel))
	w.Header().Set("Content-Disposition", disposition+`; filename="`+strings.ReplaceAll(path.Base(rel), `"`, "")+`"`)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	// A run wrote the file: an svg or an html page opens without its
	// scripts, in a sandbox of its own.
	w.Header().Set("Content-Security-Policy", "default-src 'none'; img-src 'self' data:; style-src 'unsafe-inline'; sandbox")
	w.Header().Set("Cache-Control", "no-cache")
	http.ServeContent(w, r, "", info.ModTime(), f)
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

// linkTree links each directory of a listing inside a folder to its page.
func linkTree(items []folder.Item, base string) []folder.Item {
	for i := range items {
		if items[i].Dir {
			items[i].Href = base + escapePath(items[i].Name)
		}
	}
	return items
}
