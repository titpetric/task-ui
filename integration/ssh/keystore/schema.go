package keystore

import (
	"embed"
	"io/fs"
)

//go:embed schema/*.up.sql
var migrations embed.FS

// Migrations returns the migrations of the tables this package owns: the key
// records, who may read each key and the workspace grants.
//
// No page queries them. The key material never enters PHP, so the front end
// reads keys through a host binding and the data model stays here with the
// store that writes it.
//
// migrate.Migrations merges this with every other owner's, keyed by base name,
// the key the migrations table records: see internal/migrate.
func Migrations() fs.FS {
	return migrations
}
