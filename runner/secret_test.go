package runner_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/runner"
)

// A token no running run holds gets no secret.
func TestManager_Secret(t *testing.T) {
	t.Parallel()
	m := runner.New(runner.Options{})
	_, err := m.Secret(context.Background(), "nope", "DEMO")
	assert.ErrorIs(t, err, runner.ErrRunToken)
	_, err = m.Secret(context.Background(), "", "DEMO")
	assert.ErrorIs(t, err, runner.ErrRunToken)
}

// The manager reads and sets the mode through its Environment.
func TestManager_SetParentEnv(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	m := runner.New(runner.Options{})
	assert.Equal(t, runner.EnvAll, m.ParentEnv(ctx), "no setting passes all")
	assert.Error(t, m.SetParentEnv(ctx, runner.EnvNone))

	values := settings{}
	m = runner.New(runner.Options{Environment: &runner.Environment{Settings: values}})
	assert.Equal(t, runner.EnvNone, m.ParentEnv(ctx))
	assert.NoError(t, m.SetParentEnv(ctx, runner.EnvAuto))
	assert.Equal(t, runner.EnvAuto, m.ParentEnv(ctx))
}

// The manager lists the names of its Environment.
func TestManager_EnvNames(t *testing.T) {
	t.Parallel()
	m := runner.New(runner.Options{Environment: &runner.Environment{Environ: func() []string { return []string{"B=1", "A=2"} }}})
	names := m.EnvNames()
	assert.Len(t, names, 2)
	assert.Equal(t, "A", names[0].Name)
}
