//go:build integration

package storage_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/internal/assert"
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

// Save stores one tool over its own row, and Remove drops one: the
// instance's tool editor works a tool at a time, where a start replaces
// the whole set.
func TestToolStorage_SaveAndRemove(t *testing.T) {
	t.Parallel()
	s := storage.NewToolStorage(newDB(t))
	ctx := context.Background()

	assert.NoError(t, s.Save(ctx, model.Tool{Name: "jq", Category: "custom", Version: "jq-1.7", Path: "/usr/bin/jq"}))
	assert.NoError(t, s.Save(ctx, model.Tool{Name: "yq", Category: "custom", Version: "v4.44", Path: "/usr/bin/yq"}))
	tools, err := s.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, tools, 2)

	assert.NoError(t, s.Save(ctx, model.Tool{Name: "jq", Category: "custom", Version: "jq-1.8", Path: "/usr/local/bin/jq"}))
	tools, err = s.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, tools, 2, "a second save of a name replaces its row")
	assert.Equal(t, "jq-1.8", tools[0].Version)
	assert.Equal(t, "/usr/local/bin/jq", tools[0].Path)

	assert.NoError(t, s.Remove(ctx, "jq"))
	assert.NoError(t, s.Remove(ctx, "jq"), "removing what is gone is not an error")
	tools, err = s.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, tools, 1)
	assert.Equal(t, "yq", tools[0].Name)
}

// Every method returns the driver's error.
func TestToolStorage_Closed(t *testing.T) {
	t.Parallel()
	s := storage.NewToolStorage(closedDB(t))
	ctx := context.Background()

	_, err := s.List(ctx)
	assert.ErrorContains(t, err, errClosed)
	assert.ErrorContains(t, s.Save(ctx, model.Tool{Name: "jq"}), errClosed)
	assert.ErrorContains(t, s.Remove(ctx, "jq"), errClosed)
	assert.ErrorContains(t, s.Replace(ctx, []model.Tool{{Name: "jq"}}), errClosed)
}

// The statements inside a transaction hand their error back too: the
// delete against a table that is gone, and the insert of a name the
// same call already inserted, which the primary key refuses.
func TestToolStorage_Broken(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	gone := storage.NewToolStorage(brokenDB(t, model.ToolTable))
	assert.ErrorContains(t, gone.Save(ctx, model.Tool{Name: "jq"}), errNoTable)
	assert.ErrorContains(t, gone.Replace(ctx, []model.Tool{{Name: "jq"}}), errNoTable)

	s := storage.NewToolStorage(newDB(t))
	twice := []model.Tool{{Name: "jq", Category: "custom"}, {Name: "jq", Category: "custom"}}
	assert.ErrorContains(t, s.Replace(ctx, twice), "UNIQUE constraint failed")
	tools, err := s.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, tools, 0, "the transaction rolled back")
}
