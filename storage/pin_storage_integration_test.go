//go:build integration

package storage_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

func TestPinStorage(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	pins := storage.NewPinStorage(newDB(t))

	list, err := pins.List(ctx, "claude")
	assert.NoError(t, err)
	assert.Len(t, list, 0)

	folder := &model.Pin{UserName: "claude", Folder: "malta/apps"}
	assert.NoError(t, pins.Add(ctx, folder))
	assert.NotEqual(t, "", folder.ID)
	assert.Equal(t, "[]", folder.Args)

	// The same target again is the pin there is.
	again := &model.Pin{UserName: "claude", Folder: "malta/apps"}
	assert.NoError(t, pins.Add(ctx, again))
	assert.Equal(t, folder.ID, again.ID)

	job := &model.Pin{UserName: "claude", Folder: "malta/dockerfiles", Job: "build", Dir: "task", Args: `["-v"]`}
	assert.NoError(t, pins.Add(ctx, job))
	assert.NoError(t, pins.Add(ctx, &model.Pin{UserName: "black", Folder: "malta/apps"}))
	assert.Error(t, pins.Add(ctx, &model.Pin{UserName: "claude"}))

	list, err = pins.List(ctx, "claude")
	assert.NoError(t, err)
	assert.Len(t, list, 2)
	assert.Equal(t, folder.ID, list[0].ID)
	assert.Equal(t, "task", list[1].Dir)
	assert.Equal(t, `["-v"]`, list[1].Args)

	// A pin is its user's to remove.
	assert.ErrorIs(t, pins.Delete(ctx, "black", job.ID), storage.ErrNotFound)
	assert.NoError(t, pins.Delete(ctx, "claude", job.ID))
	assert.ErrorIs(t, pins.Delete(ctx, "claude", job.ID), storage.ErrNotFound)
	list, err = pins.List(ctx, "claude")
	assert.NoError(t, err)
	assert.Len(t, list, 1)
}
