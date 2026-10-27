package view_test

import (
	"testing"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// Pin and Unpin: an item names what it pins, and carries the pin's ID
// once pinned, which a click removes.
func TestPage_PinItem(t *testing.T) {
	t.Parallel()
	p := &view.Page{Pins: []model.Pin{{ID: "p1", Folder: "malta/apps", Args: "[]"}, {ID: "p2", Folder: "malta/dockerfiles", Job: "build", Dir: "task", Args: "[]"}, {ID: "p3", Folder: "malta/dockerfiles", Job: "push", Args: `["-f"]`}}}
	pin := p.PinItem("malta/apps", "", "", "folder")
	assert.Equal(t, "Unpin folder", pin.Label)
	assert.Contains(t, string(pin.Attrs), `data-pin-id="p1"`)
	assert.Contains(t, string(pin.Attrs), `data-pin-key="malta/apps||"`)

	assert.Equal(t, "Unpin build", p.PinItem("malta/dockerfiles", "build", "task", "build").Label)
	pinItem := p.PinItem("malta/dockerfiles", "build", "", "build")
	assert.Equal(t, "Pin build", pinItem.Label)
	assert.NotContains(t, string(pinItem.Attrs), "data-pin-id")
	// A pin with args is the sidebar's; the job's own Pin stays.
	assert.Nil(t, p.PinOf("malta/dockerfiles", "push", ""))
	assert.Len(t, p.PinnedJobs("malta"), 0)
}

// A browse page pins the directory it shows, below its folder.
func TestPage_BrowseID(t *testing.T) {
	t.Parallel()
	p := &view.Page{Folder: &folder.Folder{ID: "malta"}, Browse: "apps/blog/"}
	assert.Equal(t, "malta/apps/blog", p.BrowseID())
}
