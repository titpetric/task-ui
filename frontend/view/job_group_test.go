package view_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/titpetric/task-ui/detector"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
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

// The Jobs table has a row per job of its groups, under a heading row
// per group: the job's ID linked to its page, Run, with the field for
// args, and the confirm of a skill job, and a [...] of Run and
// summarize, the last run, History, and Run force for a job with hours.
// A job the user may not run has History; a running one Attach and Kill.
func TestView_RenderJobTable(t *testing.T) {
	t.Parallel()

	v := shared(t)
	render := func(p *view.Page) string {
		var buf bytes.Buffer
		assert.NoError(t, v.Render(&buf, "folder", p))
		return buf.String()
	}
	table := func(p *view.Page) {
		f := p.Folder
		f.Hours = "Mon-Fri 9-17"
		build := &folder.Job{Folder: "infra", ID: "build", Command: []string{"make", "build"}}
		build.Name, build.Description = "build", "Build the images"
		custom := &folder.Job{Folder: "infra", ID: "@compose:custom", Command: []string{"atkins", "compose:custom"}}
		custom.Name, custom.TakesArgs = "custom", true
		up := &folder.Job{Folder: "infra", ID: "@compose:up", Command: []string{"atkins", "compose:up"}}
		up.Name = "up"
		script := &folder.Job{Folder: "infra", ID: "./deploy.sh", Kind: "file", Source: "deploy.sh", Command: []string{"./deploy.sh"}}
		script.Name = "deploy.sh"
		compose := &folder.Skill{ID: "compose", Project: "infra", Jobs: []*folder.Job{custom, up}}
		p.Actions = &view.Actions{Place: f, Self: "infra", Shell: f.ShellJob(), Skills: []*folder.Skill{compose}}
		p.JobGroups = []view.JobGroup{
			{Kind: "make", Note: "Makefile · make", Source: "Makefile", Folder: f, Jobs: []*folder.Job{build}, Targets: true},
			{Kind: "files", Folder: f, Jobs: []*folder.Job{script}},
			{Kind: "@compose", Folder: f, Skill: compose, Jobs: compose.Jobs},
			{Kind: "built-in", Folder: f, Jobs: []*folder.Job{f.ShellJob()}},
			{Kind: "make", Source: "Makefile", Dir: "tools", Folder: &folder.Folder{ID: "infra", Rel: "tools"}, Empty: "No targets in Makefile.", Targets: true},
		}
	}

	p := admin(t)
	p.Summaries = true
	table(p)
	p.Last = map[string]view.Run{
		"infra/build":       {Run: model.Run{ID: "r1", Folder: "infra", Job: "build", State: model.StateOK}},
		"infra/./deploy.sh": {Run: model.Run{ID: "r3", Folder: "infra", Job: "./deploy.sh", State: model.StateFailed}},
	}
	p.JobSummaries = map[string]*view.SummaryLink{"infra/build": view.NewSummaryLink("r1", "The build passed.")}
	out := render(p)
	// The dashboard's control component as named rows: a source's
	// targets under its file, linked to where they run, then the files,
	// the skill and the built-ins, each with its [...]; no columns, no
	// other.
	assert.Contains(t, out, `<table class="table control col-hover">`)
	assert.NotContains(t, out, `<th class="col">`)
	assert.Contains(t, out, `<a class="badge" href="/f/internal/infra">Makefile</a>`)
	assert.Contains(t, out, `<a class="badge" href="/f/internal/infra/tools">tools/Makefile</a>`)
	assert.Contains(t, out, `<span class="faint">No targets in Makefile.</span>`)
	assert.Equal(t, 5, strings.Count(out, `<tr class="named" data-span>`))
	for _, kind := range []string{"files", "@compose", "built-in"} {
		assert.Contains(t, out, `<span class="badge">`+kind+`</span>`, kind)
	}
	for _, id := range []string{"build", "./deploy.sh", "@compose:custom", "@compose:up", "@shell"} {
		assert.Contains(t, out, `data-job="`+id+`"`, id)
	}
	assert.Contains(t, out, `<button type="button" class="cc named ok" data-run data-folder="infra" data-job="build" data-title="infra / build" title="Run infra / build: Build the images">`)
	assert.Contains(t, out, `class="cc named bad" data-run data-folder="infra" data-job="./deploy.sh"`)
	assert.Contains(t, out, `class="cc named never" data-run data-folder="infra" data-job="@compose:custom"`)
	assert.Contains(t, out, `data-path="deploy.sh"`)
	assert.Contains(t, out, `data-args="atkins compose:custom --"`)
	assert.Contains(t, out, `data-confirm="Run compose up for infra?`)
	assert.Contains(t, out, `<span>Run and summarize build</span>`)
	assert.Contains(t, out, `<span>Run and summarize @compose:custom...</span>`)
	assert.Contains(t, out, `<span>Run force build</span>`)
	assert.Contains(t, out, `data-force="" title="Run out of hours, with a reason" data-run="" data-folder="infra" data-dir="" data-job="build"`)
	assert.Equal(t, 5, strings.Count(out, "data-force="), "Run force on every job, the built-ins too, in a workspace with hours")
	assert.Contains(t, out, `<span>View the last run of build</span>`)
	assert.Contains(t, out, `<span>History of @shell</span>`)
	assert.Contains(t, out, `href="/runs/r1/report" title="The build passed."`)
	assert.Equal(t, "/runs/r1/report", p.JobSummary("infra/build").Href)
	assert.Nil(t, p.JobSummary("infra/other"))

	viewer := admin(t)
	table(viewer)
	signIn(viewer, &model.User{Name: "guest", Role: model.RoleViewer}, nil)
	out = render(viewer)
	assert.NotContains(t, out, "data-run ")
	assert.Contains(t, out, `<a class="cc named idle never" href="/f/internal/infra/build" title="infra / build: Build the images: no permission to run it">`)

	running := admin(t)
	table(running)
	now := time.Now()
	run := view.Run{Run: model.Run{ID: "r2", Folder: "infra", Job: "build", State: model.StateRunning, StartedAt: &now}}
	running.Running = map[string]view.Run{"infra/build": run}
	out = render(running)
	assert.Contains(t, out, `<a class="cc named running" href="/runs/r2" data-attach="r2" title="Attach to infra / build, running">`)
	assert.Contains(t, out, `data-kill="r2"`)
}
