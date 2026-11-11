package phpfront

import (
	"embed"
	"io/fs"
)

//go:embed schema/*.up.sql
var migrations embed.FS

// ModuleMigrations returns the migrations of the table this package owns: the
// registry of the modules it discovered, one row per modinfo.php.
//
// A module declares no migrations of its own. It used to, through
// `Task::migrate` and a project name of its own, and that is gone: every table
// of the instance is migrated once at start, out of the trees app.Migrations
// merges, under the one project name. A module's pages own tables by carrying
// the files under modules/<name>/schema/, which phpfront embeds.
//
// schema.Migrations merges this with every other owner's, keyed by base name,
// which is what the migrations table records: see schema/schema.go.
func ModuleMigrations() fs.FS {
	return migrations
}
