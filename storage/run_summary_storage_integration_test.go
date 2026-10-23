//go:build integration

package storage_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

func TestRunSummaryStorage(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	summaries := storage.NewRunSummaryStorage(newDB(t))
	_, err := summaries.Get(ctx, "r1")
	assert.ErrorIs(t, err, storage.ErrNotFound)

	assert.NoError(t, summaries.Set(ctx, &model.RunSummary{RunID: "r1", Summary: "ok", UserName: "admin"}))
	assert.NoError(t, summaries.Set(ctx, &model.RunSummary{RunID: "r1", Summary: "again", UserName: "admin"}))
	got, err := summaries.Get(ctx, "r1")
	assert.NoError(t, err)
	assert.Equal(t, "again", got.Summary)
	assert.NotNil(t, got.CreatedAt)
	assert.Error(t, summaries.Set(ctx, &model.RunSummary{RunID: "r2"}))

	assert.NoError(t, summaries.Delete(ctx, "r1"))
	_, err = summaries.Get(ctx, "r1")
	assert.ErrorIs(t, err, storage.ErrNotFound)
}
