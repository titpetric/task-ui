package view_test

import (
	"bytes"
	"testing"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
)

// The Instance page shows the configuration and the footer of its step.
func TestInstance(t *testing.T) {
	t.Parallel()
	v, err := view.New()
	assert.NoError(t, err)

	p := admin(t)
	p.Step = view.StepInstance
	p.Instance = &view.Instance{Root: "/srv/root", DataDir: "/srv/data", Beszel: true}
	var buf bytes.Buffer
	assert.NoError(t, v.Render(&buf, "instance", p))
	body := buf.String()
	assert.Contains(t, body, "/srv/root")
	assert.Contains(t, body, `<a href="/settings/users">Users</a>`)
	assert.NotContains(t, body, "TASK_UI_ADMINS")
	assert.Contains(t, body, `class="setup-trail"`)
	assert.Contains(t, body, `rel="next">Next: Data retention<svg`)
	assert.NotContains(t, body, `rel="prev"`)
}
