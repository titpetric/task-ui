package view_test

import (
	"testing"
	"time"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
)

// A workspace with a limit keeps its newest projects, by their last
// commit, a project with no commit last and ties by name, with the
// folders inside each kept project; the workspace's own row stays, as
// does every project of a workspace without a limit. The root, _, takes
// a limit too, over the projects at its top.
func TestSideLimits_Apply(t *testing.T) {
	t.Parallel()
	ws := map[string]int{"malta": 0, "gozo": 1}
	folders := []*folder.Folder{
		jobs("infra", "plan"), jobs("infra/edge", "deploy"), jobs("private", "deploy"), jobs("apps", "build"),
		jobs("malta", "sync"), jobs("malta/apps", "build"), jobs("malta/apps/blog", "serve"), jobs("malta/dockerfiles", "build"), jobs("malta/vhosts", "deploy"), jobs("malta/old", "x"),
		jobs("gozo/one", "a"), jobs("gozo/two", "b"),
	}
	now := time.Date(2026, time.October, 25, 12, 0, 0, 0, time.UTC)
	when := func(id string) time.Time {
		switch id {
		case "malta/apps", "gozo/one":
			return now.Add(-time.Hour)
		case "malta/vhosts":
			return now
		case "malta/dockerfiles":
			return now.Add(-2 * time.Hour)
		case "infra":
			return now
		}
		return time.Time{}
	}

	kept := view.SideLimits{"malta": 2}.Apply(folders, ws, when)
	assert.Equal(t, []string{"infra", "infra/edge", "private", "apps", "malta", "malta/apps", "malta/apps/blog", "malta/vhosts", "gozo/one", "gozo/two"}, idsOf(kept), "vhosts and apps, the newest; dockerfiles and old out with the folders of each")

	// Limit 0, or none, keeps every project; a limit past the count too.
	assert.Equal(t, idsOf(folders), idsOf(view.SideLimits{"malta": 0}.Apply(folders, ws, when)))
	assert.Equal(t, idsOf(folders), idsOf(view.SideLimits{}.Apply(folders, ws, when)))
	assert.Equal(t, idsOf(folders), idsOf(view.SideLimits{"malta": 9}.Apply(folders, ws, when)))

	// Ties go by name, a project with no commit last: one project of
	// gozo keeps one, two has no commit; of malta's three with a commit,
	// three keeps them and drops old alone.
	assert.Equal(t, []string{"infra", "infra/edge", "private", "apps", "malta", "malta/apps", "malta/apps/blog", "malta/dockerfiles", "malta/vhosts", "gozo/one"}, idsOf(view.SideLimits{"gozo": 1, "malta": 3}.Apply(folders, ws, when)))

	// The root's limit counts the projects at its top: infra has a
	// commit, private and apps have none, so 1 keeps infra and its edge.
	assert.Equal(t, []string{"infra", "infra/edge", "malta", "malta/apps", "malta/apps/blog", "malta/dockerfiles", "malta/vhosts", "malta/old", "gozo/one", "gozo/two"}, idsOf(view.SideLimits{"internal": 1}.Apply(folders, ws, when)))
	assert.Equal(t, []string{"infra", "infra/edge", "apps"}, idsOf(view.SideLimits{"internal": 2}.Apply(folders, nil, when))[:3], "apps before private by name among those with no commit")
}
