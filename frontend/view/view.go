// Package view renders the pages of task-ui: the templates, the page
// model they bind to and the embedded assets. It knows nothing of HTTP
// routing; the frontend package fills a Page and renders it by name.
//
// The template set is parsed once per process, because it is not cheap:
// 49 pages over 19 partials, and a page is a clone of the whole layout
// tree, which is 16MB of parse nodes and 90ms of parsing. A server
// builds one and would never have noticed. The tests built one per
// fixture, and ninety of them held 1.5GB of identical template trees
// until the kernel stopped the test binary.
package view

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"path"
	"sort"
	"strings"
	"sync"
)

//go:embed templates/*.tpl
var templates embed.FS

// partialPrefix starts the files every page shares: _layout.tpl, the
// layout and most partials, and a file per partial of its own size, such
// as _actions.tpl.
const partialPrefix = "_"

// View holds the parsed pages, one template set per page over the
// shared layout.
type View struct {
	pages map[string]*template.Template
}

// parsed is the one template set of the process, built on the first
// call to New and kept. Sharing it is safe because a View is read only
// once it is built, the files are embedded and the functions are
// package level, and html/template executes a parsed template from any
// number of goroutines.
var parsed struct {
	sync.Once
	v   *View
	err error
}

// New answers the parsed templates, the same set every time.
func New() (*View, error) {
	parsed.Do(func() { parsed.v, parsed.err = parse() })
	return parsed.v, parsed.err
}

// parse reads the embedded templates into a page per file.
func parse() (*View, error) {
	entries, err := fs.ReadDir(templates, "templates")
	if err != nil {
		return nil, err
	}
	layout, err := template.New("layout").Funcs(Funcs()).ParseFS(templates, path.Join("templates", partialPrefix+"*.tpl"))
	if err != nil {
		return nil, err
	}
	v := &View{pages: map[string]*template.Template{}}
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, partialPrefix) || !strings.HasSuffix(name, ".tpl") {
			continue
		}
		page, err := layout.Clone()
		if err != nil {
			return nil, err
		}
		if _, err := page.ParseFS(templates, path.Join("templates", name)); err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		v.pages[strings.TrimSuffix(name, ".tpl")] = page
	}
	return v, nil
}

// Names returns the pages, in order.
func (v *View) Names() []string {
	names := make([]string, 0, len(v.pages))
	for name := range v.pages {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Render writes a page. It builds the page in full before it writes
// anything, so a template error never leaves half a page behind.
func (v *View) Render(w io.Writer, name string, page *Page) error {
	tpl, ok := v.pages[name]
	if !ok {
		return fmt.Errorf("no page %q", name)
	}
	var buf bytes.Buffer
	if err := tpl.ExecuteTemplate(&buf, "layout", page); err != nil {
		return err
	}
	_, err := buf.WriteTo(w)
	return err
}
