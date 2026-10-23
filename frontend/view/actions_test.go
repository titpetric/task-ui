package view_test

import (
	"testing"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/git"
	"github.com/titpetric/task-ui/internal/assert"
)

// Up and down of the compose project task-ui runs in ask first; any other
// job, project or skill does not.
func TestActions_Confirm(t *testing.T) {
	t.Parallel()
	job := func(name string) *folder.Job {
		j := &folder.Job{ID: "@compose:" + name}
		j.Name = name
		return j
	}
	a := &view.Actions{Self: "task"}
	self := &folder.Skill{ID: "compose", Project: "task"}
	other := &folder.Skill{ID: "compose", Project: "blog"}
	assert.Contains(t, a.Confirm(self, job("up")), "task-ui itself")
	assert.Contains(t, a.Confirm(self, job("down")), "stops task-ui")
	assert.Equal(t, "", a.Confirm(self, job("pull")))
	assert.Equal(t, "", a.Confirm(other, job("up")))
	assert.Equal(t, "", a.Confirm(&folder.Skill{ID: "dockerfile", Project: "task"}, job("up")))
	assert.Equal(t, "", (&view.Actions{}).Confirm(self, job("up")))
	assert.Equal(t, "", (*view.Actions)(nil).Confirm(self, job("up")))
}

// A bar without a git line, skills or jobs is empty.
func TestActions_Empty(t *testing.T) {
	t.Parallel()
	assert.True(t, (*view.Actions)(nil).Empty())
	assert.True(t, (&view.Actions{}).Empty())
	assert.False(t, (&view.Actions{Git: &git.Status{Branch: "main"}}).Empty())
	assert.False(t, (&view.Actions{Skills: []*folder.Skill{{ID: "git"}}}).Empty())
	assert.False(t, (&view.Actions{Shell: (&folder.Folder{ID: "site"}).ShellJob()}).Empty())
}
