package api_test

import (
	"testing"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/internal/assert"
)

func TestCatalogue(t *testing.T) {
	t.Parallel()

	var c api.Catalogue = catalogue{
		"b": &folder.Folder{ID: "b"},
		"a": &folder.Folder{ID: "a"},
	}
	list := c.List()
	assert.Len(t, list, 2)
	assert.Equal(t, "a", list[0].ID)
	assert.Equal(t, "b", list[1].ID)
	assert.Equal(t, "b", c.Find("b").ID)
	assert.Nil(t, c.Find("c"))
}
