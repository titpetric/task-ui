package beszel_test

import (
	"testing"

	"github.com/titpetric/task-ui/beszel"
	"github.com/titpetric/task-ui/internal/assert"
)

func TestHosts(t *testing.T) {
	t.Parallel()
	systems := []beszel.System{
		{Name: "web", Host: "web.example", Status: "up"},
		{Name: "db", Host: "10.0.0.2", Status: "paused"},
		{Name: "box", Host: "box", Status: "up"},
	}
	logins := map[string]string{"web.example": "root@web.example", "10.0.0.2": "debian@10.0.0.2"}
	assert.Equal(t, "web\tweb.example\troot\tup\ndb\t10.0.0.2\tdebian\tdown\nbox\tbox\t\tup", beszel.Hosts(systems, logins))
	assert.Equal(t, "", beszel.Hosts(nil, nil))
}
