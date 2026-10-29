//go:build integration

package api_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// links is the href of each row of a listing.
func links(rows []any) []string {
	out := make([]string, 0, len(rows))
	for _, row := range rows {
		if item, ok := row.(map[string]any); ok {
			out = append(out, item["href"].(string))
		}
	}
	return out
}

// The sidebar lists in its order, and a save writes the whole table at
// once: the order is the array's order, and the switches say what shows.
func TestSidebar(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	list := f.call(t, asAdmin, http.MethodGet, "/sidebar", nil)
	assert.Equal(t, http.StatusOK, list.Status)
	assert.Equal(t, []string{"/", "/inventory", "/workflows", "/runs", "/artifacts", "/schedules", "/notes", "/settings"},
		links(list.list("sidebar")))

	saved := f.call(t, asAdmin, http.MethodPut, "/sidebar", map[string]any{
		"items": []string{"/runs", "/", "/notes", "/inventory", "/workflows", "/artifacts", "/schedules", "/settings"},
		"shown": map[string]bool{"/inventory": false, "/schedules": false},
	})
	assert.Equal(t, http.StatusOK, saved.Status)
	assert.Equal(t, []string{"/runs", "/", "/notes", "/inventory", "/workflows", "/artifacts", "/schedules", "/settings"},
		links(saved.list("sidebar")), "the answer is the table as it stands")

	rows, err := f.opts.Sidebar.List(context.Background())
	assert.NoError(t, err)
	hidden := map[string]bool{}
	for _, row := range rows {
		hidden[row.Href] = row.Hidden
	}
	assert.True(t, hidden["/inventory"], "a switch turned off hides its row")
	assert.True(t, hidden["/schedules"])
	assert.False(t, hidden["/runs"])
	assert.Equal(t, int64(0), rows[0].Position)
	assert.Equal(t, "/runs", rows[0].Href)

	// A request that names one row changes one row: the order it posts
	// and the switch it names, and the rest keep what they had.
	again := f.call(t, asAdmin, http.MethodPut, "/sidebar", map[string]any{
		"items": []string{"/", "/runs", "/notes", "/inventory", "/workflows", "/artifacts", "/schedules", "/settings"},
		"shown": map[string]bool{"/inventory": true},
	})
	assert.Equal(t, http.StatusOK, again.Status)

	rows, err = f.opts.Sidebar.List(context.Background())
	assert.NoError(t, err)
	hidden = map[string]bool{}
	for _, row := range rows {
		hidden[row.Href] = row.Hidden
	}
	assert.False(t, hidden["/inventory"], "the switch the request named is the one that changed")
	assert.True(t, hidden["/schedules"], "a row the request did not name keeps its flag")
	assert.Equal(t, "/", rows[0].Href)
}

// Only an admin orders the sidebar, and an empty request is refused
// rather than taken as "hide everything".
func TestSidebarRefuses(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	for _, user := range []string{asOps, asViewer} {
		res := f.call(t, user, http.MethodPut, "/sidebar", map[string]any{"items": []string{"/"}})
		assert.Equal(t, http.StatusForbidden, res.Status, user)
	}

	res := f.call(t, asAdmin, http.MethodPut, "/sidebar", map[string]any{})
	assert.Equal(t, http.StatusBadRequest, res.Status)
	assert.Contains(t, res.str("error"), "in the order to list them")

	// Nothing moved.
	rows, err := f.opts.Sidebar.List(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "/", rows[0].Href)
}

// A link the table does not hold is left alone rather than added: the
// page lists what the table holds, and a link it did not list is a stale
// form or a changed instance.
func TestSidebarUnknown(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	saved := f.call(t, asAdmin, http.MethodPut, "/sidebar", map[string]any{
		"items": []string{"/gone", "/"},
		"shown": map[string]bool{"/gone": true, "/": true},
	})
	assert.Equal(t, http.StatusOK, saved.Status)
	assert.Len(t, saved.list("sidebar"), len(model.SidebarDefaults()))
	assert.NotContains(t, links(saved.list("sidebar")), "/gone")
}
