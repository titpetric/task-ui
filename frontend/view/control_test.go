package view_test

import (
	"bytes"
	"testing"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// jobs makes a folder of jobs: "name" or "kind:name".
func jobs(id string, ids ...string) *folder.Folder {
	f := &folder.Folder{ID: id}
	for _, jobID := range ids {
		j := &folder.Job{Folder: id, ID: jobID}
		j.Name = jobID
		if kind, name, ok := cutKind(jobID); ok {
			j.Kind, j.Name = kind, name
		}
		f.Jobs = append(f.Jobs, j)
	}
	return f
}

func cutKind(id string) (string, string, bool) {
	for i := range id {
		if id[i] == ':' {
			return id[:i], id[i+1:], true
		}
	}
	return "", "", false
}

// Every job name that occurs more than once, across folders or within
// one, gets a column, the most common first; a cell holds the
// folder's jobs of that name. Other holds the names that occur once. A
// folder without jobs has no row.
func TestControl(t *testing.T) {
	t.Parallel()
	folders := []*folder.Folder{
		jobs("api", "build", "test", "push", "lint"),
		jobs("web", "build", "test", "serve"),
		jobs("docs", "build", "deploy"),
		jobs("empty"),
		jobs("infra", "plan", "apply", "test"),
		jobs("tools", "make:release", "script:release", "make:lint"),
	}

	c := view.NewControl(folders, 2)
	assert.Equal(t, []string{"build", "test", "lint", "release"}, c.Columns)

	// A name needs as many as min to get a column: build and test
	// occur three times, lint and release twice.
	three := view.NewControl(folders, 3)
	assert.Equal(t, []string{"build", "test"}, three.Columns)
	assert.Equal(t, []string{"make:release", "script:release", "make:lint"}, names(three.Rows[4].Extra))
	assert.Len(t, view.NewControl(folders, 4).Columns, 0)
	assert.Equal(t, 2, view.ControlRepeats) // more than once
	assert.Len(t, c.Rows, 5)

	api := c.Rows[0]
	assert.Equal(t, "api", api.Folder.ID)
	assert.Equal(t, []string{"build"}, names(api.Cells[0]))
	assert.Equal(t, []string{"lint"}, names(api.Cells[2]))
	assert.Equal(t, []string{"push"}, names(api.Extra))

	docs := c.Rows[2]
	assert.Len(t, docs.Cells[1], 0) // docs has no test
	assert.Equal(t, []string{"deploy"}, names(docs.Extra))

	// Two of a name in one folder share its cell; nothing repeats under
	// other.
	tools := c.Rows[4]
	assert.Equal(t, []string{"make:release", "script:release"}, names(tools.Cells[3]))
	assert.Equal(t, []string{"make:lint"}, names(tools.Cells[2]))
	assert.Len(t, tools.Extra, 0)

	// default leads, whatever the counts.
	lead := view.NewControl([]*folder.Folder{jobs("a", "build", "test", "default"), jobs("b", "build", "test", "default"), jobs("c", "build")}, 2)
	assert.Equal(t, []string{"default", "build", "test"}, lead.Columns)
}

func names(js []*folder.Job) []string {
	out := []string{}
	for _, j := range js {
		out = append(out, j.ID)
	}
	return out
}

// A row is a folder: its jobs of each column, and the rest under other.
func TestControlRow(t *testing.T) {
	t.Parallel()
	c := view.NewControl([]*folder.Folder{jobs("api", "build", "push"), jobs("web", "build", "serve")}, view.ControlRepeats)
	var row view.ControlRow = c.Rows[1]
	assert.Equal(t, "web", row.Folder.ID)
	assert.Equal(t, []string{"build"}, names(row.Cells[0]))
	assert.Equal(t, []string{"serve"}, names(row.Extra))
}

// A folder's page lays its Jobs out as named rows, no columns and no
// other: a row per job source under its file, below its directory for
// one inside, then a row per other group under its kind.
func TestNewJobControl(t *testing.T) {
	t.Parallel()
	place := jobs("malta/dockerfiles", "default", "build", "push")
	task := jobs("malta/dockerfiles", "default", "build", "push", "deploy")
	task.Rel = "task"
	broken := &folder.Folder{ID: "malta/dockerfiles", Rel: "broken"}
	git := jobs("malta/dockerfiles", "@git:pull", "@git:status")
	groups := []view.JobGroup{
		{Kind: "atkins", Note: "atkins.yml · atkins -f atkins.yml", Source: "atkins.yml", Folder: place, Jobs: place.Jobs, Targets: true},
		{Kind: "@git", Note: "Git", Folder: place, Jobs: git.Jobs},
		{Kind: "atkins", Note: "atkins.yml", Source: "atkins.yml", Dir: "task", Folder: task, Jobs: task.Jobs, Targets: true},
		{Kind: "make", Note: "Makefile", Source: "Makefile", Dir: "broken", Folder: broken, Targets: true, Empty: "No targets in Makefile."},
	}
	c := view.NewJobControl(groups)
	assert.Len(t, c.Columns, 0)
	labels := []string{}
	for _, row := range c.Rows {
		assert.True(t, row.Named)
		labels = append(labels, row.Label)
	}
	assert.Equal(t, []string{"atkins.yml", "@git", "task/atkins.yml", "broken/Makefile"}, labels)
	assert.True(t, c.Rows[0].Place)
	assert.False(t, c.Rows[1].Place)
	assert.Equal(t, "atkins atkins.yml · atkins -f atkins.yml", c.Rows[0].Title)
	assert.Equal(t, []string{"default", "build", "push", "deploy"}, names(c.Rows[2].Jobs()))
	assert.Equal(t, []string{"@git:pull", "@git:status"}, names(c.Rows[1].Extra))
	assert.Equal(t, "No targets in Makefile.", c.Rows[3].Empty)
}

// The dashboard draws the same control component: a row per folder, a
// column per shared job, the rest under other, and no [...] per row.
func TestView_RenderDashboard(t *testing.T) {
	t.Parallel()

	v, err := view.New()
	assert.NoError(t, err)
	p := admin(t)
	p.Running = nil
	web := &folder.Job{Folder: "web", ID: "plan"}
	web.Name = "plan"
	p.Folders = append(p.Folders, &folder.Folder{ID: "web", Jobs: []*folder.Job{web}})
	p.Control = view.NewControl(p.Folders, view.ControlRepeats)
	p.Last = map[string]view.Run{"infra/plan": {Run: model.Run{ID: "r1", Folder: "infra", Job: "plan", State: model.StateOK}}}
	var buf bytes.Buffer
	assert.NoError(t, v.Render(&buf, "home", p))
	out := buf.String()
	assert.Contains(t, out, `<thead><tr><th>Folder</th><th class="col"><span>plan</span></th><th class="col"><span>other</span></th></tr></thead>`)
	assert.Contains(t, out, `<td class="key nowrap"><a href="/f/_/infra">infra</a></td>`)
	assert.Contains(t, out, `<span class="cc-slot" data-cc="infra/plan" data-folder="infra" data-job="plan" data-name="plan" data-label="" data-last="ok" data-can><button type="button" class="cc ok" data-run data-folder="infra" data-job="plan" data-title="infra / plan" title="Run infra / plan"><svg class="ic"><use href="#i-play"/></svg></button></span>`)
	assert.NotContains(t, out, `class="named"`)
	assert.NotContains(t, out, `<td class="actions">`)
}
