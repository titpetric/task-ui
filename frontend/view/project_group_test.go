package view_test

import (
	"testing"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// The Projects page groups the entries by workspace, the root first in
// the workspaces' order: a folder at the top goes to the root unless it
// is a workspace, a checkout inside a workspace to that workspace, and
// an empty workspace keeps a group without rows.
func TestProjectGroups(t *testing.T) {
	t.Parallel()

	entries := []folder.Entry{
		{ID: "acme"},
		{ID: "acme/api"},
		{ID: "acme/web"},
		{ID: "infra"},
		{ID: "labs"},
		{ID: "tools"},
	}
	workspaces := []model.Workspace{{Folder: model.Default}, {Folder: "acme"}, {Folder: "labs"}}
	groups := view.ProjectGroups(entries, workspaces)
	assert.Len(t, groups, 3)

	assert.Equal(t, model.Default, groups[0].Workspace)
	assert.True(t, groups[0].Default())
	assert.Equal(t, "Projects", groups[0].Title())
	assert.Equal(t, "in /srv/src", groups[0].Note("/srv"))
	assert.Equal(t, []string{"infra", "tools"}, ids(groups[0].Entries))

	assert.Equal(t, "acme", groups[1].Workspace)
	assert.Equal(t, "acme", groups[1].Title())
	assert.Equal(t, "inside acme/src/", groups[1].Note("/srv"))
	assert.Equal(t, []string{"acme/api", "acme/web"}, ids(groups[1].Entries))

	assert.Equal(t, "labs", groups[2].Workspace)
	assert.Len(t, groups[2].Entries, 0)
	assert.Equal(t, "No projects in labs yet.", groups[2].Empty())
	assert.Equal(t, "No projects yet: add one from a git repository, or put a folder with jobs under src.", groups[0].Empty())

	// Without the default workspace there is no root group, and an
	// entry at the top belongs to no group.
	groups = view.ProjectGroups(entries, workspaces[1:])
	assert.Len(t, groups, 2)
	assert.Equal(t, "acme", groups[0].Workspace)
	assert.Equal(t, []string{"acme/api", "acme/web"}, ids(groups[0].Entries))
	assert.Len(t, view.ProjectGroups(entries[3:], nil), 0)
}

// ids returns the entries' IDs in order.
func ids(entries []folder.Entry) []string {
	result := []string{}
	for _, entry := range entries {
		result = append(result, entry.ID)
	}
	return result
}
