package runner_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/runner"
)

// fakeEnviron is a server environment with a credential in it.
var fakeEnviron = []string{
	"PATH=/usr/bin", "HOME=/root", "LANG=C.UTF-8", "LC_ALL=C", "TZ=Europe/Ljubljana",
	"TASK_UI_URL=https://task.example.com", "TASK_UI_WORKSPACE=/data/workspaces",
	"WORKSPACE_KEY=hunter2", "SECRETSMANAGER_KEY=hunter3", "github_token=hunter4", "DB_PASSWORD=hunter5",
	"TF_CLI_ARGS_init=-backend-config=x", "GOTOOLCHAIN=auto",
}

// Each mode passes its share of the server's environment, and none
// passes what task-ui sets for each run.
func TestParentEnv(t *testing.T) {
	t.Parallel()
	base := []string{"PATH=/usr/bin", "HOME=/root", "LANG=C.UTF-8", "LC_ALL=C", "TZ=Europe/Ljubljana", "TASK_UI_URL=https://task.example.com"}
	assert.Equal(t, base, runner.ParentEnv(fakeEnviron, runner.EnvNone))
	assert.Equal(t, base, runner.ParentEnv(fakeEnviron, ""), "unset is none")
	assert.Equal(t, append(append([]string{}, base...), "TF_CLI_ARGS_init=-backend-config=x", "GOTOOLCHAIN=auto"), runner.ParentEnv(fakeEnviron, runner.EnvAuto))
	all := runner.ParentEnv(fakeEnviron, runner.EnvAll)
	assert.Len(t, all, len(fakeEnviron)-1)
	assert.Contains(t, all, "SECRETSMANAGER_KEY=hunter3")
	assert.NotContains(t, all, "TASK_UI_WORKSPACE=/data/workspaces")
}

// A name with a credential's word, in any case, is sensitive.
func TestSensitive(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"WORKSPACE_KEY", "SECRETSMANAGER_KEY", "github_token", "AWS_SECRET_ACCESS_KEY", "DB_PASSWORD", "PGPASS", "GOOGLE_CREDENTIALS", "PRIVATE_URL"} {
		assert.True(t, runner.Sensitive(name), name)
	}
	for _, name := range []string{"PATH", "HOME", "SSH_AUTH_SOCK", "TASK_UI_URL", "TZ"} {
		assert.False(t, runner.Sensitive(name), name)
	}
	assert.True(t, runner.Base("LC_TIME"))
	assert.False(t, runner.Base("GOTOOLCHAIN"))
}

// None passes a base name, Auto any name, All a sensitive one too.
func TestForwarded(t *testing.T) {
	t.Parallel()
	assert.True(t, runner.Forwarded("PATH", runner.EnvNone))
	assert.False(t, runner.Forwarded("GOTOOLCHAIN", runner.EnvNone))
	assert.True(t, runner.Forwarded("GOTOOLCHAIN", runner.EnvAuto))
	assert.False(t, runner.Forwarded("WORKSPACE_KEY", runner.EnvAuto))
	assert.True(t, runner.Forwarded("WORKSPACE_KEY", runner.EnvAll))
	assert.False(t, runner.Forwarded("TASK_UI_TOKEN", runner.EnvNone), "a base name that is sensitive stays out")
}

// The setting stores a mode, refuses another, and lists the names with
// what each mode passes, never a value.
func TestEnvironment(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	values := settings{}
	env := &runner.Environment{Settings: values, Environ: func() []string { return fakeEnviron }}
	assert.Equal(t, runner.EnvNone, env.Mode(ctx), "a fresh instance passes none")
	assert.NotContains(t, env.Env(ctx), "GOTOOLCHAIN=auto")

	assert.ErrorIs(t, env.Set(ctx, "some"), runner.ErrEnvironmentInvalid)
	assert.NoError(t, env.Set(ctx, runner.EnvAuto))
	assert.Equal(t, runner.EnvAuto, values[runner.EnvironmentSetting])
	assert.Equal(t, runner.EnvAuto, env.Mode(ctx))
	assert.Contains(t, env.Env(ctx), "GOTOOLCHAIN=auto")

	names := env.Names()
	assert.Len(t, names, len(fakeEnviron)-1)
	assert.Equal(t, "DB_PASSWORD", names[0].Name, "by name")
	data, err := json.Marshal(names)
	assert.NoError(t, err)
	for _, value := range []string{"hunter2", "hunter3", "hunter4", "hunter5", "/usr/bin"} {
		assert.NotContains(t, string(data), value)
	}
	for _, n := range names {
		if n.Name == "WORKSPACE_KEY" {
			assert.True(t, n.Sensitive)
			assert.False(t, n.None)
			assert.False(t, n.Auto)
		}
	}

	var unset *runner.Environment
	assert.Equal(t, runner.EnvAll, unset.Mode(ctx), "a manager without the setting passes all")
	assert.Error(t, unset.Set(ctx, runner.EnvNone))
}
