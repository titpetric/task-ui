package phpscript

import (
	"context"
)

// Migrator applies a module's own schema into task-ui's database. The Manager
// holds one and `Task::migrate` reaches it, so the SQL a module ships is
// applied by the same mig the instance's own migrations are.
//
// dir is relative to the module's root, and a path leaving the root is
// refused: a module migrates its own schema and no other.
type Migrator interface {
	Migrate(ctx context.Context, m *Module, dir string) error
}
