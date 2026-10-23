//go:build integration

package frontend_test

import (
	"testing"
	"time"

	"github.com/titpetric/task-ui/frontend"
	"github.com/titpetric/task-ui/internal/assert"
)

func TestOptions(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	var opts frontend.Options = f.opts
	assert.NotNil(t, opts.Guard)
	assert.NotNil(t, opts.Users)
	assert.NotNil(t, opts.Projects)
	assert.NotNil(t, opts.Runs)
	assert.NotNil(t, opts.Artifacts)
	assert.NotNil(t, opts.Schedules)
	assert.NotNil(t, opts.Secrets)
	assert.NotNil(t, opts.Runner)
	assert.NotNil(t, opts.Catalogue)
	assert.NotNil(t, opts.Git)
	assert.NotNil(t, opts.Keys)
	assert.NotNil(t, opts.Env)
	assert.Equal(t, f.root, opts.Root)
	assert.Equal(t, "test", opts.Version)
	assert.Equal(t, time.Friday, opts.Now().Weekday())
}
