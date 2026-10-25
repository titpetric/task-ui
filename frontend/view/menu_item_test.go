package view_test

import (
	"html/template"
	"testing"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
)

// An item takes href, download, note and disabled as fields and the
// data-*, aria-* and title attributes escaped; any other is an error.
func TestNewMenuItem(t *testing.T) {
	t.Parallel()

	link, err := view.NewMenuItem("edit", "Configure", "href", "/settings/projects/web", "title", `say "hi"`)
	assert.NoError(t, err)
	assert.Equal(t, "/settings/projects/web", link.Href)
	assert.Equal(t, template.HTMLAttr(`title="say &#34;hi&#34;"`), link.Attrs)
	assert.False(t, link.Danger)

	btn, err := view.NewMenuItem("rotate", "Rotate key", "data-api", "/api/keys/k/rotate", "data-confirm", "Rotate <k>?", "disabled", "Delete the backup first")
	assert.NoError(t, err)
	assert.Equal(t, template.HTMLAttr(`data-api="/api/keys/k/rotate" data-confirm="Rotate &lt;k&gt;?"`), btn.Attrs)
	assert.True(t, btn.Disabled)
	assert.Equal(t, "Delete the backup first", btn.Note)

	tgz, err := view.NewMenuItem("download", "Download as .tgz", "href", "/a.tgz", "download", "")
	assert.NoError(t, err)
	assert.True(t, tgz.Download)

	_, err = view.NewMenuItem("x", "X", "onclick", "alert(1)")
	assert.Error(t, err)
	_, err = view.NewMenuItem("x", "X", "data-next")
	assert.Error(t, err)

	del, err := view.NewDangerItem("trash", "Delete", "data-delete", "/api/x")
	assert.NoError(t, err)
	assert.True(t, del.Danger)
}

// A primary item is the main button of a head, an item as any other.
func TestNewPrimaryItem(t *testing.T) {
	t.Parallel()

	add, err := view.NewPrimaryItem("plus", "Add project", "href", "/settings/projects/new")
	assert.NoError(t, err)
	assert.True(t, add.Primary)
	assert.False(t, add.Danger)
	assert.Equal(t, "/settings/projects/new", add.Href)

	_, err = view.NewPrimaryItem("plus", "Add", "onclick", "x")
	assert.Error(t, err)
}
