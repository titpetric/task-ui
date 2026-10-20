package view_test

import (
	"testing"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
)

// A box a folder, checked for those seen; a seen folder the catalogue
// lacks keeps its box.
func TestFolderChoice(t *testing.T) {
	t.Parallel()
	var got []view.FolderChoice = view.FolderChoices([]*folder.Folder{{ID: "infra"}, {ID: "web"}}, []string{"web", "old"})
	assert.Equal(t, []view.FolderChoice{{ID: "infra"}, {ID: "web", Checked: true}, {ID: "old", Checked: true}}, got)
}
