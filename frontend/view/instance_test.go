package view_test

import (
	"bytes"
	"testing"

	"github.com/titpetric/task-ui/diskusage"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/toolenv"
)

// The Instance page's Storage table lists each item with its size and
// files, an ellipsis for a pending one, and the total.
func TestInstance_Storage(t *testing.T) {
	t.Parallel()
	v, err := view.New()
	assert.NoError(t, err)
	p := admin(t)
	p.Instance = &view.Instance{Root: "/data/workspaces", Storage: &diskusage.Table{
		Rows: []diskusage.Row{
			{Name: "task-ui database", Where: "/data/task-ui.db", Size: 811008, Files: 3},
			{Name: "Checkouts · malta", Where: "/data/workspaces/malta/src", Pending: true},
		},
		Total: diskusage.Row{Name: "Total", Size: 811008, Files: 3},
	}}
	var buf bytes.Buffer
	assert.NoError(t, v.Render(&buf, "instance", p))
	body := buf.String()
	assert.Contains(t, body, `<section class="block instance-section" id="storage">`)
	assert.Contains(t, body, `<td class="key">task-ui database</td><td class="where mono" data-label="Where">/data/task-ui.db</td><td class="num mono" data-label="Size">792.0 KB</td><td class="num mono" data-label="Files">3</td>`)
	assert.Contains(t, body, `<td class="key">Checkouts · malta</td>`)
	assert.Contains(t, body, `<span class="faint pending" aria-label="not measured yet"></span>`)
	assert.Contains(t, body, `<tfoot><tr><th>Total</th><td></td><td class="num mono" data-label="Size">792.0 KB</td>`)

	p.Instance.Storage = nil
	buf.Reset()
	assert.NoError(t, v.Render(&buf, "instance", p))
	assert.NotContains(t, buf.String(), `id="storage"`)
}

// The Tooling section lists the detected tools in a table per category,
// the tool, its version, its path and what it is used for; a category
// with no tool has no table, and no tool at all says so.
func TestInstance_Tooling(t *testing.T) {
	t.Parallel()
	v, err := view.New()
	assert.NoError(t, err)
	p := admin(t)
	p.Instance = &view.Instance{Tooling: view.ToolGroups([]model.Tool{
		{Name: "go", Category: toolenv.CategoryRuntime, Version: "go1.22.1", Path: "/usr/local/go/bin/go", Description: "Go toolchain: builds and tests Go"},
		{Name: "git", Category: toolenv.CategoryTooling, Version: "2.43.0", Path: "/usr/bin/git", Description: "version control"},
	})}
	var buf bytes.Buffer
	assert.NoError(t, v.Render(&buf, "instance", p))
	body := buf.String()
	assert.Contains(t, body, `<section class="block instance-section" id="tooling">`)
	assert.Contains(t, body, `<h4>Runtimes</h4>`)
	assert.Contains(t, body, `<td class="key mono" data-label="Tool">go</td><td class="mono" data-label="Version">go1.22.1</td><td class="path mono" data-label="Path">/usr/local/go/bin/go</td><td class="desc" data-label="Used for">Go toolchain: builds and tests Go</td><td class="actions"><form class="inline" data-api="/api/instance/tools/go/check" data-method="POST" data-next="/settings/instance?notice=Tool+checked#tooling"><button class="btn btn-sm" type="submit">Check</button></form></td>`)
	assert.Contains(t, body, `data-api="/api/instance/tools/check" data-method="POST" data-next="/settings/instance?notice=Tooling+checked#tooling"><button class="btn btn-sm" type="submit">Check all</button>`)
	assert.Contains(t, body, `<h4>Tooling</h4>`)
	assert.Contains(t, body, `<td class="key mono" data-label="Tool">git</td>`)
	assert.Contains(t, body, `data-api="/api/instance/tools/git/check"`)
	assert.NotContains(t, body, `<h4>Custom tooling</h4>`, "a category with no tool is left out")

	p.Instance = &view.Instance{}
	buf.Reset()
	assert.NoError(t, v.Render(&buf, "instance", p))
	assert.Contains(t, buf.String(), "No tooling was detected: Check all probes the PATH.")
}

// The Parent environment section reads the base variables and the
// sensitive words from the runner.
func TestInstance_Env(t *testing.T) {
	t.Parallel()
	i := &view.Instance{}
	assert.Equal(t, runner.BaseEnv, i.EnvBase())
	assert.Equal(t, runner.SensitiveWords, i.EnvSensitive())
	assert.Contains(t, i.EnvBase(), "DOCKER_CONFIG")
	assert.Contains(t, i.EnvBase(), "GOPATH")
}
