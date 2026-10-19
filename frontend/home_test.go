package frontend_test

import (
	"context"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

func TestHome(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.run(t, "infra", "plan", "ops", model.StateOK)
	assert.NoError(t, f.opts.Schedules.Create(context.Background(), &model.Schedule{Folder: "private", Job: "deploy", Expression: "@daily", Enabled: 1}))

	body := f.page(t, asAdmin, "/")
	assert.Contains(t, body, "<h2>Dashboard</h2>")
	assert.Contains(t, body, "Nothing is running")
	assert.Contains(t, body, "<b>Scheduled</b><span>1</span>")
	assert.Contains(t, body, "<b>Folders</b><span>3</span>")
	assert.Contains(t, body, "<b>Jobs</b><span>5</span>")
	// No recent runs or folders tables: the control centre instead.
	assert.NotContains(t, body, "<h3>Recent runs</h3>")
	assert.NotContains(t, body, "<h3>Folders</h3>")
	assert.Contains(t, body, "<h3>Control centre")

	limited := f.page(t, asLimited, "/")
	assert.Contains(t, limited, "<b>Scheduled</b><span>0</span>")
	assert.Contains(t, limited, "<b>Folders</b><span>1</span>")
}

// The control centre gives deploy, which private and infra/edge share, a
// column; infra's own jobs are under other, each a button by name.
func TestHome_Control(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	session := f.start(t, f.infra, "slow")

	body := f.page(t, asOps, "/")
	assert.Contains(t, body, `<thead><tr><th>Folder</th><th class="col">deploy</th><th>Other</th></tr></thead>`)
	edge := body[strings.Index(body, `<a href="/f/infra/edge">infra/edge</a>`):]
	edge = edge[:strings.Index(edge, "</tr>")]
	assert.Contains(t, edge, `<button type="button" class="cc" data-run data-folder="infra/edge" data-job="deploy"`)

	infra := body[strings.Index(body, `<td class="key nowrap"><a href="/f/infra">infra</a>`):]
	infra = infra[:strings.Index(infra, "</tr>")]
	assert.Contains(t, infra, `<td class="col" data-label="deploy"></td>`) // infra has no deploy
	assert.Contains(t, infra, `<td class="other" data-label="Other"><div>`)
	assert.Contains(t, infra, `class="cc named" data-run data-folder="infra" data-job="plan"`)
	assert.Contains(t, infra, `<span>plan</span>`)
	assert.Contains(t, infra, `<a class="cc named running" href="/runs/`+session.ID()+`"`) // slow runs

	// ops may not run private's deploy: a #! to its page instead.
	private := body[strings.Index(body, `<a href="/f/private">private</a>`):]
	private = private[:strings.Index(private, "</tr>")]
	assert.Contains(t, private, `<a class="cc idle" href="/f/private/deploy"`)
	assert.NotContains(t, private, "data-run")
}

func TestHome_Live(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	session := f.start(t, f.infra, "slow")

	body := f.page(t, asViewer, "/")
	assert.Contains(t, body, "<b>Running</b><span>1</span>")
	assert.Contains(t, body, `<div class="frame" data-frame="`+session.ID()+`" role="img"`)
	assert.Contains(t, body, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 672.0 136.0" font-family="JetBrains Mono`)
	assert.Contains(t, body, `data-attach="`+session.ID()+`"`)
	assert.Contains(t, body, `<i class="dot running"></i>`)
}
