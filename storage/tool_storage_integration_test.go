//go:build integration

package storage_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// Tools store by name, and a start replaces the whole set.
func TestToolStorage(t *testing.T) {
	t.Parallel()
	s := storage.NewToolStorage(newDB(t))
	ctx := context.Background()
	tools, err := s.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, tools, 0)

	assert.NoError(t, s.Replace(ctx, []model.Tool{
		{Name: "node", Category: "runtime", Version: "v20.11.0", Path: "/usr/bin/node", Description: "JavaScript runtime"},
		{Name: "git", Category: "tooling", Version: "2.43.0", Path: "/usr/bin/git", Description: "version control"},
	}))
	tools, err = s.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, tools, 2)
	assert.Equal(t, "node", tools[0].Name, "the runtime comes before the tooling")
	assert.Equal(t, "2.43.0", tools[1].Version)
	assert.Equal(t, "/usr/bin/git", tools[1].Path)
	assert.Equal(t, "version control", tools[1].Description)
	assert.True(t, tools[1].UpdatedAt != nil, "a row is stamped")

	assert.NoError(t, s.Replace(ctx, []model.Tool{
		{Name: "jq", Category: "custom", Version: "jq-1.7.1", Path: "/usr/bin/jq", Description: "JSON processor"},
	}))
	tools, err = s.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, tools, 1)
	assert.Equal(t, "jq", tools[0].Name)
}
