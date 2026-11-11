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
	assert.Equal(t, []string{"/", "/inventory", "/workflows", "/threads", "/refinement", "/artifacts", "/schedules", "/settings"},
		hrefs(rows), "the defaults are the sidebar as it stood before the table held it")
	assert.Equal(t, int64(0), rows[0].Position)
	assert.Equal(t, int64(7), rows[7].Position)
	assert.Equal(t, model.RuleSettings, rows[7].Rule)
	assert.Equal(t, model.SidebarView, rows[0].Kind)
	assert.False(t, rows[0].IsModule())

	// A second call changes nothing, whatever the table holds now.
	assert.NoError(t, s.Order(ctx, []string{"/threads", "/"}, map[string]bool{"/threads": true}))
	assert.NoError(t, s.EnsureDefaults(ctx, model.SidebarDefaults()))
	rows, err = s.List(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "/threads", rows[0].Href)
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

	// A row with no kind is a view: model.SidebarView is the zero value.
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
		[]string{"/threads", "/", "/inventory", "/workflows", "/artifacts", "/refinement", "/schedules", "/settings"},
		map[string]bool{"/inventory": true, "/schedules": true}))

	rows, err := s.List(ctx)
	assert.NoError(t, err)
	assert.Equal(t, []string{"/threads", "/", "/inventory", "/workflows", "/artifacts", "/refinement", "/schedules", "/settings"},
		hrefs(rows))

	hidden := map[string]bool{}
	for _, row := range rows {
		hidden[row.Href] = row.Hidden
	}
	assert.True(t, hidden["/inventory"])
	assert.True(t, hidden["/schedules"])
	assert.False(t, hidden["/threads"])

	// A link the table does not hold is skipped, and the row count stays.
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

// Every method returns the driver's error.
func TestSidebarStorage_Closed(t *testing.T) {
	t.Parallel()
	s := storage.NewSidebarStorage(closedDB(t))
	ctx := context.Background()
	item := &model.Sidebar{Href: "/runs", Title: "Runs"}

	_, err := s.List(ctx)
	assert.ErrorContains(t, err, errClosed)
	_, err = s.Get(ctx, "/runs")
	assert.ErrorContains(t, err, errClosed)
	assert.ErrorContains(t, s.Save(ctx, item), errClosed)
	assert.ErrorContains(t, s.EnsureDefaults(ctx, []model.Sidebar{*item}), errClosed)
	assert.ErrorContains(t, s.Order(ctx, []string{"/runs"}, nil), errClosed)
	_, err = s.Keep(ctx, nil)
	assert.ErrorContains(t, err, errClosed)
}

// Save upserts and then reads the row back, so the item is the row. The
// read is its own error: the upsert lands and the scan of what came back
// fails, which an EnsureDefaults of an empty table passes on.
func TestSidebarStorage_SaveReadBack(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	item := model.Sidebar{Href: "/runs", Title: "Runs"}

	s := storage.NewSidebarStorage(widenedDB(t, model.SidebarTable))
	row := item
	assert.ErrorContains(t, s.Save(ctx, &row), errNoField)

	empty := storage.NewSidebarStorage(newDB(t))
	defaults := []model.Sidebar{item}
	assert.NoError(t, empty.EnsureDefaults(ctx, defaults))

	refused := newDB(t)
	refuse(t, refused, model.SidebarTable, "INSERT")
	assert.ErrorContains(t, storage.NewSidebarStorage(refused).EnsureDefaults(ctx, []model.Sidebar{item}), errRefused)
}

// Order writes a row per link in one transaction, and returns the write's error.
func TestSidebarStorage_OrderBroken(t *testing.T) {
	t.Parallel()
	s := storage.NewSidebarStorage(brokenDB(t, model.SidebarTable))
	assert.ErrorContains(t, s.Order(context.Background(), []string{"/runs"}, nil), errNoTable)
}
