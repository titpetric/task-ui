package skills_test

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/skills"
)

// String says the condition in a line, the matched file by its name.
func TestWhen_String(t *testing.T) {
	t.Parallel()
	w := &skills.When{Files: []string{"compose.yml", "compose.yaml"}, Here: true, Contains: "build:", Matched: "/w/app/compose.yml"}
	assert.Equal(t, "compose.yml or compose.yaml here containing build:, matched compose.yml", w.String())
	assert.Equal(t, ".git", (&skills.When{Files: []string{".git"}}).String())
	assert.Equal(t, "", (*skills.When)(nil).String())
	assert.Equal(t, "", (&skills.When{}).String())
}
