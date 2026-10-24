package view_test

import (
	"bytes"
	"testing"

	"github.com/titpetric/task-ui/diskusage"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/runner"
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
