package view_test

import (
	"testing"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// A Jobs row's [...] holds a group per job; Run force needs hours, Kill a run.
func TestPage_ControlMenu(t *testing.T) {
	t.Parallel()
	f := &folder.Folder{ID: "infra", Hours: "mon-fri 9-17"}
	build := &folder.Job{Folder: "infra", ID: "build"}
	build.Name = "build"
	up := &folder.Job{Folder: "infra", ID: "@compose:up"}
	up.Name = "up"
	f.Jobs = []*folder.Job{build}
	p := admin(t)
	p.Summaries = true
	p.Running = map[string]view.Run{"infra/build": {Run: model.Run{ID: "r2", Folder: "infra", Job: "build", State: model.StateRunning}}}
	p.Actions = &view.Actions{Place: f, Self: "infra", Skills: []*folder.Skill{{ID: "compose", Project: "infra", Jobs: []*folder.Job{up}}}}
	p.JobGroups = []view.JobGroup{{Kind: "make", Folder: f, Jobs: f.Jobs, Targets: true}, {Kind: "@compose", Folder: f, Jobs: []*folder.Job{up}}}
	c := p.JobControl()
	assert.Len(t, c.Rows, 2)

	labels := func(m *view.Menu) []string {
		out := []string{}
		for _, g := range m.Groups() {
			for _, item := range g {
				out = append(out, item.Label)
			}
		}
		return out
	}
	assert.Equal(t, []string{"History of build", "Kill build"}, labels(p.ControlMenu(c.Rows[0])))
	assert.Equal(t, []string{"Run and summarize @compose:up", "History of @compose:up", "Run force @compose:up"}, labels(p.ControlMenu(c.Rows[1])))
	assert.Equal(t, "Actions on @compose", p.ControlMenu(c.Rows[1]).Label)
}

// A skill's job takes the question Actions.Confirm returns; other jobs take none.
func TestPage_RunConfirm(t *testing.T) {
	t.Parallel()
	up := &folder.Job{ID: "@compose:up"}
	up.Name = "up"
	p := &view.Page{}
	assert.Equal(t, "", p.RunConfirm(up))
	p.Actions = &view.Actions{Self: "infra", Skills: []*folder.Skill{{ID: "compose", Project: "infra", Jobs: []*folder.Job{up}}}}
	assert.Contains(t, p.RunConfirm(up), "Run compose up for infra?")
	assert.Equal(t, "", p.RunConfirm(&folder.Job{ID: "build"}))
}
