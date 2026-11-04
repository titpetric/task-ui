package access

import (
	"embed"
	"io/fs"
)

//go:embed schema/*.up.sql
var migrations embed.FS

// Migrations returns the migrations of the tables this package owns: the
// passwords, the API tokens, the mail addresses and the sessions.
//
// The Guard is the boundary in front of every mount, and minting or reading a
// session stays Go while anything sits behind it, so no page queries these
// tables and the data model stays with the code that authenticates.
//
// schema.Migrations merges this with every other owner's, keyed by base name,
// which is what the migrations table records: see schema/schema.go.
func Migrations() fs.FS {
	return migrations
}
