package view_test

import (
	"testing"
	"time"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

func TestPage_CanRun(t *testing.T) {
	t.Parallel()
	p := admin(t)
	f, job := p.Folder, p.Job

	assert.True(t, p.CanRun(f, job))
	assert.False(t, p.CanRun(nil, job))
	assert.False(t, p.CanRun(f, nil))

	p.User = &model.User{Name: "guest", Role: model.RoleViewer}
	assert.False(t, p.CanRun(f, job))

	p.User = &model.User{Name: "ops", Role: model.RoleRunner}
	assert.True(t, p.CanRun(f, job))
	job.Allow = []string{"admin"}
	assert.False(t, p.CanRun(f, job))
}

func TestPage_Runs(t *testing.T) {
	t.Parallel()
	p := admin(t)
	key := p.Recents[0].Entries[0].Run.Key()

	assert.NotNil(t, p.RunningRun(key))
	assert.Equal(t, "ops", p.RunningRun(key).UserName)
	assert.Nil(t, p.RunningRun("infra/other"))

	assert.Nil(t, p.LastRun(key))
	p.Last = map[string]view.Run{key: *p.Recents[0].Entries[0].Run}
	assert.NotNil(t, p.LastRun(key))
}

func TestPage_Paths(t *testing.T) {
	t.Parallel()
	p := &view.Page{Path: "/settings/keys"}

	assert.True(t, p.IsPath("/settings/keys"))
	assert.False(t, p.IsPath("/settings"))
	assert.True(t, p.Under("/settings"))
	assert.True(t, p.Under("/settings/keys"))
	assert.False(t, p.Under("/set"))
}

func TestPage_InHours(t *testing.T) {
	t.Parallel()

	p := &view.Page{Now: time.Date(2026, time.October, 16, 22, 0, 0, 0, time.UTC)}
	assert.False(t, p.InHours("mon-fri 9-17"))
	assert.True(t, p.InHours(""))
	assert.True(t, (&view.Page{}).InHours("daily 0-24"))
}

func TestPage_WhoCanRun(t *testing.T) {
	t.Parallel()
	p := admin(t)
	f, job := p.Folder, p.Job

	assert.Equal(t, "admin, runner", p.WhoCanRun(f, job))
	f.Allow = []string{"ops"}
	job.Allow = []string{"release"}
	assert.Equal(t, "admin, ops, release", p.WhoCanRun(f, job))
	assert.Equal(t, []string{"ops"}, f.Allow)

	var _ *folder.Folder = f
}

func TestPage_CanRunJob(t *testing.T) {
	t.Parallel()
	p := admin(t)

	assert.True(t, p.CanRunAny(p.Folder))
	assert.True(t, p.CanRunJob("infra", "plan"))
	assert.False(t, p.CanRunJob("infra", "nothing"))
	assert.False(t, p.CanRunJob("nowhere", "plan"))

	p.User = &model.User{Name: "guest", Role: model.RoleViewer}
	assert.False(t, p.CanRunAny(p.Folder))
	assert.False(t, p.CanRunJob("infra", "plan"))
}

// The sidebar's groups come from the folders it lists and the
// workspaces: a workspace holds its projects.
func TestPage_Groups(t *testing.T) {
	t.Parallel()
	p := &view.Page{SideFolders: []*folder.Folder{{ID: "malta"}, {ID: "malta/deploy"}}}
	assert.Equal(t, []string{"malta/deploy"}, idsOf(p.Groups()[0].Nested))

	p.Workspaces = map[string]int{"malta": 0}
	groups := p.Groups()
	assert.Len(t, groups, 1)
	assert.True(t, groups[0].Workspace)
	assert.Equal(t, "malta/deploy", groups[0].Projects[0].Top.ID)
}

// A page links a folder among its workspaces: a named one by its folder,
// a project at the root under /f/_/.
func TestPage_Href(t *testing.T) {
	t.Parallel()
	p := &view.Page{Workspaces: map[string]int{"acme": 0}}
	assert.Equal(t, "/f/acme/api", p.Href("acme/api"))
	assert.Equal(t, "/f/_/infra", p.Href("infra"))
	assert.Equal(t, "/f/_/infra", (&view.Page{}).Href("infra"))
}

func TestPage_Links(t *testing.T) {
	t.Parallel()
	p := &view.Page{Workspaces: map[string]int{"acme": 0}}
	links := p.Links([]view.Step{{Folder: "acme/api", Job: "build"}, {Folder: "infra", Job: "plan"}})
	assert.Equal(t, "/f/acme/api/build", links[0].Href)
	assert.Equal(t, "/f/_/infra/plan", links[1].Href)
}
