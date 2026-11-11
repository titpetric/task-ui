package phpscript

import (
	"context"

	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// Store is the Registry over the database: a module lands in the module
// table and gets a row in the sidebar, and a module a scan no longer finds
// loses both.
type Store struct {
	Modules *storage.ModuleStorage
	Sidebar *storage.SidebarStorage
}

// NewStore returns the Registry over the two tables.
func NewStore(modules *storage.ModuleStorage, sidebar *storage.SidebarStorage) *Store {
	return &Store{Modules: modules, Sidebar: sidebar}
}

// Register records a module and the sidebar row that links it.
//
// The row keeps its position and its hidden flag, so a module registering
// on every start does not walk to the end of the sidebar, and one the user
// hid stays hidden. The title and the icon are the modinfo.php's, so
// editing that file and restarting is how a module renames its own row.
func (s *Store) Register(ctx context.Context, m *Module) error {
	if err := m.Validate(); err != nil {
		return err
	}
	row := &model.Module{
		Name:        m.Name(),
		Vendor:      m.Vendor,
		Plugin:      m.Plugin,
		Root:        m.Root,
		Title:       m.Label(),
		Description: m.Description,
		Icon:        m.Icon,
		Version:     m.Version,
	}
	if err := s.Modules.Save(ctx, row); err != nil {
		return err
	}
	return s.Sidebar.Save(ctx, &model.Sidebar{
		Kind:  model.SidebarModule,
		Href:  m.Href(),
		Title: m.Label(),
		Icon:  m.Icon,
		Rule:  ModuleRule,
		// A module registering for the first time goes last, after the
		// views and after the modules already there. EnsureDefaults filled
		// the views from zero, so a position past them is past everything.
		Position: sidebarEnd,
	})
}

// ModuleRule is the permission a module's sidebar row needs. Every module
// shares it, so granting a role the modules grants every module: a plugin
// is installed by putting its directory under a root, and a rule per
// module would be a rule nobody wrote down.
const ModuleRule = "module"

// sidebarEnd is the position a row takes when it is new. It is past every
// view the defaults filled and past any order a user has saved, which is
// a position per row from zero, so a new module lands at the end and the
// sidebar page is where it moves from.
const sidebarEnd = 1000

// Forget drops the module rows and the sidebar rows of every module the
// scan did not find, and returns how many rows went.
func (s *Store) Forget(ctx context.Context, found []string) (int, error) {
	hrefs := make([]string, 0, len(found))
	for _, name := range found {
		hrefs = append(hrefs, Prefix+"/"+name)
	}
	modules, err := s.Modules.Keep(ctx, found)
	if err != nil {
		return 0, err
	}
	rows, err := s.Sidebar.Keep(ctx, hrefs)
	if err != nil {
		return 0, err
	}
	return modules + rows, nil
}
