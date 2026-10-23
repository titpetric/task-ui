package storage

import (
	"context"
	"errors"
	"os"
)

// Copy writes a consistent copy of the database to a new file, the write-ahead log folded in.
// A file already at the path stays: Copy reports os.ErrExist and writes nothing.
func Copy(ctx context.Context, db DB, path string) error {
	if _, err := os.Stat(path); err == nil {
		return os.ErrExist
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	_, err := db.ExecContext(ctx, "VACUUM INTO ?", path)
	return err
}
