// Package assets serves a tree's static files: one directory per segment,
// with the content hash their URLs carry and the entry stylesheet's imports
// versioned.
//
// It is the one owner of the static files. A front end holds a Tree and
// serves it; nothing else reads the directories.
package assets

import (
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"log/slog"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Prefix is where a Tree is served.
const Prefix = "/assets/"

// Theme is the segment the theme's own files answer under.
const Theme = "theme"

// Stylesheet is the entry stylesheet of the theme segment.
const Stylesheet = "task-ui.css"

// MaxAge is how long a browser keeps a file. The URLs carry ?v=, so an hour
// never outlives a deploy.
const MaxAge = "public, max-age=3600"

// importURL matches the url of an @import in the stylesheet.
var importURL = regexp.MustCompile(`@import url\("(css/[^"?]+\.css)"\)`)

// Tree is a set of named static directories. New reads each one once, for the
// content hash and the versioned stylesheet, and a request reads no more than
// the file it asked for.
type Tree struct {
	// Segments is the filesystem each URL segment answers from.
	Segments map[string]fs.FS

	revision string
	version  string
	css      []byte
}

// New returns the Tree of a segment name to filesystem map, under the revision
// the server made for this run.
func New(segments map[string]fs.FS, revision string) *Tree {
	t := &Tree{Segments: segments, revision: revision}
	if t.Segments == nil {
		t.Segments = map[string]fs.FS{}
	}
	t.version = t.hash()
	t.css = t.stylesheet()
	return t
}

// Dirs is the segment map of an application tree: the theme's assets, the
// editor's bundle and every module's, leaving out a directory that is not
// there.
func Dirs(root string) map[string]fs.FS {
	segments := map[string]fs.FS{}
	add := func(name, dir string) {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			segments[name] = os.DirFS(dir)
		}
	}
	add(Theme, filepath.Join(root, "theme", "assets"))
	add("mdedit", filepath.Join(root, "ext", "titpetric", "mdedit", "dist"))
	entries, err := os.ReadDir(filepath.Join(root, "modules"))
	if err != nil {
		return segments
	}
	for _, entry := range entries {
		if entry.IsDir() {
			add(entry.Name(), filepath.Join(root, "modules", entry.Name(), "assets"))
		}
	}
	return segments
}

// Version is the content hash the URLs carry as ?v=. The build version cannot
// do it: a `go install` build is "dev" every time.
func (t *Tree) Version() string {
	if t == nil {
		return ""
	}
	return t.version
}

// Revision is the identifier of this tree, made once when the server starts.
func (t *Tree) Revision() string {
	if t == nil {
		return ""
	}
	return t.revision
}

// Href is where the tree is served, the revision segment included.
func (t *Tree) Href() string {
	href := strings.TrimSuffix(Prefix, "/")
	if t == nil || t.revision == "" {
		return href
	}
	return href + "/r_" + t.revision
}

// CSS is the entry stylesheet as served, each import of css/ carrying Version.
func (t *Tree) CSS() []byte {
	if t == nil {
		return nil
	}
	return t.css
}

// Open reads one file of a segment, as the segment's own filesystem.
func (t *Tree) Open(segment, name string) ([]byte, error) {
	root, ok := t.Segments[segment]
	if !ok {
		return nil, fs.ErrNotExist
	}
	if !fs.ValidPath(name) {
		return nil, fs.ErrInvalid
	}
	return fs.ReadFile(root, name)
}

// ServeHTTP answers <Prefix><segment>/<path> from that segment's directory.
func (t *Tree) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := t.serveHTTP(w, r); err != nil {
		http.NotFound(w, r)
	}
}

// serveHTTP writes the file and reports what was not there.
func (t *Tree) serveHTTP(w http.ResponseWriter, r *http.Request) error {
	rest := strings.TrimPrefix(path.Clean(r.URL.Path), Prefix)
	if first, after, ok := strings.Cut(rest, "/"); ok && first == "r_"+t.revision {
		rest = after
	}
	segment, rest, ok := strings.Cut(rest, "/")
	if !ok || rest == "" {
		return fs.ErrNotExist
	}
	if segment == "modules" {
		segment, rest, ok = strings.Cut(rest, "/")
		if !ok || rest == "" {
			return fs.ErrNotExist
		}
	}
	if _, known := t.Segments[segment]; !known {
		return fs.ErrNotExist
	}
	if segment == Theme && rest == Stylesheet && t.css != nil {
		w.Header().Set("Cache-Control", MaxAge)
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		write(w, r, t.css)
		return nil
	}
	data, err := t.Open(segment, rest)
	if err != nil {
		return err
	}
	w.Header().Set("Cache-Control", MaxAge)
	if kind := mime.TypeByExtension(filepath.Ext(rest)); kind != "" {
		w.Header().Set("Content-Type", kind)
	}
	write(w, r, data)
	return nil
}

// hash is the content hash over all files in all segments, in a fixed order.
func (t *Tree) hash() string {
	h := sha256.New()
	names := make([]string, 0, len(t.Segments))
	for name := range t.Segments {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		root := t.Segments[name]
		_ = fs.WalkDir(root, ".", func(at string, entry fs.DirEntry, err error) error {
			if err != nil || entry.IsDir() {
				return nil
			}
			data, err := fs.ReadFile(root, at)
			if err != nil {
				return nil
			}
			h.Write([]byte(name + "/" + at))
			h.Write(data)
			return nil
		})
	}
	return hex.EncodeToString(h.Sum(nil))[:12]
}

// stylesheet returns the entry stylesheet with each import versioned.
func (t *Tree) stylesheet() []byte {
	data, err := t.Open(Theme, Stylesheet)
	if err != nil {
		return nil
	}
	return importURL.ReplaceAll(data, []byte(`@import url("$1?v=`+t.version+`")`))
}

// write sends a body and logs what a dropped connection cost.
func write(w http.ResponseWriter, r *http.Request, data []byte) {
	if _, err := w.Write(data); err != nil {
		slog.Debug("writing an asset", "path", r.URL.Path, "err", err)
	}
}
