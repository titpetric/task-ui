package view_test

import (
	"testing"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/skills"
)

// A row of the alert holds the finding and its folder.
func TestPipelineError(t *testing.T) {
	t.Parallel()
	f := &folder.Folder{ID: "app", Errors: []skills.Finding{{Path: "atkins.yml", Detail: "d"}}}
	p := &view.Page{Folders: []*folder.Folder{f}}
	row := p.PipelineAlert("app").Shown[0]
	assert.Equal(t, view.PipelineError{Folder: f, Finding: f.Errors[0]}, row)
}
