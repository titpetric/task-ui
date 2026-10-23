package access_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/internal/assert"
)

// A request carries the ID of its token and how it reached the API: an
// MCP tool's call, or a client's.
func TestTokenContext(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	assert.Equal(t, "", access.TokenFrom(ctx))
	assert.Equal(t, access.ViaAPI, access.Via(ctx))
	ctx = access.WithVia(access.WithToken(ctx, "t1"), access.ViaMCP)
	assert.Equal(t, "t1", access.TokenFrom(ctx))
	assert.Equal(t, access.ViaMCP, access.Via(ctx))
}
