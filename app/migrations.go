package app

import (
	"io/fs"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/keystore"
	"github.com/titpetric/task-ui/phpfront"
	"github.com/titpetric/task-ui/phpscript"
	"github.com/titpetric/task-ui/schema"
)

// Migrations returns the schema of the instance: every owner's migrations,
// merged flat by base name.
//
// This is the one place that names the owners. A table's migrations live with
// the code that owns it - a module's pages, the key store, the Guard, the
// module registry - and the list of those four is a fact about the program and
// not about any one of them, so it sits at the root that composes the program.
// storage.Migrate takes the filesystem rather than reading one, for the same
// reason: a package that owns tables cannot be imported by the package that
// applies them.
//
// schema.Migrations is what holds the merged names flat, which is what keeps
// an instance's migrations table valid when a file moves between owners.
//
// It answers no error. The one thing that can go wrong is two owners carrying
// one base name, and every owner here is an embedded tree: whether they do is
// settled when the binary is built, so it is a panic like phpfront's sub and
// not a condition a start has to report. TestMigrations is what catches it.
func Migrations() fs.FS {
	merged, err := schema.Migrations(
		phpfront.Migrations(),
		keystore.Migrations(),
		access.Migrations(),
		phpscript.Migrations(),
	)
	if err != nil {
		panic(err)
	}
	return merged
}
