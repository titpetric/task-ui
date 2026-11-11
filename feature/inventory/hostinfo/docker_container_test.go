package hostinfo_test

import (
	"testing"

	"github.com/titpetric/task-ui/hostinfo"
	"github.com/titpetric/task-ui/internal/assert"
)

// docker ps lines parse, with the sizes in bytes.
func TestParseDockerPS(t *testing.T) {
	t.Parallel()
	out := `{"Names":"postgres","Image":"postgres:17","State":"running","Status":"Up 12 days","Size":"63.2MB (virtual 1.5GB)"}
not json
{"Names":"a,b","Image":"x","State":"exited","Status":"Exited (0)","Size":"0B (virtual 12.4kB)"}
{"Names":"odd","Size":"lots"}
`
	containers := hostinfo.ParseDockerPS(out)
	assert.Len(t, containers, 3)
	assert.Equal(t, hostinfo.DockerContainer{Name: "postgres", Image: "postgres:17", State: "running", Status: "Up 12 days", SizeRw: 63200000, SizeVirtual: 1500000000}, containers[0])
	assert.Equal(t, "a", containers[1].Name)
	assert.Equal(t, uint64(12400), containers[1].SizeVirtual)
	assert.Equal(t, uint64(0), containers[2].SizeRw)
}
