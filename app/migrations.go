package app

import (
	"io/fs"

	"github.com/titpetric/task-ui/integration/ssh/keystore"
	"github.com/titpetric/task-ui/internal/access"
	"github.com/titpetric/task-ui/internal/migrate"
	"github.com/titpetric/task-ui/phpfront"
)

// Migrations returns the schema of the instance: every owner's migrations,
// merged flat by base name.
//
// This is the one place that names the owners. A table's migrations live with
// the code that owns it (a module's pages, the key store, the Guard, the module
// registry); the list of those four belongs to the program and to no single
// owner, so it sits at the root that composes the program. storage.Migrate
// takes the filesystem it applies for the same reason: a package that owns
// tables cannot be imported by the package that applies them.
//
// migrate.Migrations flattens the merged names, which keeps an instance's
// migrations table valid when a file moves between owners.
//
// It answers no error. Two owners can collide on one base name, and every
// owner here is an embedded tree, so the collision is settled at build time and
// panics like phpfront's sub. TestMigrations catches it.
func Migrations() fs.FS {
	merged, err := migrate.Migrations(
		phpfront.Migrations(),
		keystore.Migrations(),
		access.Migrations(),
		phpfront.ModuleMigrations(),
	)
	if err != nil {
		panic(err)
	}
	return merged
}
