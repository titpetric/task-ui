package search_test

import (
	"testing"

	"github.com/titpetric/task-ui/feature/search"
	"github.com/titpetric/task-ui/internal/assert"
)

// Stats counts what a run did.
func TestStats(t *testing.T) {
	t.Parallel()
	stats := search.Stats{Documents: 3, Indexed: 1, Skipped: 2, Deleted: 0, Jobs: 4, JobsDeleted: 1}
	assert.Equal(t, 3, stats.Documents)
	assert.Equal(t, 1, stats.JobsDeleted)
}
