package phpfront

import (
	"context"

	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/storage"
)

// Store is the Registry over the database: a module lands in the module
// table, and a module a scan no longer finds loses its row.
type Store struct {
	Modules *storage.ModuleStorage
}

// NewStore returns the Registry over the table.
func NewStore(modules *storage.ModuleStorage) *Store {
	return &Store{Modules: modules}
}

// Register records a module.
func (s *Store) Register(ctx context.Context, m *Module) error {
	if err := m.Validate(); err != nil {
		return err
	}
	return s.Modules.Save(ctx, &model.Module{
		Name:        m.Name(),
		Vendor:      m.Vendor,
		Plugin:      m.Plugin,
		Root:        m.Root,
		Title:       m.Label(),
		Description: m.Description,
		Icon:        m.Icon,
		Version:     m.Version,
	})
}

// Forget drops the rows of every module the scan did not find, and returns how
// many went.
func (s *Store) Forget(ctx context.Context, found []*Module) (int, error) {
	names := make([]string, 0, len(found))
	for _, module := range found {
		names = append(names, module.Name())
	}
	return s.Modules.Keep(ctx, names)
}
