package skills_test

import (
	"encoding/json"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/skills"
)

// A command reads as atkins lists it, a job's own condition with it.
func TestCommand(t *testing.T) {
	t.Parallel()
	var c skills.Command
	assert.NoError(t, json.Unmarshal([]byte(`{"id":"compose:custom","name":"custom","cmd":"atkins compose:custom","interactive":false,"args":true,"exec":["docker","compose"],"active":true,"when":{"files":["compose.yml"],"here":true}}`), &c))
	assert.Equal(t, "compose:custom", c.ID)
	assert.Equal(t, "custom", c.Name)
	assert.True(t, c.Args)
	assert.Equal(t, []string{"docker", "compose"}, c.Exec)
	assert.True(t, c.When.Here)
}
