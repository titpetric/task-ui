//go:build integration

package runner_test

import (
	"context"
	"testing"
	"time"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/recording"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/storage"
)

func TestManager_Secrets(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	assert.NoError(t, f.secrets.Set(ctx, &model.Secret{Scope: storage.ScopeAll, Name: "GREETING", Content: "hello from all"}))
	assert.NoError(t, f.secrets.Set(ctx, &model.Secret{Scope: "env", Name: "GREETING", Content: "hello from the folder"}))
	assert.NoError(t, f.secrets.Set(ctx, &model.Secret{Scope: storage.ScopeAll, Name: "REGION", Content: "eu"}))
	assert.NoError(t, f.secrets.Set(ctx, &model.Secret{Scope: storage.ScopeAll, Name: "MODE", Content: "test", IsSecret: 0}))

	req := f.shellJob("env", `echo "$GREETING" "$MODE" "$TASK_UI_FOLDER" "$TASK_UI_JOB" "$TASK_UI_USER"; test -d "$TASK_UI_ARTIFACTS"`, false)
	session, err := f.manager.Start(ctx, req)
	assert.NoError(t, err)
	out, status := collect(t, session.Attach(true))
	assert.Equal(t, model.StateOK, status.State)
	assert.Contains(t, out, "hello from the folder test env run ops", "an env entry reaches the run as a secret does")

	// Each variable the run got logs a read: the folder's own GREETING,
	// MODE and REGION. Names, never values.
	reads, err := f.reads.ForRun(ctx, session.ID())
	assert.NoError(t, err)
	assert.Len(t, reads, 3)
	assert.Equal(t, "GREETING", reads[0].Name)
	assert.Equal(t, "env", reads[0].Folder)
	assert.Equal(t, "run", reads[0].Job)
	assert.Equal(t, "ops", reads[0].UserName)
	assert.Equal(t, "MODE", reads[1].Name)
	assert.Equal(t, "REGION", reads[2].Name)
	own, err := f.secrets.Get(ctx, model.Secret{Scope: "env", Name: "GREETING"})
	assert.NoError(t, err)
	assert.Equal(t, own.ID, reads[0].SecretID)
}

// A request's Env reaches the run; the TASK_UI_ variables keep task-ui's values.
func TestManager_RequestEnv(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	req := f.shellJob("reqenv", `echo "note=[$TASK_UI_NOTE_ID] user=[$TASK_UI_USER]"`, false)
	req.Env = []string{"TASK_UI_NOTE_ID=n1", "TASK_UI_USER=mallory"}
	session, err := f.manager.Start(context.Background(), req)
	assert.NoError(t, err)
	out, status := collect(t, session.Attach(true))
	assert.Equal(t, model.StateOK, status.State)
	assert.Contains(t, out, "note=[n1] user=[ops]")
}

// An integration's variables reach the runs of the users it names alone.
func TestManager_Integrations(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	m := runner.New(runner.Options{Runs: f.runs, Recordings: &recording.Store{Root: t.TempDir()}, Grace: time.Second, KillTimeout: 5 * time.Second, Keep: time.Minute, Integrations: fixedIntegrations{user: "ops"}})
	t.Cleanup(func() { m.Shutdown(context.Background()) })
	for user, want := range map[string]string{"ops": "token=[sk-1]", "guest": "token=[]"} {
		req := f.shellJob("integrations-"+user, `echo "token=[$CLAUDE_CODE_OAUTH_TOKEN]"`, false)
		req.User = &model.User{ID: user, Name: user, Role: model.RoleRunner}
		session, err := m.Start(context.Background(), req)
		assert.NoError(t, err)
		out, _ := collect(t, session.Attach(true))
		assert.Contains(t, out, want, user)
	}
}
