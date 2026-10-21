package runner_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/runner"
)

func TestManager_Freeze(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()

	release, err := f.manager.Freeze("web")
	assert.NoError(t, err)
	_, err = f.manager.Freeze("web")
	assert.ErrorIs(t, err, runner.ErrMoving, "one move at a time")
	_, err = f.manager.Start(ctx, f.shellJob("web", "true", false))
	assert.ErrorIs(t, err, runner.ErrMoving)
	_, err = f.manager.Start(ctx, f.shellJob("web/docs", "true", false))
	assert.ErrorIs(t, err, runner.ErrMoving, "a folder inside the project")
	other, err := f.manager.Start(ctx, f.shellJob("website", "true", false))
	assert.NoError(t, err, "a folder whose ID only starts the same")
	<-other.Done()

	release()
	session, err := f.manager.Start(ctx, f.shellJob("web", "true", false))
	assert.NoError(t, err, "released, it runs again")
	<-session.Done()
}

func TestManager_FreezeBusy(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	session, err := f.manager.Start(context.Background(), f.shellJob("web/docs", "sleep 30", false))
	assert.NoError(t, err)

	_, err = f.manager.Freeze("web")
	assert.ErrorIs(t, err, runner.ErrBusy, "a run inside the project holds it")

	assert.NoError(t, f.manager.Kill(session.ID()))
	_, status := collect(t, session.Attach(true))
	assert.Equal(t, model.StateKilled, status.State)
	<-session.Done()
	release, err := f.manager.Freeze("web")
	assert.NoError(t, err)
	release()
}
