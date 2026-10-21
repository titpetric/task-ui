// Package view renders the pages of task-ui: the templates, the page
// model they bind to and the embedded assets. It knows nothing of HTTP
// routing; the frontend package fills a Page and renders it by name.
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

// New parses the embedded templates.
func New() (*View, error) {
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
