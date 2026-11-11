package app

import (
	"io/fs"

	"github.com/titpetric/task-ui/integration/ssh/keystore"
	"github.com/titpetric/task-ui/internal/migrate"
	"github.com/titpetric/task-ui/phpfront"
)

// Migrations returns the schema of the instance, merged flat by base name.
// The owners are named here and nowhere else; internal/migrate says why flat.
func Migrations() fs.FS {
	merged, err := migrate.Migrations(
		phpfront.Migrations(),
		keystore.Migrations(),
		phpfront.ModuleMigrations(),
	)
	if err != nil {
		panic(err)
	}
	return merged
}
