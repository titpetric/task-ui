package runner_test

import (
	"context"
	"errors"
	"runtime"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/runner"
)

// settings are instance settings in memory.
type settings map[string]string

func (s settings) Values(context.Context) (map[string]string, error) {
	values := map[string]string{}
	for k, v := range s {
		values[k] = v
	}
	return values, nil
}

func (s settings) Set(_ context.Context, values map[string]string) error {
	for k, v := range values {
		s[k] = v
	}
	return nil
}

var _ runner.SettingValues = settings{}

// brokenSettings fail every read.
type brokenSettings struct{}

func (brokenSettings) Values(context.Context) (map[string]string, error) {
	return nil, errors.New("broken")
}

func (brokenSettings) Set(context.Context, map[string]string) error { return errors.New("broken") }

func TestBacklog(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	values := settings{}
	b := &runner.Backlog{Settings: values, CPUs: 4}
	assert.Equal(t, 80, b.Auto(), "20 runs per CPU")
	assert.Equal(t, 0, b.Configured(ctx))
	assert.Equal(t, 80, b.Length(ctx), "the automatic length is in effect with none set")

	assert.NoError(t, b.Set(ctx, " 12 "))
	assert.Equal(t, "12", values[runner.BacklogSetting])
	assert.Equal(t, 12, b.Configured(ctx))
	assert.Equal(t, 12, b.Length(ctx), "the setting overrides the automatic length")

	for _, bad := range []string{"-1", "abc", "1.5", "100001"} {
		assert.ErrorIs(t, b.Set(ctx, bad), runner.ErrBacklogInvalid, bad)
	}
	assert.Equal(t, 12, b.Length(ctx), "a refused value changes nothing")

	assert.NoError(t, b.Set(ctx, "0"))
	assert.Equal(t, "", values[runner.BacklogSetting], "0 goes back to the automatic length")
	assert.Equal(t, 80, b.Length(ctx))

	values[runner.BacklogSetting] = "garbage"
	assert.Equal(t, 80, b.Length(ctx), "a stored value that does not parse is the automatic length")
}

func TestBacklog_Defaults(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var none *runner.Backlog
	assert.Equal(t, runner.BacklogPerCPU*runtime.NumCPU(), none.Auto())
	assert.Equal(t, none.Auto(), none.Length(ctx))
	assert.Error(t, none.Set(ctx, "5"))

	broken := &runner.Backlog{Settings: brokenSettings{}}
	assert.Equal(t, broken.Auto(), broken.Length(ctx), "settings that fail read as none")
}

func TestParseBacklog(t *testing.T) {
	t.Parallel()
	n, err := runner.ParseBacklog("")
	assert.NoError(t, err)
	assert.Equal(t, 0, n)
	n, err = runner.ParseBacklog(" 100000 ")
	assert.NoError(t, err)
	assert.Equal(t, runner.MaxBacklog, n)
	_, err = runner.ParseBacklog("100001")
	assert.ErrorIs(t, err, runner.ErrBacklogInvalid)
}

func TestManager_Backlog(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	m := runner.New(runner.Options{Backlog: &runner.Backlog{Settings: settings{}, CPUs: 2}})
	state := m.Backlog(ctx)
	assert.Equal(t, runtime.NumCPU(), state.Parallel)
	assert.Equal(t, 40, state.Auto)
	assert.Equal(t, 0, state.Configured)
	assert.Equal(t, 40, state.Length)
	assert.Equal(t, 0, state.Waiting)

	assert.NoError(t, m.SetBacklog(ctx, "7"))
	state = m.Backlog(ctx)
	assert.Equal(t, 7, state.Configured)
	assert.Equal(t, 7, state.Length, "the setting applies at once")
	assert.ErrorIs(t, m.SetBacklog(ctx, "x"), runner.ErrBacklogInvalid)
}
