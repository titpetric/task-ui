package beszel_test

import (
	"context"
	"testing"
	"time"

	"github.com/titpetric/task-ui/beszel"
	"github.com/titpetric/task-ui/internal/assert"
)

// The inventory follows the settings, and is off without them.
func TestInventory(t *testing.T) {
	t.Parallel()
	server, lists := hub(t)
	url, token := "", ""
	inv := &beszel.Inventory{TTL: time.Hour, Settings: func(context.Context) (string, string, error) { return url, token, nil }}

	_, err := inv.Systems(context.Background())
	assert.ErrorIs(t, err, beszel.ErrOff)

	url, token = server.URL, "t1"
	systems, err := inv.Systems(context.Background())
	assert.NoError(t, err)
	assert.Len(t, systems, 2)
	_, _ = inv.Systems(context.Background())
	assert.Equal(t, int32(1), lists.Load(), "the same client, kept")

	token = "expired"
	_, err = inv.Systems(context.Background())
	assert.ErrorIs(t, err, beszel.ErrToken, "a new token is a new client")
}
