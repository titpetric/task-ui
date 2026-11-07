package storage

import (
	"context"

	"github.com/titpetric/task-ui/model"
)

// ModuleStorage keeps the plugins the instance found, a row per module,
// as its modinfo.php registered it.
type ModuleStorage struct {
	db DB
}

// NewModuleStorage returns the storage over a pool.
func NewModuleStorage(db DB) *ModuleStorage {
	return &ModuleStorage{db: db}
}

// List returns every module by name, unpaged: the caller gets the whole table.
func (s *ModuleStorage) List(ctx context.Context) ([]model.Module, error) {
	return client(s.db).Select[model.Module](ctx, "SELECT * FROM module ORDER BY name")
}

// Get returns the module of a name, ErrNotFound when no module has it.
func (s *ModuleStorage) Get(ctx context.Context, name string) (*model.Module, error) {
	row, err := client(s.db).Get[model.Module](ctx, "SELECT * FROM module WHERE name = ?", name)
	return row, notFound(err)
}

// Save stores a module over the row its name holds, keeping that row's id
// and created_at. It reads the row back into module, so the ID and the
// dates are the row's whichever branch the upsert took.
func (s *ModuleStorage) Save(ctx context.Context, module *model.Module) error {
	if module.ID == "" {
		module.ID = model.NewID()
	}
	now := stamp()
	err := client(s.db).Exec(ctx,
		"INSERT INTO module (id, name, vendor, plugin, root, title, description, icon, version, created_at, updated_at)"+
			" VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"+
			" ON CONFLICT(name) DO UPDATE SET vendor = excluded.vendor, plugin = excluded.plugin,"+
			" root = excluded.root, title = excluded.title, description = excluded.description,"+
			" icon = excluded.icon, version = excluded.version, updated_at = excluded.updated_at",
		module.ID, module.Name, module.Vendor, module.Plugin, module.Root,
		module.Title, module.Description, module.Icon, module.Version, now, now)
	if err != nil {
		return err
	}
	stored, err := s.Get(ctx, module.Name)
	if err != nil {
		return err
	}
	*module = *stored
	return nil
}

// Keep deletes every module whose name is not in names, and returns how
// many it dropped. A scan that found nothing empties the table.
func (s *ModuleStorage) Keep(ctx context.Context, names []string) (int, error) {
	db := client(s.db)
	if err := db.Exec(ctx, "DELETE FROM module"+notIn("WHERE", "name", len(names)), anyList(names)...); err != nil {
		return 0, err
	}
	return int(db.RowsAffected()), nil
}
