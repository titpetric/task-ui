package hostinfo_test

import (
	"testing"

	"github.com/titpetric/task-ui/hostinfo"
	"github.com/titpetric/task-ui/internal/assert"
)

// The host's answer splits into its filesystems and its containers.
func TestParseFacts(t *testing.T) {
	t.Parallel()
	facts := hostinfo.ParseFacts(df + "@@docker\n" + `{"Names":"db","State":"running"}` + "\n@@exit 0\n")
	assert.Len(t, facts.Filesystems, 4)
	assert.Len(t, facts.Containers, 1)
	assert.True(t, facts.Docker)

	var empty *hostinfo.Facts
	assert.Nil(t, empty)
	none := hostinfo.ParseFacts(df + "@@docker\n@@exit 127\n")
	assert.False(t, none.Docker)
	assert.Len(t, none.Containers, 0)
}
