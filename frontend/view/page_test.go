package view_test

import (
	"testing"

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
	key := p.Recents[0].Key()

	assert.NotNil(t, p.RunningRun(key))
	assert.Equal(t, "ops", p.RunningRun(key).UserName)
	assert.Nil(t, p.RunningRun("infra/other"))

	assert.Nil(t, p.LastRun(key))
	p.Last = map[string]view.Run{key: p.Recents[0]}
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
