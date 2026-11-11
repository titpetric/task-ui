package retention_test

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/retention"
)

// A usage says its count and size.
func TestUsage_Says(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "1 run, 2.0 KB", retention.Usage{Count: 1, Bytes: 2048}.Says("run", "runs"))
	assert.Equal(t, "3 runs, 10 B", retention.Usage{Count: 3, Bytes: 10}.Says("run", "runs"))
}

func TestUsage_Name(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "1 read", retention.Usage{Count: 1}.Name("read", "reads"))
	assert.Equal(t, "0 reads", retention.Usage{}.Name("read", "reads"))
}
