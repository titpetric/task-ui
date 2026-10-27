//go:build integration

package frontend_test

import (
	"context"
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
	// No numbers above the tiles, and no control centre.
	assert.NotContains(t, body, `class="metrics"`)
	assert.NotContains(t, body, "Control centre")

	limited := f.page(t, asLimited, "/")
	assert.Contains(t, limited, "<h2>Dashboard</h2>")
}

func TestHome_Live(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	session := f.start(t, f.infra, "slow")

	body := f.page(t, asViewer, "/")
	assert.Contains(t, body, `<div class="frame" data-frame="`+session.ID()+`" role="img"`)
	assert.Contains(t, body, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 672.0 136.0" font-family="JetBrains Mono`)
	assert.Contains(t, body, `data-attach="`+session.ID()+`"`)
	assert.Contains(t, body, `<i class="dot running"></i>`)
}
