package beszel_test

import (
	"testing"

	"github.com/titpetric/task-ui/beszel"
	"github.com/titpetric/task-ui/internal/assert"
)

// A sample is a record's time and measures.
func TestSample(t *testing.T) {
	t.Parallel()
	s := beszel.Sample{Stats: beszel.Stats{CPU: 5}}
	assert.Equal(t, 5.0, s.Stats.CPU)
	cs := beszel.ContainerSample{Containers: []beszel.Container{{Name: "db"}}}
	assert.Equal(t, "db", cs.Containers[0].Name)
	assert.Equal(t, 0.0, beszel.FsStats{}.Total)
	assert.Equal(t, beszel.Pair{1, 2}, beszel.Pair{1, 2})
}
