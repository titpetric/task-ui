package phpscript

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/go-bridget/mig/migrate"

	"github.com/titpetric/task-ui/storage"
)

// Schema applies a module's own schema into task-ui's database. It is
// what `Task::migrate("schema/")` reaches.
//
// The module declares the SQL and the instance applies it, through the
// same mig the instance's own migrations go through, so a module's
// tables are task-ui's tables and `mig docs` reads them beside the rest.
//
// Each module migrates under a project of its own name, so two modules
// cannot claim one applied row and a module added later starts from its
// own first migration rather than from wherever the instance stands.
type Schema struct {
	DB storage.DB
}

// NewSchema returns the Migrator over a pool.
func NewSchema(db storage.DB) *Schema {
	return &Schema{DB: db}
}

// Migrate applies the *.up.sql of a directory under the module's root.
//
// A directory that is not there is not an error: a module with no schema
// stores nothing, and a migrate.php in a module that has none yet should
// not keep it from registering.
func (s *Schema) Migrate(ctx context.Context, module *Module, dir string) error {
	if s == nil || s.DB == nil {
		return errors.New("this instance applies no module schema")
	}
	rel, err := Inside(module.Root, dir)
	if err != nil {
		return err
	}
	path := filepath.Join(module.Root, rel)
	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return errors.New(dir + ": a module's schema is a directory of *.up.sql")
	}
	manager, err := migrate.NewManager(s.DB, os.DirFS(path), module.Project())
	if err != nil {
		return err
	}
	_, err = manager.Apply(ctx)
	return err
}
