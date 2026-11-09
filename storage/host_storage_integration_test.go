//go:build integration

package storage_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/storage"
)

// The table is read by name, and Names keys it by the address a run on the
// hosts signs in to. The writes are the Inventory page's.
func TestHostStorage(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := newDB(t)
	hosts := storage.NewHostStorage(db)

	list, err := hosts.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, list, 0)
	assert.Len(t, hosts.Names(ctx), 0)

	// The rows go in the way the page writes them, which is SQL of its own.
	_, err = db.ExecContext(ctx, "INSERT INTO host (id, name, address, source, login, added_by) VALUES "+
		"('1', 'web', 'web.example', 'beszel', '', ''), ('2', 'box', '10.0.0.2', 'manual', 'debian@10.0.0.2', 'black')")
	assert.NoError(t, err)

	list, err = hosts.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, list, 2)
	assert.Equal(t, "box", list[0].Name)
	assert.Equal(t, "debian@10.0.0.2", list[0].Login)

	names := hosts.Names(ctx)
	assert.Len(t, names, 2)
	assert.Equal(t, "box", names["10.0.0.2"])
	assert.Equal(t, "web", names["web.example"])
}
