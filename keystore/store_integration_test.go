//go:build integration

package keystore_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/keystore"
)

// The key folder is a Store: another store keeps the same contract.
func TestStore(t *testing.T) {
	t.Parallel()
	var store keystore.Store = newStore(t)
	keys, err := store.List(context.Background())
	assert.NoError(t, err)
	assert.Len(t, keys, 0)
}
