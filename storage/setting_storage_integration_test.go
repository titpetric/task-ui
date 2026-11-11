//go:build integration

package storage_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/storage"
)

// Settings store by name; a second Set replaces a value and keeps the
// others.
func TestSettingStorage(t *testing.T) {
	t.Parallel()
	s := storage.NewSettingStorage(newDB(t))
	ctx := context.Background()
	values, err := s.Values(ctx)
	assert.NoError(t, err)
	assert.Len(t, values, 0)
	assert.NoError(t, s.Set(ctx, map[string]string{"a": "1", "b": "2"}))
	assert.NoError(t, s.Set(ctx, map[string]string{"b": "3"}))
	values, err = s.Values(ctx)
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"a": "1", "b": "3"}, values)
}

// Every method returns the driver's error, from a closed pool and from a
// statement inside the transaction alike.
func TestSettingStorage_Errors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	closed := storage.NewSettingStorage(closedDB(t))
	_, err := closed.Values(ctx)
	assert.ErrorContains(t, err, errClosed)
	assert.ErrorContains(t, closed.Set(ctx, map[string]string{"a": "1"}), errClosed)

	gone := storage.NewSettingStorage(brokenDB(t, model.SettingTable))
	assert.ErrorContains(t, gone.Set(ctx, map[string]string{"a": "1"}), errNoTable)
}
