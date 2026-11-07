package access

import (
	"embed"
	"io/fs"
)

//go:embed schema/*.up.sql
var migrations embed.FS

// Migrations returns the migrations of the tables this package owns, as the package doc describes them.
func Migrations() fs.FS {
	return migrations
}
