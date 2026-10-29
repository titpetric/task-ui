package phpscript

import (
	"context"
)

// Registry is where a module lands once its modinfo.php registered it: the
// module table, and the sidebar row that links it.
type Registry interface {
	// Register upserts a module that passed Validate. A module already in the
	// table keeps its row, its identifier and its place in the sidebar, and
	// takes the rest from the modinfo.php just read.
	Register(ctx context.Context, m *Module) error

	// Forget takes the names of every module a scan found and drops the rows
	// of the others, whose directories are gone. It returns how many.
	Forget(ctx context.Context, found []string) (int, error)
}
