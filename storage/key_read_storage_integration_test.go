//go:build integration

package storage_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// The log of key reads counts each key's reads, names the last, and
// drops the oldest for the retention rules.
func TestKeyReadStorage(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	reads := storage.NewKeyReadStorage(newDB(t))
	assert.Error(t, reads.Record(ctx, &model.KeyRead{}))
	first := &model.KeyRead{KeyName: "deploy_rsa4096", UserName: "claude", TokenID: "t1", Via: "mcp"}
	assert.NoError(t, reads.Record(ctx, first))
	assert.NoError(t, reads.Record(ctx, &model.KeyRead{KeyName: "deploy_rsa4096", UserName: "black"}))
	counts, err := reads.Counts(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 2, counts["deploy_rsa4096"])
	last, err := reads.Last(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "black", last["deploy_rsa4096"].UserName)
	assert.Equal(t, "api", last["deploy_rsa4096"].Via)
	all, err := reads.All(ctx)
	assert.NoError(t, err)
	assert.Len(t, all, 2)
	assert.NoError(t, reads.DeleteThrough(ctx, first.ID))
	all, _ = reads.All(ctx)
	assert.Len(t, all, 1)
}
