//go:build integration

package storage_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/app/model"
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

// Pins follow a project that moves: a folder pin and a job pin of the
// project, or of a directory below the project, get the new ID, for every
// user; a pin the user holds at the new target already takes the moving
// pin's place, and a folder whose ID merely starts the same stays.
func TestPinStorage_MoveFolder(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	pins := storage.NewPinStorage(newDB(t))
	for _, pin := range []*model.Pin{
		{UserName: "claude", Folder: "web"},
		{UserName: "claude", Folder: "web/docs", Job: "build", Dir: "site"},
		{UserName: "black", Folder: "web", Job: "deploy"},
		{UserName: "black", Folder: "website"},
		{UserName: "black", Folder: "acme/web", Job: "deploy"},
	} {
		assert.NoError(t, pins.Add(ctx, pin))
	}

	moved, err := pins.MoveFolder(ctx, "web", "acme/web")
	assert.NoError(t, err)
	assert.Equal(t, 2, moved, "the pin black holds at acme/web already stays, the moving one goes")

	list, err := pins.List(ctx, "claude")
	assert.NoError(t, err)
	assert.Len(t, list, 2)
	assert.Equal(t, "acme/web", list[0].Folder)
	assert.Equal(t, "acme/web/docs", list[1].Folder)
	assert.Equal(t, "build", list[1].Job)
	list, err = pins.List(ctx, "black")
	assert.NoError(t, err)
	assert.Len(t, list, 2)
	assert.Equal(t, "website", list[0].Folder)
	assert.Equal(t, "acme/web", list[1].Folder)

	moved, err = pins.MoveFolder(ctx, "nowhere", "acme/nowhere")
	assert.NoError(t, err)
	assert.Equal(t, 0, moved)
}

// Every method returns the driver's error. Add and MoveFolder look for the
// pin a user has already: a dead pool answers that with an error, not with
// "not there", and neither writes.
func TestPinStorage_Closed(t *testing.T) {
	t.Parallel()
	pins := storage.NewPinStorage(closedDB(t))
	ctx := context.Background()

	_, err := pins.List(ctx, "admin")
	assert.ErrorContains(t, err, errClosed)
	assert.ErrorContains(t, pins.Add(ctx, &model.Pin{UserName: "admin", Folder: "acme"}), errClosed)
	assert.ErrorContains(t, pins.Delete(ctx, "admin", "p1"), errClosed)
	_, err = pins.MoveFolder(ctx, "acme", "acme2")
	assert.ErrorContains(t, err, errClosed)
}

// A move is two statements and stops at the one that failed: the delete
// of the pins whose user holds the target, and the update that moves
// the rest.
func TestPinStorage_MoveFolderWrites(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	moving := newDB(t)
	pins := storage.NewPinStorage(moving)
	assert.NoError(t, pins.Add(ctx, &model.Pin{UserName: "admin", Folder: "acme", Job: "build"}))
	refuse(t, moving, model.PinTable, "UPDATE")
	_, err := pins.MoveFolder(ctx, "acme", "acme2")
	assert.ErrorContains(t, err, errRefused)

	dropping := newDB(t)
	held := storage.NewPinStorage(dropping)
	assert.NoError(t, held.Add(ctx, &model.Pin{UserName: "admin", Folder: "acme", Job: "build"}))
	assert.NoError(t, held.Add(ctx, &model.Pin{UserName: "admin", Folder: "acme2", Job: "build"}))
	refuse(t, dropping, model.PinTable, "DELETE")
	_, err = held.MoveFolder(ctx, "acme", "acme2")
	assert.ErrorContains(t, err, errRefused)
}
