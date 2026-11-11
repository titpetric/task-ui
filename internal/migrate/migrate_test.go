package migrate_test

import (
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/internal/migrate"
)

// owner is one owner's tree, at whatever depth it carries its migrations.
func owner(files map[string]string) fs.FS {
	tree := fstest.MapFS{}
	for name, contents := range files {
		tree[name] = &fstest.MapFile{Data: []byte(contents)}
	}
	return tree
}

// The merge flattens: a file comes out under its base name whether the owner
// kept it in a module's schema/ or at its own root, and that name is what the
// migrations table records.
func TestMigrationsFlattens(t *testing.T) {
	t.Parallel()
	merged, err := migrate.Migrations(
		owner(map[string]string{
			"notes/schema/note.up.sql":  "CREATE TABLE note",
			"runs/schema/run.up.sql":    "CREATE TABLE run",
			"notes/templates/index.tpl": "not a migration",
			"notes/schema/schema.puml":  "not a migration",
			"runs/schema/run.down.sql":  "not an up migration",
			"runs/README.md":            "not a migration",
			"deep/er/still/user.up.sql": "CREATE TABLE user",
		}),
		owner(map[string]string{"schema/key_info.up.sql": "CREATE TABLE key_info"}),
	)
	assert.NoError(t, err)

	files, err := fs.Glob(merged, "*.up.sql")
	assert.NoError(t, err)
	assert.Equal(t, []string{"key_info.up.sql", "note.up.sql", "run.up.sql", "user.up.sql"}, files,
		"flat, under the base name, in the order a run applies them")

	data, err := fs.ReadFile(merged, "note.up.sql")
	assert.NoError(t, err)
	assert.Equal(t, "CREATE TABLE note", string(data))
}

// Two owners carrying one base name is an error and not a file one of them
// shadows: the loser's table would never be created, and nothing would say so.
func TestMigrationsRefusesTwoOwners(t *testing.T) {
	t.Parallel()
	_, err := migrate.Migrations(
		owner(map[string]string{"notes/schema/note.up.sql": "CREATE TABLE note"}),
		owner(map[string]string{"schema/note.up.sql": "CREATE TABLE note"}),
	)
	assert.ErrorContains(t, err, "note.up.sql is carried by two owners")
}

// A merge finding nothing is an error. The owners are embedded trees, so an
// empty merge means a //go:embed stopped matching, and a start that migrates
// nothing leaves an empty database every query then fails against.
func TestMigrationsRefusesEmpty(t *testing.T) {
	t.Parallel()
	_, err := migrate.Migrations()
	assert.ErrorContains(t, err, "no migrations in 0 owners")

	_, err = migrate.Migrations(owner(map[string]string{"notes/README.md": "no SQL here"}))
	assert.ErrorContains(t, err, "no migrations in 1 owners")
}

// The merged filesystem answers the whole of fs.FS, because mig reads the root
// directory and stats each entry before it reads a file.
func TestMigrationsIsAWholeFS(t *testing.T) {
	t.Parallel()
	merged, err := migrate.Migrations(owner(map[string]string{"notes/schema/note.up.sql": "CREATE TABLE note"}))
	assert.NoError(t, err)

	entries, err := fs.ReadDir(merged, ".")
	assert.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, "note.up.sql", entries[0].Name())

	info, err := fs.Stat(merged, "note.up.sql")
	assert.NoError(t, err)
	assert.Equal(t, int64(len("CREATE TABLE note")), info.Size())
}
