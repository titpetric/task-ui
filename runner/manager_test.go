package runner_test

import (
	"context"
	"errors"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/runner"
)

func TestNew(t *testing.T) {
	t.Parallel()
	m := runner.New(runner.Options{})
	assert.NotNil(t, m)
	assert.Len(t, m.List(), 0)
	_, ok := m.Get("missing")
	assert.False(t, ok)
	_, ok = m.Active("missing")
	assert.False(t, ok)
	assert.ErrorIs(t, m.Stop("missing"), runner.ErrNotFound)
	assert.ErrorIs(t, m.Kill("missing"), runner.ErrNotFound)
}

func TestErrors(t *testing.T) {
	t.Parallel()
	for _, err := range []error{runner.ErrBusy, runner.ErrNotRunning, runner.ErrNotInteractive, runner.ErrNotFound} {
		assert.Error(t, err)
		assert.False(t, errors.Is(err, context.Canceled))
	}
}

// fixedIntegrations adds one variable to the runs of one user.
type fixedIntegrations struct{ user string }

func (f fixedIntegrations) Env(_ context.Context, user, _ string) (map[string]string, error) {
	if user != f.user {
		return nil, nil
	}
	return map[string]string{"CLAUDE_CODE_OAUTH_TOKEN": "sk-1"}, nil
}
