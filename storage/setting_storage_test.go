package storage_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
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
