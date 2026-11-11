package skills_test

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/skills"
)

// IsSkill holds for an active skill, and for no other section.
func TestSection_IsSkill(t *testing.T) {
	t.Parallel()
	assert.True(t, (&skills.Section{Kind: skills.KindSkill, Active: true}).IsSkill())
	assert.False(t, (&skills.Section{Kind: skills.KindSkill}).IsSkill())
	assert.False(t, (&skills.Section{Kind: skills.KindPipeline, Active: true}).IsSkill())
}
