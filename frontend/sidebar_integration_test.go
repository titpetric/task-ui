//go:build integration

package frontend_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/permission"
)

// hrefsOf is the links of a row list, in the order it came back.
func hrefsOf(rows []model.Sidebar) []string {
	out := make([]string, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.Href)
	}
	return out
}

// The page is the sidebar as a table: every row, the hidden ones too,
// each with a grip and a switch; an admin saves, a user who may see the
// settings reads it, and a viewer may not open it.
func TestSidebarSettings(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()

	body := f.page(t, asAdmin, "/settings/sidebar")
	assert.Contains(t, body, "<h2>Sidebar</h2>")
	assert.Contains(t, body, "<thead><tr><th class=\"num\">Order</th><th>Item</th><th>Link</th><th class=\"num\">Shown</th><th></th></tr></thead>")
	assert.Contains(t, body, `data-api="/api/sidebar"`)
	assert.Contains(t, body, `data-sort-rows`)

	// Every default row has one, in the sidebar's order.
	last := -1
	for _, row := range model.SidebarDefaults() {
		at := strings.Index(body, `data-row data-href="`+row.Href+`"`)
		assert.True(t, at > last, row.Href+" after the row before it")
		last = at
	}

	// A hidden row loads with its switch off and still lists.
	rows, err := f.opts.Sidebar.List(ctx)
	assert.NoError(t, err)
	assert.NoError(t, f.opts.Sidebar.Order(ctx, hrefsOf(rows), map[string]bool{"/inventory": true}))
	body = f.page(t, asAdmin, "/settings/sidebar")
	assert.Contains(t, body, `data-row data-href="/inventory"`)
	assert.Contains(t, body, `data-row-shown="false"`)

	// A user who may see a step of the settings reads the page and saves
	// nothing; a viewer with nothing to open may not reach it.
	assert.NoError(t, f.opts.Users.Policies().Put(ctx, permission.KindUser, asOps, "secret/view", permission.Allow))
	body = f.page(t, asOps, "/settings/sidebar")
	assert.NotContains(t, body, `data-api="/api/sidebar"`)
	assert.Contains(t, body, "An admin saves the order.")

	assert.Equal(t, http.StatusForbidden, f.get(t, asViewer, "/settings/sidebar").Status)
}

// The step is in the settings trail, last, and the sidebar's own row for
// it is the one the page edits.
func TestSidebarSettingsStep(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	body := f.page(t, asAdmin, "/settings")
	assert.Contains(t, body, `href="/settings/sidebar"`)
	assert.Contains(t, body, "The order of the sidebar, and what it leaves out.")

	body = f.page(t, asAdmin, "/settings/sidebar")
	assert.Contains(t, body, `href="/settings/projects"`, "the trail links the step before it")
}
