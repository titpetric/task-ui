package channel_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/internal/channel"
)

// The bridge needs an instance and a token.
func TestBridge_Usage(t *testing.T) {
	t.Parallel()
	err := (&channel.Bridge{}).Run(context.Background())
	assert.Error(t, err)
}
