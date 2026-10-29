package view_test

import (
	"testing"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
)

// The sidebar groups the folders by project: a top folder with the
// folders inside it, in their order. A folder whose top is not listed
// stands in a group of its own, without a top.
func TestGroups(t *testing.T) {
	t.Parallel()
	ids := []string{"apps", "apps-x", "apps/blog", "apps/lib/web", "infra", "infra/edge", "private/deep"}
	folders := make([]*folder.Folder, 0, len(ids))
	for _, id := range ids {
		folders = append(folders, &folder.Folder{ID: id})
	}

	groups := view.Groups(folders, nil)
	assert.Len(t, groups, 4)
	assert.Equal(t, "apps", groups[0].Top.ID)
	assert.Equal(t, []string{"apps/blog", "apps/lib/web"}, idsOf(groups[0].Nested))
	assert.Equal(t, "apps-x", groups[1].Top.ID)
	assert.Len(t, groups[1].Nested, 0)
	assert.Equal(t, "infra", groups[2].Top.ID)
	assert.Equal(t, []string{"infra/edge"}, idsOf(groups[2].Nested))
	assert.True(t, groups[3].Top == nil, "private is not listed")
	assert.Equal(t, "private", groups[3].ID)
	assert.Equal(t, []string{"private/deep"}, idsOf(groups[3].Nested))
	for _, g := range groups {
		assert.False(t, g.Workspace, g.ID)
		assert.Len(t, g.Projects, 0)
	}
}

// A workspace holds its projects, ws/project, each a group with the
// folders inside it; a folder whose project is not listed stands in a
// project of its own, without a top. The projects at the root come
// first, the workspaces after them by rank.
func TestGroups_workspace(t *testing.T) {
	t.Parallel()
	ids := []string{"apps", "apps/blog", "malta", "malta/deploy", "malta/deploy/web", "malta/deploy/web/edge", "malta/pkg", "malta/vhosts/site", "lab/deep"}
	folders := make([]*folder.Folder, 0, len(ids))
	for _, id := range ids {
		folders = append(folders, &folder.Folder{ID: id})
	}

	groups := view.Groups(folders, map[string]int{"malta": 0, "lab": 1})
	assert.Len(t, groups, 3)
	assert.Equal(t, "apps", groups[0].Top.ID)
	assert.False(t, groups[0].Workspace)
	assert.Equal(t, []string{"apps/blog"}, idsOf(groups[0].Nested))

	malta := groups[1]
	assert.True(t, malta.Workspace)
	assert.Equal(t, "malta", malta.Top.ID)
	assert.Len(t, malta.Nested, 0)
	assert.Len(t, malta.Projects, 3)
	assert.Equal(t, "malta/deploy", malta.Projects[0].Top.ID)
	assert.Equal(t, []string{"malta/deploy/web", "malta/deploy/web/edge"}, idsOf(malta.Projects[0].Nested))
	assert.Equal(t, "malta/pkg", malta.Projects[1].Top.ID)
	assert.Len(t, malta.Projects[1].Nested, 0)
	assert.True(t, malta.Projects[2].Top == nil, "malta/vhosts is not listed")
	assert.Equal(t, "malta/vhosts", malta.Projects[2].ID)
	assert.Equal(t, []string{"malta/vhosts/site"}, idsOf(malta.Projects[2].Nested))
	for _, p := range malta.Projects {
		assert.False(t, p.Workspace, p.ID)
	}

	// A workspace whose own folder is not listed still holds its projects.
	lab := groups[2]
	assert.True(t, lab.Workspace)
	assert.True(t, lab.Top == nil, "lab is not listed")
	assert.Len(t, lab.Projects, 1)
	assert.Equal(t, "lab/deep", lab.Projects[0].Top.ID)
}

func idsOf(folders []*folder.Folder) []string {
	ids := []string{}
	for _, f := range folders {
		ids = append(ids, f.ID)
	}
	return ids
}

// A group is a project: its top folder, and the folders inside it.
func TestGroup(t *testing.T) {
	t.Parallel()
	var g view.Group = view.Groups([]*folder.Folder{{ID: "deploy"}, {ID: "deploy/terraform"}}, nil)[0]
	assert.Equal(t, "deploy", g.ID)
	assert.Equal(t, "deploy", g.Top.ID)
	assert.Equal(t, []string{"deploy/terraform"}, idsOf(g.Nested))
}

// The top-level projects come first, in their order, and the workspaces
// after them: a project cloned at the root lists above every workspace,
// with any name, and none of it falls under a workspace's fold.
func TestGroups_newProject(t *testing.T) {
	t.Parallel()
	ids := []string{"acme", "acme/api", "alpha", "beta", "beta/app", "infra", "zeta"}
	folders := make([]*folder.Folder, 0, len(ids))
	for _, id := range ids {
		folders = append(folders, &folder.Folder{ID: id})
	}

	groups := view.Groups(folders, map[string]int{"acme": 0, "beta": 1})
	order := []string{}
	for _, g := range groups {
		order = append(order, g.ID)
	}
	assert.Equal(t, []string{"alpha", "infra", "zeta", "acme", "beta"}, order)
}

// The workspaces go by rank, then by folder; a project at the root
// whose top is not listed keeps its place among the others, and a user
// who sees a workspace's project but not the workspace finds it under
// the workspace, in the workspace's place.
func TestGroups_order(t *testing.T) {
	t.Parallel()
	ids := []string{"acme/api", "apps/blog", "beta", "beta/app", "cobalt", "infra", "zeta"}
	folders := make([]*folder.Folder, 0, len(ids))
	for _, id := range ids {
		folders = append(folders, &folder.Folder{ID: id})
	}

	groups := view.Groups(folders, map[string]int{"acme": 1, "beta": 1, "zeta": 0})
	order := []string{}
	for _, g := range groups {
		order = append(order, g.ID)
	}
	assert.Equal(t, []string{"apps", "cobalt", "infra", "zeta", "acme", "beta"}, order)
	assert.True(t, groups[0].Top == nil, "apps is not listed")
	assert.Equal(t, []string{"apps/blog"}, idsOf(groups[0].Nested))
	acme := groups[4]
	assert.True(t, acme.Workspace)
	assert.True(t, acme.Top == nil, "acme is not listed")
	assert.Len(t, acme.Projects, 1)
	assert.Equal(t, "acme/api", acme.Projects[0].Top.ID)
}
