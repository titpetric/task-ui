package view_test

import (
	"testing"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
)

func TestCrumbs(t *testing.T) {
	t.Parallel()

	crumbs := view.Crumbs("Settings", "/settings", "Keys")
	assert.Len(t, crumbs, 2)
	assert.Equal(t, view.Crumb{Text: "Settings", Href: "/settings"}, crumbs[0])
	assert.Equal(t, view.Crumb{Text: "Keys"}, crumbs[1])
	assert.Nil(t, view.Crumbs())
}
