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

	groups := view.Groups(folders)
	assert.Len(t, groups, 4)
	assert.Equal(t, "apps", groups[0].Top.ID)
	assert.Equal(t, []string{"apps/blog", "apps/lib/web"}, idsOf(groups[0].Nested))
	assert.Equal(t, "apps-x", groups[1].Top.ID)
	assert.Len(t, groups[1].Nested, 0)
	assert.Equal(t, "infra", groups[2].Top.ID)
	assert.Equal(t, []string{"infra/edge"}, idsOf(groups[2].Nested))
	assert.True(t, groups[3].Top == nil, "private is not listed")
	assert.Equal(t, []string{"private/deep"}, idsOf(groups[3].Nested))
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
	var g view.Group = view.Groups([]*folder.Folder{{ID: "deploy"}, {ID: "deploy/terraform"}})[0]
	assert.Equal(t, "deploy", g.Top.ID)
	assert.Equal(t, []string{"deploy/terraform"}, idsOf(g.Nested))
}
