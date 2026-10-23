package storage_test

import (
	"context"
	"testing"
	"time"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// A run's reads list by secret, newest first, count and name the last
// read; the log keeps the newest of each secret.
func TestSecretReadStorage(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	reads := storage.NewSecretReadStorage(newDB(t))
	reads.Keep = 3

	run := func(id string, secrets ...string) {
		var rows []model.SecretRead
		for _, secret := range secrets {
			rows = append(rows, model.SecretRead{SecretID: secret, Name: "N_" + secret, RunID: id, Folder: "infra", Job: "deploy", UserName: "ops"})
		}
		assert.NoError(t, reads.Record(ctx, rows))
	}
	run("r1", "s1", "s2")
	assert.Error(t, reads.Record(ctx, []model.SecretRead{{SecretID: "s1"}}))

	got, err := reads.ForRun(ctx, "r1")
	assert.NoError(t, err)
	assert.Len(t, got, 2)
	assert.Equal(t, "N_s1", got[0].Name)
	assert.NotNil(t, got[0].CreatedAt)

	for _, id := range []string{"r2", "r3", "r4"} {
		run(id, "s1")
	}
	counts, err := reads.Counts(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 3, counts["s1"])
	assert.Equal(t, 1, counts["s2"])
	last, err := reads.Last(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "r4", last["s1"].RunID)
	assert.Equal(t, "r1", last["s2"].RunID)

	page, err := reads.List(ctx, "s1", "", "", 2)
	assert.NoError(t, err)
	assert.Len(t, page, 2)
	assert.Equal(t, "r4", page[0].RunID)
	older, err := reads.List(ctx, "s1", page[1].ID, "", 2)
	assert.NoError(t, err)
	assert.Len(t, older, 1)
	assert.Equal(t, "r2", older[0].RunID)
	newer, err := reads.List(ctx, "s1", "", older[0].ID, 1)
	assert.NoError(t, err)
	assert.Equal(t, "r3", newer[0].RunID)

	rows, bytes, err := reads.Usage(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 4, rows)
	assert.True(t, bytes > 0)
	assert.NoError(t, reads.Prune(ctx, time.Now().Add(time.Hour)))
	rows, _, err = reads.Usage(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, rows)
}

// DeleteThrough removes the reads up to one, oldest first.
func TestSecretReadStorage_DeleteThrough(t *testing.T) {
	t.Parallel()
	reads := storage.NewSecretReadStorage(newDB(t))
	ctx := context.Background()
	batch := []model.SecretRead{{SecretID: "s1", RunID: "r1"}, {SecretID: "s1", RunID: "r2"}, {SecretID: "s2", RunID: "r3"}}
	assert.NoError(t, reads.Record(ctx, batch))
	assert.NoError(t, reads.DeleteThrough(ctx, batch[1].ID))
	left, err := reads.All(ctx)
	assert.NoError(t, err)
	assert.Len(t, left, 1)
	assert.Equal(t, batch[2].ID, left[0].ID)
}
