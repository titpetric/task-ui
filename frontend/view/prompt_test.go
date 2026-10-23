package view_test

import (
	"testing"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
)

// The prompt box of a folder posts its ID; of a directory inside it, the
// directory too.
func TestPromptIn(t *testing.T) {
	t.Parallel()
	assert.Equal(t, &view.Prompt{API: "/api/prompts", Folder: "malta", Where: "malta/", Place: "malta"}, view.PromptIn("malta", ""))
	assert.Equal(t, &view.Prompt{API: "/api/prompts", Folder: "malta", Dir: "dockerfiles/task", Where: "malta/dockerfiles/task/", Place: "malta/dockerfiles/task"}, view.PromptIn("malta", "dockerfiles/task"))
}

// The prompt box of a run's page posts the run, and stands open.
func TestPromptAbout(t *testing.T) {
	t.Parallel()
	assert.Equal(t, &view.Prompt{API: "/api/prompts", Run: "01RUN", Where: "malta/dockerfiles/", Place: "malta/dockerfiles", Open: true}, view.PromptAbout("01RUN", "malta/dockerfiles"))
}

// The Workflows page's box asks in the area, in the root unless its
// form picks a folder, open.
func TestPromptOfArea(t *testing.T) {
	t.Parallel()
	p := view.PromptOfArea("workflows")
	assert.Equal(t, "workflows", p.Area)
	assert.Equal(t, "/api/prompts", p.API)
	assert.Equal(t, "_", p.Place)
	assert.True(t, p.Open)
}
