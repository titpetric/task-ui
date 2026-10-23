package view_test

import (
	"bytes"
	"strconv"
	"testing"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/skills"
)

// broken is a page whose folders hold lint findings: two in one
// pipeline of the workspace malta, one in a project at the root.
func broken(t *testing.T) *view.Page {
	t.Helper()
	p := admin(t)
	p.Workspaces = map[string]int{"malta": 1}
	clean := &folder.Folder{ID: "malta", Title: "Malta"}
	app := &folder.Folder{ID: "malta/apps/task", Title: "task", Errors: []skills.Finding{
		{Path: "atkins.yml", Job: "build", Issue: "missing task reference", Detail: `step references task 'go:fmt', but job "go:fmt" not found <b>`},
		{Path: "atkins.yml", Job: "test", Detail: "second"},
	}, Error: "atkins.yml: pipeline 'task' has errors"}
	root := &folder.Folder{ID: "site", Title: "site", Errors: []skills.Finding{{Path: ".atkins.yml", Detail: "third"}}, Error: "x"}
	p.Folders = append(p.Folders, clean, app, root)
	p.SideFolders = p.Folders
	p.Folder = clean
	return p
}

// The alert of a place gathers the findings of the folder and every
// folder inside it, counts the pipelines, and is nil where none is.
func TestPage_PipelineAlert(t *testing.T) {
	t.Parallel()
	p := broken(t)

	alert := p.PipelineAlert("malta")
	assert.NotNil(t, alert)
	assert.Equal(t, 1, alert.Pipelines)
	assert.Equal(t, 2, alert.Rows())
	assert.Equal(t, []string{"malta"}, alert.Workspaces)
	assert.Equal(t, "malta/apps/task", alert.Shown[0].Folder.ID)

	assert.Equal(t, 1, p.PipelineAlert("malta/apps/task").Pipelines)
	assert.Nil(t, p.PipelineAlert("malta/apps"+"x"))
	assert.Nil(t, p.PipelineAlert("infra"))

	root := p.PipelineAlert("_")
	assert.Equal(t, 1, root.Pipelines)
	assert.Equal(t, "site", root.Shown[0].Folder.ID)

	all := p.PipelineAlert("")
	assert.Equal(t, 2, all.Pipelines)
	assert.Equal(t, []string{"malta", "_"}, all.Workspaces)
}

// Past five rows the rest fold under Show more.
func TestPipelineAlert_Rows(t *testing.T) {
	t.Parallel()
	p := admin(t)
	f := &folder.Folder{ID: "many"}
	for i := range 7 {
		f.Errors = append(f.Errors, skills.Finding{Path: "atkins.yml", Job: "j" + strconv.Itoa(i), Detail: "d"})
	}
	p.Folders = []*folder.Folder{f}
	alert := p.PipelineAlert("many")
	assert.Len(t, alert.Shown, 5)
	assert.Len(t, alert.More, 2)
	assert.Equal(t, 7, alert.Rows())
	assert.Equal(t, 1, alert.Pipelines)
}

// The workspace page shows the alert with the detail escaped, the
// sidebar a badge with the count, and the dashboard its line; a clean
// page shows none of it.
func TestView_RenderPipelineAlert(t *testing.T) {
	t.Parallel()
	v, err := view.New()
	assert.NoError(t, err)
	render := func(name string, p *view.Page) string {
		var buf bytes.Buffer
		assert.NoError(t, v.Render(&buf, name, p))
		return buf.String()
	}

	page := render("folder", broken(t))
	assert.Contains(t, page, "1 pipeline has errors.")
	assert.Contains(t, page, `<a href="/f/malta/apps/task">malta/apps/task</a> <code>atkins.yml</code>`)
	assert.Contains(t, page, "go:fmt&#34; not found &lt;b&gt;")
	assert.NotContains(t, page, "not found <b>")
	assert.Contains(t, page, `<small class="bad" title="step references task &#39;go:fmt&#39;, but job &#34;go:fmt&#34; not found &lt;b&gt;">!2</small>`)
	assert.Contains(t, page, "The pipeline has errors.")
	assert.NotContains(t, page, "more</summary>")

	home := render("home", broken(t))
	assert.Contains(t, home, "2 pipelines have errors</b> in")
	assert.Contains(t, home, `<a href="/f/_">the default workspace</a>`)

	clean := admin(t)
	for _, name := range []string{"folder", "home"} {
		out := render(name, clean)
		assert.NotContains(t, out, "pipeline-alert", name)
		assert.NotContains(t, out, "have errors", name)
		assert.NotContains(t, out, `class="bad" title`, name)
	}
}
