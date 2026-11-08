package frontend

import (
	"testing"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// A pin resolves against the folders the reader sees: a folder of the
// catalogue, or a job of one. A folder nothing lists and a job no folder
// declares resolve to nothing, so a stale pin places no row.
func TestSidePin(t *testing.T) {
	t.Parallel()

	edge := &folder.Folder{ID: "infra/edge", Title: "edge", Jobs: []*folder.Job{{ID: "plan", Name: "plan"}}}
	p := &view.Page{Folders: []*folder.Folder{{ID: "infra", Title: "infra"}, edge}}

	side, ok := sidePin(p, model.Pin{Folder: "infra/edge"})
	assert.True(t, ok)
	assert.Equal(t, "infra/edge", side.Folder.ID)
	assert.Nil(t, side.Job)

	side, ok = sidePin(p, model.Pin{Folder: "infra/edge", Job: "plan"})
	assert.True(t, ok)
	assert.Equal(t, "plan", side.Job.ID)

	_, ok = sidePin(p, model.Pin{Folder: "nothing"})
	assert.False(t, ok)

	_, ok = sidePin(p, model.Pin{Folder: "infra/edge", Job: "apply"})
	assert.False(t, ok)
}

// The names of the sidebar's folders are what the tree lists, and a pin
// of a folder the defaults left out puts it and the folders above it in.
func TestSidebarPlacesAPin(t *testing.T) {
	t.Parallel()

	edge := &folder.Folder{ID: "malta/infra/edge", Title: "edge"}
	folders := view.NewSidebar(nil, []view.SidePin{{Pin: model.Pin{Folder: edge.ID}, Folder: edge}}, map[string]int{"malta": 1}, nil)

	var ids []string
	for _, f := range folders {
		ids = append(ids, f.ID)
	}
	assert.Equal(t, []string{"malta", "malta/infra", "malta/infra/edge"}, ids)
}
