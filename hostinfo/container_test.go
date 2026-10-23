package hostinfo_test

import (
	"testing"

	"github.com/titpetric/task-ui/beszel"
	"github.com/titpetric/task-ui/hostinfo"
	"github.com/titpetric/task-ui/internal/assert"
)

// The two sources merge by name, the busiest first; a container one of
// them lacks says which it has.
func TestMerge(t *testing.T) {
	t.Parallel()
	measured := []beszel.Container{
		{Name: "web", CPU: 1, Memory: 100},
		{Name: "db", CPU: 5, Memory: 2048, Image: "postgres:17", Status: "Up 2 days"},
		{Name: "cache", CPU: 1, Memory: 400},
	}
	sized := []hostinfo.DockerContainer{
		{Name: "db", Image: "postgres:17.1", State: "running", Status: "Up 2 days (healthy)", SizeRw: 10, SizeVirtual: 100},
		{Name: "job", State: "exited"},
	}
	rows := hostinfo.Merge(measured, sized, 8)
	var names []string
	for _, row := range rows {
		names = append(names, row.Name)
	}
	assert.Equal(t, []string{"db", "cache", "web", "job"}, names)
	db := rows[0]
	assert.True(t, db.Measured && db.Sized)
	assert.Equal(t, 25.0, db.MemoryShare)
	assert.Equal(t, uint64(10), db.SizeRw)
	assert.Equal(t, "Up 2 days (healthy)", db.Status, "docker's word wins")
	assert.False(t, rows[1].Sized)
	assert.False(t, rows[3].Measured)
	assert.Equal(t, 0.0, hostinfo.Merge(measured, nil, 0)[0].MemoryShare)
	assert.Equal(t, hostinfo.Container{Name: "job", State: "exited", Sized: true}, rows[3])
}
