// Package testdb hands a test a migrated database without migrating one.
//
// A fixture used to call storage.OpenAndMigrate(ctx, ":memory:") and pay
// for all eighty migrations: 64ms, measured, every time. Across the
// integration suite that is most of a minute spent writing the same
// schema into a database nobody kept.
//
// The schema is the same for every test, so it is built once per process
// into a template file and copied. A copy and an open is 0.22ms, which is
// 292 times less, and the database a test gets is its own file with
// nothing shared.
package testdb

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/titpetric/task-ui/storage"
)

// template is the migrated database every test is copied from, built on
// the first call and kept for the life of the test binary.
var template struct {
	sync.Once
	path string
	err  error
}

// New is a migrated database of this test's own, removed when it ends.
//
// It is a file rather than :memory: because a file is what can be copied.
// The pragmas are the ones a test wants and a server does not: a database
// that lives for one test has nothing to recover and no reason to wait
// for the disk.
func New(t *testing.T) storage.DB {
	t.Helper()
	from, err := seed()
	if err != nil {
		t.Fatalf("the template database: %v", err)
	}
	to := filepath.Join(t.TempDir(), "task-ui.db")
	if err := copyFile(from, to); err != nil {
		t.Fatalf("copying the template database: %v", err)
	}
	db, err := storage.Open(to + "?_pragma=journal_mode(MEMORY)&_pragma=synchronous(OFF)")
	if err != nil {
		t.Fatalf("opening the database: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

// seed builds the template on the first call and answers its path after.
// The directory is the test binary's own and the operating system clears
// it; a test binary that crashes leaves one small file behind.
func seed() (string, error) {
	template.Do(func() {
		dir, err := os.MkdirTemp("", "task-ui-schema")
		if err != nil {
			template.err = err
			return
		}
		path := filepath.Join(dir, "template.db")
		db, err := storage.OpenAndMigrate(context.Background(), path)
		if err != nil {
			template.err = err
			return
		}
		if err := db.Close(); err != nil {
			template.err = err
			return
		}
		template.path = path
	})
	return template.path, template.err
}

// copyFile writes one file over another, the whole of it.
func copyFile(from, to string) error {
	in, err := os.Open(from)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(to)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}
