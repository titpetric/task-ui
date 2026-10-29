//go:build integration

package storage_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// A module saves by name and reads back by name, and a name already in
// the table keeps its row: a start reads every modinfo.php again, and a
// module whose title changed is the same module.
func TestModuleStorage(t *testing.T) {
	t.Parallel()
	s := storage.NewModuleStorage(newDB(t))
	ctx := context.Background()

	rows, err := s.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, rows, 0)

	_, err = s.Get(ctx, "titpetric/refinement")
	assert.ErrorIs(t, err, storage.ErrNotFound)

	first := &model.Module{
		Name: "titpetric/refinement", Vendor: "titpetric", Plugin: "refinement",
		Root: "/data/modules/titpetric/refinement", Title: "Refinement",
		Description: "Refine a design", Icon: "flow", Version: "0.1.0",
	}
	assert.NoError(t, s.Save(ctx, first))
	assert.NotEqual(t, "", first.ID, "Save fills the identifier it wrote")
	assert.NotNil(t, first.CreatedAt)

	read, err := s.Get(ctx, "titpetric/refinement")
	assert.NoError(t, err)
	assert.Equal(t, "Refinement", read.Title)
	assert.Equal(t, "0.1.0", read.Version)
	assert.Equal(t, "/data/modules/titpetric/refinement", read.Root)

	// The same name again: the row, its identifier and its created date
	// stay, and everything the modinfo.php named is replaced.
	again := &model.Module{
		Name: "titpetric/refinement", Vendor: "titpetric", Plugin: "refinement",
		Root: "/srv/modules/titpetric/refinement", Title: "Refinement, renamed",
		Icon: "box", Version: "0.2.0",
	}
	assert.NoError(t, s.Save(ctx, again))
	assert.Equal(t, first.ID, again.ID, "a module already in the table keeps its row")
	assert.Equal(t, first.CreatedAt, again.CreatedAt)
	assert.Equal(t, "Refinement, renamed", again.Title)
	assert.Equal(t, "/srv/modules/titpetric/refinement", again.Root)
	assert.Equal(t, "", again.Description, "a field the modinfo.php stopped naming is cleared")

	rows, err = s.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, rows, 1)
}

// List is by name, so two starts read the modules in one order.
func TestModuleStorage_List(t *testing.T) {
	t.Parallel()
	s := storage.NewModuleStorage(newDB(t))
	ctx := context.Background()

	for _, name := range []string{"titpetric/refinement", "acme/notes", "titpetric/mysql"} {
		assert.NoError(t, s.Save(ctx, &model.Module{Name: name, Root: "/x/" + name}))
	}
	rows, err := s.List(ctx)
	assert.NoError(t, err)
	assert.Equal(t, []string{"acme/notes", "titpetric/mysql", "titpetric/refinement"},
		[]string{rows[0].Name, rows[1].Name, rows[2].Name})
}

// Keep drops the modules a scan did not find, which is a directory gone.
func TestModuleStorage_Keep(t *testing.T) {
	t.Parallel()
	s := storage.NewModuleStorage(newDB(t))
	ctx := context.Background()

	for _, name := range []string{"a/one", "a/two", "b/three"} {
		assert.NoError(t, s.Save(ctx, &model.Module{Name: name, Root: "/x/" + name}))
	}

	dropped, err := s.Keep(ctx, []string{"a/one", "b/three"})
	assert.NoError(t, err)
	assert.Equal(t, 1, dropped)

	rows, err := s.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, rows, 2)

	// A scan that found nothing empties the table.
	dropped, err = s.Keep(ctx, nil)
	assert.NoError(t, err)
	assert.Equal(t, 2, dropped)

	rows, err = s.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, rows, 0)
}
