package view_test

import (
	"testing"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/tooling"
)

// TestToolingSegment covers one command's share of a day.
func TestToolingSegment(t *testing.T) {
	t.Parallel()

	segment := view.NewTooling(&tooling.Report{Project: commandScope()}, toolingNow).Scopes[0].Windows[1].Bars[6].Segments[1]
	assert.Equal(t, "docker build", segment.Name)
	assert.Equal(t, "s1", segment.Color)
	assert.Contains(t, segment.Title, "docker build: 6.0s")

	assert.Equal(t, "", string(view.ToolingSegment{}.Style))
}
