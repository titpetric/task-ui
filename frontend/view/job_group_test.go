package view_test

import (
	"testing"

	"github.com/titpetric/task-ui/detector"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
)

// The Jobs table lists a group per source of the place, the files that
// run, a group per skill, the built-ins the user may run, then a group
// per source of each directory with targets of its own; every target is
// one row.
func TestNewJobGroups(t *testing.T) {
	t.Parallel()
	job := func(id, kind, source string) *folder.Job {
		j := &folder.Job{Folder: "infra", ID: id, Kind: kind, Source: source}
		j.Name = id
		return j
	}
	f := &folder.Folder{ID: "infra", Sources: []*detector.Source{
		{Kind: "atkins", File: "atkins.yml", Command: []string{"atkins", "-f", "atkins.yml"}},
		{Kind: "script", File: "scripts"},
	}}
	f.Jobs = []*folder.Job{job("default", "atkins", "atkins.yml"), job("build", "atkins", "atkins.yml"), job("script:seed", "script", "scripts")}
	files := []*folder.Job{job("./run.sh", "file", "run.sh")}
	git := &folder.Skill{ID: "git", Title: "Git checkout", Jobs: []*folder.Job{job("@git:pull", "skill", ""), job("@git:status", "skill", "")}}
	tools := &folder.Folder{ID: "infra", Rel: "tools", Sources: []*detector.Source{{Kind: "make", File: "Makefile"}}}
	check := job("check", "make", "Makefile")
	check.Dir = "tools"
	tools.Jobs = []*folder.Job{check}

	ids := func(groups []view.JobGroup) (kinds []string, rows []string) {
		for _, g := range groups {
			kinds = append(kinds, g.Kind)
			for _, j := range g.Jobs {
				rows = append(rows, j.ID)
			}
		}
		return kinds, rows
	}

	// Without the shell permission Actions holds no built-ins.
	groups := view.NewJobGroups(f, f, files, &view.Actions{Place: f, Skills: []*folder.Skill{git}}, []*folder.Folder{tools})
	kinds, rows := ids(groups)
	assert.Equal(t, []string{"atkins", "script", "files", "@git", "make"}, kinds)
	assert.Equal(t, []string{"default", "build", "script:seed", "./run.sh", "@git:pull", "@git:status", "check"}, rows)
	assert.Equal(t, "atkins.yml · atkins -f atkins.yml", groups[0].Note)
	assert.Equal(t, "Git checkout", groups[3].Note)
	assert.Equal(t, git, groups[3].Skill)
	assert.Equal(t, "tools", groups[4].Dir)
	assert.Equal(t, tools, groups[4].Folder)

	// With it, @shell and @claude follow the skills.
	a := &view.Actions{Place: f, Shell: f.ShellJob(), Claude: f.ClaudeJob()}
	kinds, rows = ids(view.NewJobGroups(f, f, nil, a, nil))
	assert.Equal(t, []string{"atkins", "script", "built-in"}, kinds)
	assert.Equal(t, []string{"default", "build", "script:seed", folder.ShellID, folder.ClaudeID}, rows)

	// A browsed directory with no source of its own lists its files and
	// skills, its directories by their path below it.
	sub := &folder.Folder{ID: "infra", Rel: "deploy"}
	nested := &folder.Folder{ID: "infra", Rel: "deploy/tf", Sources: []*detector.Source{{Kind: "make", File: "Makefile"}}}
	groups = view.NewJobGroups(f, sub, files, &view.Actions{Place: sub}, []*folder.Folder{nested})
	kinds, _ = ids(groups)
	assert.Equal(t, []string{"files", "make"}, kinds)
	assert.Equal(t, f, groups[0].Folder)
	assert.Equal(t, "tf", groups[1].Dir)
	assert.Equal(t, "No targets in Makefile.", groups[1].Empty)

	// A source with errors says so.
	broken := &folder.Folder{ID: "x", Error: "bad yaml", Sources: []*detector.Source{{Kind: "atkins", File: "atkins.yml"}}}
	assert.Equal(t, "bad yaml", view.NewJobGroups(broken, broken, nil, nil, nil)[0].Empty)
}
