// Package migrate merges the SQL migrations of task-ui and holds what is
// generated from them: schema.yml.
//
// The migrations are the source of truth for the data model. One file owns one
// table, and files are append-only once applied anywhere: a later change
// belongs in a new *.up.sql file.
//
// Each file lives with the code that owns its table, and nowhere else:
//
//	phpfront/app/modules/<name>/schema/   the tables a module's pages own
//	keystore/schema/                      the keys and who may read them
//	access/schema/                        the passwords, tokens and sessions
//	phpscript/schema/                     the module registry
//
// Migrations is the merged view of every owner, flat, keyed by base name. Flat
// is the load-bearing part: the migrations table records the name of a file
// relative to the root of the filesystem it was applied from, so a file that
// gains a directory is a file nobody ran. It re-applies, and the second
// ALTER TABLE of a column that is already there takes the database with it.
// Keeping the merged view flat lets a migration move between owners without
// touching a single instance.
//
// A base name therefore belongs to one owner. Two owners carrying the same one
// is an error Migrations reports, never a file one owner shadows.
//
// Run `atkins schema` to regenerate ./schema.yml, ./docs and ../model after
// adding a migration anywhere, then `atkins models` in phpfront/app for the
// PHP side of the same schema.yml.
package migrate

import (
	"fmt"
	"io/fs"
	"path"
	"testing/fstest"
)

// Pattern selects the migrations of an owner's tree, at whatever depth they
// sit: the module tree carries them under <name>/schema/ and a Go package
// under schema/. It is the glob mig matches against base names, which is how
// the two layouts read as one.
const Pattern = "*.up.sql"

// Migrations returns every owner's migrations as one flat filesystem, each
// file at its base name.
//
// The order a run applies them in is the order of those names, which is the
// order every instance has already applied: see the package comment for what
// rests on that.
func Migrations(owners ...fs.FS) (fs.FS, error) {
	// fstest.MapFS, not a type of this package: mig walks the whole of fs.FS.
	merged := fstest.MapFS{}
	for _, owner := range owners {
		if err := collect(merged, owner); err != nil {
			return nil, err
		}
	}
	if len(merged) == 0 {
		return nil, fmt.Errorf("migrate: no migrations in %d owners", len(owners))
	}
	return merged, nil
}

// collect reads one owner's migrations into merged, under their base names.
func collect(merged fstest.MapFS, owner fs.FS) error {
	return fs.WalkDir(owner, ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		base := path.Base(name)
		if ok, _ := path.Match(Pattern, base); !ok {
			return nil
		}
		if _, taken := merged[base]; taken {
			return fmt.Errorf("migrate: %s is carried by two owners", base)
		}
		contents, err := fs.ReadFile(owner, name)
		if err != nil {
			return err
		}
		merged[base] = &fstest.MapFile{Data: contents}
		return nil
	})
}
