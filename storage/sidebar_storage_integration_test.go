//go:build integration

package storage_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// hrefs is the links of a row list, in the order it came back.
func hrefs(rows []model.Sidebar) []string {
	out := make([]string, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.Href)
	}
	return out
}

// An empty table takes the defaults, in their order, and a table with a
// row in it is left alone: an order the user saved is never overwritten.
func TestSidebarStorage_EnsureDefaults(t *testing.T) {
	t.Parallel()
	s := storage.NewSidebarStorage(newDB(t))
	ctx := context.Background()

	rows, err := s.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, rows, 0)

	assert.NoError(t, s.EnsureDefaults(ctx, model.SidebarDefaults()))
	rows, err = s.List(ctx)
	assert.NoError(t, err)
	assert.Equal(t, []string{"/", "/inventory", "/workflows", "/runs", "/artifacts", "/schedules", "/notes", "/settings"},
		hrefs(rows), "the defaults are the sidebar as it stood before the table held it")
	assert.Equal(t, int64(0), rows[0].Position)
	assert.Equal(t, int64(7), rows[7].Position)
	assert.Equal(t, model.RuleSettings, rows[7].Rule)
	assert.Equal(t, model.SidebarView, rows[0].Kind)
	assert.False(t, rows[0].IsModule())

	// A second call changes nothing, whatever the table holds now.
	assert.NoError(t, s.Order(ctx, []string{"/notes", "/"}, map[string]bool{"/notes": true}))
	assert.NoError(t, s.EnsureDefaults(ctx, model.SidebarDefaults()))
	rows, err = s.List(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "/notes", rows[0].Href)
	assert.True(t, rows[0].Hidden)
}

// Save keeps the position and the hidden flag of a link already there, so
// a module registering on every start does not walk to the end of the
// sidebar or come back shown after the user hid it.
func TestSidebarStorage_Save(t *testing.T) {
	t.Parallel()
	s := storage.NewSidebarStorage(newDB(t))
	ctx := context.Background()

	row := &model.Sidebar{Kind: model.SidebarModule, Href: "/m/titpetric/refinement", Title: "Refinement", Icon: "flow", Position: 1000}
	assert.NoError(t, s.Save(ctx, row))
	assert.NotEqual(t, "", row.ID)
	assert.Equal(t, int64(1000), row.Position)

	assert.NoError(t, s.Order(ctx, []string{"/m/titpetric/refinement"}, map[string]bool{"/m/titpetric/refinement": true}))

	again := &model.Sidebar{Kind: model.SidebarModule, Href: "/m/titpetric/refinement", Title: "Refinement, renamed", Icon: "box", Position: 1000}
	assert.NoError(t, s.Save(ctx, again))
	assert.Equal(t, row.ID, again.ID, "the row of a link stays the row")
	assert.Equal(t, int64(0), again.Position, "the order the user saved stays")
	assert.True(t, again.Hidden, "a row the user hid stays hidden")
	assert.Equal(t, "Refinement, renamed", again.Title, "the title is the modinfo.php's")
	assert.Equal(t, "box", again.Icon)

	// A row with no kind is a view, which is what the zero value means.
	plain := &model.Sidebar{Href: "/custom", Title: "Custom"}
	assert.NoError(t, s.Save(ctx, plain))
	assert.Equal(t, model.SidebarView, plain.Kind)
}

// Order writes the whole table at once: the position is the index, and
// hidden names the links to leave out.
func TestSidebarStorage_Order(t *testing.T) {
	t.Parallel()
	s := storage.NewSidebarStorage(newDB(t))
	ctx := context.Background()
	assert.NoError(t, s.EnsureDefaults(ctx, model.SidebarDefaults()))

	assert.NoError(t, s.Order(ctx,
		[]string{"/runs", "/", "/notes", "/inventory", "/workflows", "/artifacts", "/schedules", "/settings"},
		map[string]bool{"/inventory": true, "/schedules": true}))

	rows, err := s.List(ctx)
	assert.NoError(t, err)
	assert.Equal(t, []string{"/runs", "/", "/notes", "/inventory", "/workflows", "/artifacts", "/schedules", "/settings"},
		hrefs(rows))

	hidden := map[string]bool{}
	for _, row := range rows {
		hidden[row.Href] = row.Hidden
	}
	assert.True(t, hidden["/inventory"])
	assert.True(t, hidden["/schedules"])
	assert.False(t, hidden["/runs"])

	// A link the table does not hold is left alone rather than added.
	assert.NoError(t, s.Order(ctx, []string{"/gone"}, nil))
	rows, err = s.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, rows, 8)
}

// Keep drops the module rows a scan did not find and never a view's: the
// instance owns those.
func TestSidebarStorage_Keep(t *testing.T) {
	t.Parallel()
	s := storage.NewSidebarStorage(newDB(t))
	ctx := context.Background()
	assert.NoError(t, s.EnsureDefaults(ctx, model.SidebarDefaults()))

	for _, href := range []string{"/m/a/one", "/m/a/two"} {
		assert.NoError(t, s.Save(ctx, &model.Sidebar{Kind: model.SidebarModule, Href: href, Title: href}))
	}

	dropped, err := s.Keep(ctx, []string{"/m/a/one"})
	assert.NoError(t, err)
	assert.Equal(t, 1, dropped)

	rows, err := s.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, rows, 9, "the eight views stay and one module went")

	// A scan with no module at all drops every module row and no view.
	dropped, err = s.Keep(ctx, nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, dropped)

	rows, err = s.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, rows, 8)

	_, err = s.Get(ctx, "/settings")
	assert.NoError(t, err)
	_, err = s.Get(ctx, "/m/a/one")
	assert.ErrorIs(t, err, storage.ErrNotFound)
}
