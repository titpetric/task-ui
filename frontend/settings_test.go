package frontend_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

func TestSettings(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.withKey(t)
	assert.NoError(t, f.opts.Secrets.Set(context.Background(), &model.Secret{Scope: "all", Name: "TOKEN", Content: "x"}))

	body := f.page(t, asAdmin, "/settings")
	assert.Contains(t, body, "<h2>Settings</h2>")
	assert.Contains(t, body, "2 folders")
	assert.Contains(t, body, "1 key ")
	assert.Contains(t, body, "1 secret<")
	assert.Contains(t, body, "4 users")
	assert.Contains(t, body, f.opts.Keys.Dir)
}
