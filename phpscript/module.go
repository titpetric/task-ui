// Package phpscript is the plugin boundary: a module is a directory with a
// modinfo.php in it, run by the phpscript interpreter the binary holds, served
// at /m/<vendor>/<name> as the body of an ordinary task-ui page.
//
// A Manager finds the modules under its roots, runs each modinfo.php once to
// learn what it registers, and keeps a pool of runtimes per module to serve
// requests from. Four host symbols reach the PHP: the Task\Module constructor
// a modinfo.php builds, Task::register, which records it, Task::migrate, which
// applies the module's own schema, and Task::dispatch, which starts a Claude
// run. Which module is calling is not an argument of any of them; it is on the
// runtime's context, so the four bindings are the same for every module.
//
// The handler rewrites nothing. $_SERVER["REQUEST_URI"] holds the path the
// browser requested, the /m/<vendor>/<name> prefix included, and a module
// builds its links from what it reads. Writing one is docs/modules.md.
package phpscript

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// ErrName is returned when a module's name is not <vendor>/<plugin>.
var ErrName = errors.New("a module name is vendor/plugin")

// Module is what a modinfo.php registers: the name that decides its link and
// its row in the module table, the directory holding it, and what the sidebar
// shows. A PHP script builds one with `new Task\Module` and fills it through
// the setters, so every field has one.
//
// Root is not a setter. The Manager sets it to the directory the modinfo.php
// sits in, so a module cannot claim a directory it does not live in.
type Module struct {
	// Vendor and Plugin are the two halves of the name, as SetName split it.
	Vendor string
	Plugin string

	// Root is the directory holding the modinfo.php, absolute.
	Root string

	// Title, Description and Icon are the sidebar row and the module's page
	// heading. Icon names a symbol of the sprite without the i- prefix.
	Title       string
	Description string
	Icon        string

	// Version is whatever the module calls its own, for the module table.
	Version string
}

// NewModule returns the empty module `new Task\Module` builds, which names
// itself through the setters. The context is the interpreter's.
func NewModule(_ context.Context) (*Module, error) {
	return &Module{}, nil
}

// SetName splits <vendor>/<plugin> into the two halves and refuses anything
// else. It is the one setter that can fail, so a modinfo.php that names itself
// wrongly throws where it wrote the name.
func (m *Module) SetName(name string) error {
	vendor, plugin, err := ParseName(name)
	if err != nil {
		return err
	}
	m.Vendor, m.Plugin = vendor, plugin
	return nil
}

// SetTitle sets the title of the sidebar row and the page heading.
func (m *Module) SetTitle(title string) { m.Title = strings.TrimSpace(title) }

// SetDescription sets the line under the heading.
func (m *Module) SetDescription(description string) {
	m.Description = strings.TrimSpace(description)
}

// SetIcon sets the sprite symbol of the sidebar row, without the i- prefix.
func (m *Module) SetIcon(icon string) { m.Icon = strings.TrimSpace(icon) }

// SetVersion sets what the module calls its own version.
func (m *Module) SetVersion(version string) { m.Version = strings.TrimSpace(version) }

// Name is the module's name, <vendor>/<plugin>, and the unique key of its row.
func (m *Module) Name() string {
	if m.Vendor == "" || m.Plugin == "" {
		return ""
	}
	return m.Vendor + "/" + m.Plugin
}

// Href is the link the sidebar row points at.
func (m *Module) Href() string {
	if name := m.Name(); name != "" {
		return Prefix + "/" + name
	}
	return ""
}

// Label is the title, falling back to the plugin half of the name, so a module
// that set none still reads as something in the sidebar.
func (m *Module) Label() string {
	if m.Title != "" {
		return m.Title
	}
	return m.Plugin
}

// Validate reports what keeps a module out of the table: a name it never set,
// and a root the Manager could not fill.
func (m *Module) Validate() error {
	if m.Name() == "" {
		return fmt.Errorf("%w: the module registered no name", ErrName)
	}
	if m.Root == "" {
		return errors.New("the module has no root directory")
	}
	return nil
}
