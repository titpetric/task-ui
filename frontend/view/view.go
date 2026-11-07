// Package view renders the pages of task-ui: the templates, the page
// model they bind to and the embedded assets. It knows nothing of HTTP
// routing; the frontend package fills a Page and renders it by name.
//
// One set holds everything: the layout, the 20 partials and the content
// block of each of the 49 pages, named for its page. A page block is
// executed on its own and the layout prints the result as Page.Body, so
// no page owns a copy of the partials: 49 clones of 164 KB of partials
// were 40 MB of parse nodes the process kept.
//
// The set is parsed once per process, on the first call to New.
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

// partialPrefix starts the files every page shares: _layout.tpl with the
// layout and most partials, then one file per partial big enough to keep
// apart, such as _actions.tpl.
const partialPrefix = "_"

// contentBlock is the block a page file defines, which the set keeps by
// the page's name.
const contentBlock = "content"

// View holds the parsed set and the content block of every page in it.
type View struct {
	set   *template.Template
	pages map[string]*template.Template
}

// parsed is the template set of the process, built on the first call to
// New and kept. Sharing it is safe because a View is read only once it
// is built, the files are embedded and the functions are package level,
// and html/template executes a parsed template from any number of
// goroutines.
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

// parse reads the partials into one set, then every page file into the
// same set.
func parse() (*View, error) {
	entries, err := fs.ReadDir(templates, "templates")
	if err != nil {
		return nil, err
	}
	set, err := template.New("layout").Funcs(Funcs()).ParseFS(templates, path.Join("templates", partialPrefix+"*.tpl"))
	if err != nil {
		return nil, err
	}
	v := &View{set: set, pages: map[string]*template.Template{}}
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, partialPrefix) || !strings.HasSuffix(name, ".tpl") {
			continue
		}
		if err := v.add(strings.TrimSuffix(name, ".tpl"), name); err != nil {
			return nil, err
		}
	}
	return v, nil
}

// add parses one page file into the set: its content block by the page's
// name, every other block it defines by its own. It refuses a file that
// defines no content block, and a block name the set already holds.
func (v *View) add(page, file string) error {
	parsed, err := template.New(file).Funcs(Funcs()).ParseFS(templates, path.Join("templates", file))
	if err != nil {
		return fmt.Errorf("%s: %w", file, err)
	}
	for _, block := range parsed.Templates() {
		name := block.Name()
		switch {
		case name == file:
			continue
		case name == contentBlock:
			name = page
		}
		if v.set.Lookup(name) != nil {
			return fmt.Errorf("%s: %q is defined twice", file, name)
		}
		if _, err := v.set.AddParseTree(name, block.Tree); err != nil {
			return fmt.Errorf("%s: %w", file, err)
		}
	}
	content := v.set.Lookup(page)
	if content == nil {
		return fmt.Errorf("%s: defines no %q", file, contentBlock)
	}
	v.pages[page] = content
	return nil
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

// Render writes a page: its content block, then the layout printing
// that as Page.Body. It builds the page in full before it writes
// anything, so a template error never leaves half a page behind.
func (v *View) Render(w io.Writer, name string, page *Page) error {
	content, ok := v.pages[name]
	if !ok {
		return fmt.Errorf("no page %q", name)
	}
	var body bytes.Buffer
	if err := content.Execute(&body, page); err != nil {
		return err
	}
	page.Body = template.HTML(body.String())
	var buf bytes.Buffer
	if err := v.set.ExecuteTemplate(&buf, "layout", page); err != nil {
		return err
	}
	_, err := buf.WriteTo(w)
	return err
}
