//go:build integration

package phpscript_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/internal/testdb"
	"github.com/titpetric/task-ui/phpscript"
	"github.com/titpetric/task-ui/storage"
	_ "modernc.org/sqlite"
)

// newDB opens a task-ui database of its own.
func newDB(t *testing.T) storage.DB {
	t.Helper()
	db := testdb.New(t)
	return db
}

// A module reads and writes the instance's database through the one
// connection the provider answers, and ulid() gives it the identifiers
// the schema rules want.
func TestDatabase(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	module(t, root, "titpetric/store", `<?php
$db = new Database("task-ui");
$db->query("CREATE TABLE IF NOT EXISTS probe (id CHAR(26) PRIMARY KEY NOT NULL, body TEXT NOT NULL DEFAULT '')");
$id = ulid();
$db->insert("probe", array("id" => $id, "body" => "a thread"));
$rows = $db->get_all("SELECT * FROM probe ORDER BY id");
echo "rows=", count($rows), " len=", strlen($id), "\n";
foreach ($rows as $row) { echo "row=", $row["id"], ":", $row["body"], "\n"; }
$one = $db->get("SELECT * FROM probe WHERE id = ?", $id);
echo "one=", $one["body"], "\n";
`)

	opts := phpscript.NewOptions(root)
	opts.Database = phpscript.NewDatabase(newDB(t))
	m := phpscript.New(opts)
	assert.NoError(t, m.Start(context.Background()))

	r := httptest.NewRequest(http.MethodGet, "/m/titpetric/store", nil)
	out, err := m.Render(context.Background(), httptest.NewRecorder(), r)
	assert.NoError(t, err)

	body := string(out.Body)
	assert.Contains(t, body, "rows=1 len=26")
	assert.Contains(t, body, ":a thread")
	assert.Contains(t, body, "one=a thread")
}

// A connection the instance did not configure is refused, so a module
// reaches the database it migrated into and no other.
func TestDatabaseRefuses(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	module(t, root, "titpetric/store", `<?php $db = new Database("elsewhere"); echo "opened";`)

	opts := phpscript.NewOptions(root)
	opts.Database = phpscript.NewDatabase(newDB(t))
	m := phpscript.New(opts)
	assert.NoError(t, m.Start(context.Background()))

	r := httptest.NewRequest(http.MethodGet, "/m/titpetric/store", nil)
	_, err := m.Render(context.Background(), httptest.NewRecorder(), r)
	assert.ErrorContains(t, err, "elsewhere")
}

// An instance that wired no database leaves the constructor to say so.
func TestDatabaseNone(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	module(t, root, "titpetric/store", `<?php $db = new Database("task-ui"); echo "opened";`)

	m := phpscript.New(phpscript.NewOptions(root))
	assert.NoError(t, m.Start(context.Background()))

	r := httptest.NewRequest(http.MethodGet, "/m/titpetric/store", nil)
	_, err := m.Render(context.Background(), httptest.NewRecorder(), r)
	assert.Error(t, err)
}
