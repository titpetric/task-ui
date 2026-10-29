package view_test

import (
	"testing"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// A row reads as its title, and a module's row takes the description the
// modinfo.php registered.
func TestSidebarRows(t *testing.T) {
	t.Parallel()
	items := []model.Sidebar{
		{Kind: model.SidebarView, Title: "Dashboard", Href: "/", Icon: "grid"},
		{Kind: model.SidebarModule, Title: "Refinement", Href: "/m/titpetric/refinement", Icon: "flow"},
		{Kind: model.SidebarModule, Title: "Gone", Href: "/m/acme/gone", Icon: "box"},
		{Kind: model.SidebarView, Href: "/custom"},
	}
	modules := []model.Module{{Name: "titpetric/refinement", Title: "Refinement", Description: "Refine a design"}}

	rows := view.SidebarRows(items, modules)
	assert.Len(t, rows, 4)

	assert.Equal(t, "Dashboard", rows[0].Title())
	assert.Equal(t, "", rows[0].Note(), "a view of the instance needs no line under it")
	assert.Nil(t, rows[0].Module)

	assert.Equal(t, "Refinement", rows[1].Title())
	assert.Equal(t, "Refine a design", rows[1].Note())
	assert.NotNil(t, rows[1].Module)

	assert.Equal(t, "a module", rows[2].Note(), "a row whose module is gone still reads as one")
	assert.Nil(t, rows[2].Module)

	assert.Equal(t, "/custom", rows[3].Title(), "a row with no title reads as its link")
}

// The Sidebar settings page lists every row, the hidden ones too, with a
// grip to drag it and a switch that leaves it out.
func TestSidebarPage(t *testing.T) {
	t.Parallel()
	items := model.SidebarDefaults()
	items[1].Hidden = true
	items = append(items, model.Sidebar{Kind: model.SidebarModule, Title: "Refinement", Href: "/m/titpetric/refinement", Icon: "flow"})

	p := admin(t)
	p.Path = "/settings/sidebar"
	p.Step = view.StepSidebar
	p.SidebarRows = view.SidebarRows(items, []model.Module{{Name: "titpetric/refinement", Description: "Refine a design"}})
	body := render(t, "sidebar", p)

	assert.Contains(t, body, `data-sort-rows`)
	assert.Contains(t, body, `data-row data-href="/m/titpetric/refinement"`)
	assert.Contains(t, body, `data-row-grip`)
	assert.Contains(t, body, `data-row-shown="false"`, "the hidden row loads with its switch off")
	assert.Contains(t, body, `data-row-shown="true"`)
	assert.Contains(t, body, `data-row-move="up"`)
	assert.Contains(t, body, `data-sidebar-changed`)
	assert.Contains(t, body, "Refine a design")

	// The switch carries no name: serialize splits a field name on a dot
	// into a map, and a module's link can hold one.
	assert.NotContains(t, body, `name="/m/titpetric/refinement"`)

	// A viewer reads the table and saves nothing.
	signIn(p, &model.User{Name: "guest", Role: model.RoleViewer}, nil)
	body = render(t, "sidebar", p)
	assert.NotContains(t, body, `data-sort-rows`)
	assert.NotContains(t, body, "data-row-grip")
	assert.Contains(t, body, "disabled")
}
