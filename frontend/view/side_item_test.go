package view_test

import (
	"testing"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/permission"
)

// titles is the title of each row that came back.
func titles(items []view.SideItem) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, item.Title)
	}
	return out
}

// An admin sees every default row, in the table's order, and the two
// rows that carry more than a link say so.
func TestSideItems(t *testing.T) {
	t.Parallel()
	p := &view.Page{Path: "/"}
	signIn(p, &model.User{Name: "root", Role: model.RoleAdmin}, nil)
	p.FillSide(model.SidebarDefaults())

	assert.Equal(t, []string{"Dashboard", "Inventory", "Workflows", "Runs", "Artifacts", "Scheduled", "Notes", "Settings"},
		titles(p.SideItems))
	assert.True(t, p.SideItems[0].Active, "the dashboard is active on its own path")
	assert.True(t, p.SideItems[2].Flows, "the workflows are listed under their own row")
	assert.True(t, p.SideItems[3].Pulse, "Runs carries the dot that pulses")
	assert.False(t, p.SideItems[0].Flows)
	assert.Equal(t, "grid", p.SideItems[0].Icon)
}

// A row is active on anything under its link, and the dashboard only on
// its own, so one page does not light every row.
func TestSideItemsActive(t *testing.T) {
	t.Parallel()
	p := &view.Page{Path: "/settings/keys"}
	signIn(p, &model.User{Name: "root", Role: model.RoleAdmin}, nil)
	p.FillSide(model.SidebarDefaults())

	for _, item := range p.SideItems {
		assert.Equal(t, item.Href == "/settings", item.Active, item.Href)
	}
	assert.False(t, p.SideActive("/"), "the dashboard is not active under /settings")
	assert.True(t, p.SideActive("/settings"))
}

// A row the table hides is out, and a row whose rule the user lacks is
// out whether it is hidden or not.
func TestSideItemsHidden(t *testing.T) {
	t.Parallel()
	rows := model.SidebarDefaults()
	rows[1].Hidden = true
	rows = append(rows, model.Sidebar{Kind: model.SidebarModule, Title: "Refinement", Href: "/m/titpetric/refinement", Icon: "flow", Rule: "module"})

	p := &view.Page{Path: "/"}
	signIn(p, &model.User{Name: "root", Role: model.RoleAdmin}, nil)
	p.FillSide(rows)
	assert.Equal(t, []string{"Dashboard", "Workflows", "Runs", "Artifacts", "Scheduled", "Notes", "Settings", "Refinement"},
		titles(p.SideItems), "the hidden row is out and the module is in")

	// A viewer with no rule sees neither the settings nor a module.
	signIn(p, &model.User{Name: "guest", Role: model.RoleViewer}, nil)
	p.FillSide(rows)
	assert.Equal(t, []string{"Dashboard", "Workflows", "Runs", "Artifacts", "Scheduled", "Notes"}, titles(p.SideItems))

	// The module rule alone brings the module back.
	signIn(p, &model.User{Name: "ops", Role: model.RoleViewer}, permission.Policy{"module": permission.Allow})
	p.FillSide(rows)
	assert.Contains(t, titles(p.SideItems), "Refinement")
}

// An icon from the database reaches the markup as a sprite name and
// nothing else, and one with nothing left falls back.
func TestSideItemsIcon(t *testing.T) {
	t.Parallel()
	p := &view.Page{Path: "/"}
	signIn(p, &model.User{Name: "root", Role: model.RoleAdmin}, nil)

	for icon, want := range map[string]string{
		"flow":          "flow",
		"FLOW":          "flow",
		" folder-open ": "folder-open",
		`"/><script>`:   "script",
		"#i-grid":       "i-grid",
		"":              "box",
		"!!!":           "box",
		"i-1":           "i-1",
	} {
		p.FillSide([]model.Sidebar{{Title: "One", Href: "/one", Icon: icon}})
		assert.Equal(t, want, p.SideItems[0].Icon, icon)
	}
}
