package phpfront

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
)

// ErrName is returned when a module's name is not <vendor>/<plugin>.
var ErrName = errors.New("a module name is vendor/plugin")

// ModuleRule is the permission a module's sidebar row needs unless it sets one.
const ModuleRule = "module"

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

	// Tables are the tables the module bundles: the ones its own schema/
	// creates, which nothing else may create. A table it reads and does not
	// own is not one of these.
	Tables []string

	// Sidebar is whether the module's row belongs in the sidebar, and
	// SidebarHref the link it opens. SidebarRule is the permission the row
	// needs. A module that never enables one is served but not listed.
	Sidebar     bool
	SidebarHref string
	SidebarRule string
}

// NewModule returns the empty module `new Task\Module` builds, which names
// itself through the setters. The context is the interpreter's.
func NewModule(_ context.Context) (*Module, error) {
	return &Module{SidebarRule: ModuleRule}, nil
}

// SetName splits <vendor>/<plugin> into the two halves and refuses anything else.
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

// SetTables names the tables the module bundles, the array PHP hands over.
func (m *Module) SetTables(tables any) {
	m.Tables = names(tables)
	sort.Strings(m.Tables)
}

// EnableSideBar puts the module's row in the sidebar.
func (m *Module) EnableSideBar() { m.Sidebar = true }

// SetSideBarHref sets the link the sidebar row opens.
func (m *Module) SetSideBarHref(href string) { m.SidebarHref = strings.TrimSpace(href) }

// SetSideBarRule sets the permission the sidebar row needs.
func (m *Module) SetSideBarRule(rule string) { m.SidebarRule = strings.TrimSpace(rule) }

// GetSidebar reports whether the module enabled a sidebar row.
func (m *Module) GetSidebar() bool { return m.Sidebar }

// GetSideBarHref returns the link the sidebar row opens.
func (m *Module) GetSideBarHref() string {
	if m.SidebarHref != "" {
		return m.SidebarHref
	}
	return m.Href()
}

// GetSideBarRule returns the permission the sidebar row needs.
func (m *Module) GetSideBarRule() string { return m.SidebarRule }

// GetTitle returns the title of the sidebar row.
func (m *Module) GetTitle() string { return m.Label() }

// GetDescription returns the line under the sidebar row.
func (m *Module) GetDescription() string { return m.Description }

// GetIcon returns the sprite symbol of the sidebar row.
func (m *Module) GetIcon() string { return m.Icon }

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

// ProjectLimit is how long a migration project name may be: the column
// mig records one in holds sixteen characters.
const ProjectLimit = 16

// Project is the migration project name the module's schema is applied under.
func (m *Module) Project() string {
	name := m.Name()
	if name == "" {
		return ""
	}
	digest := sha256.Sum256([]byte(name))
	short := m.Plugin
	if len(short) > 7 {
		short = short[:7]
	}
	return short + "_" + hex.EncodeToString(digest[:])[:ProjectLimit-len(short)-1]
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
