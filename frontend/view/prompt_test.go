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
