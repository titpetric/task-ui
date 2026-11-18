//go:build integration

package app_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/titpetric/task-ui/app"
	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/internal/assert"
)

// newServices returns the services over a private in-memory database. The
// compose project is named so the constructor runs no docker inspect.
func newServices(t *testing.T) *app.AppServices {
	t.Helper()
	dir := t.TempDir()
	cfg := &app.Config{
		Listen: ":3000", Root: dir, DataDir: dir, SSHDir: filepath.Join(dir, "keys"),
		DB: filepath.Join(dir, "task-ui.db"), ComposeProject: "task-ui",
		SessionTTL: 7 * 24 * time.Hour, SessionMax: 30 * 24 * time.Hour,
		Keep: 15 * time.Minute, Grace: 10 * time.Second, ArtifactHistory: 50,
	}
	services, err := app.NewAppServices(cfg, app.NewAppStorage(newDB(t)))
	assert.NoError(t, err)
	return services
}

// Every service an instance shares is constructed. The wiring, the API and
// both front ends read these fields, so a field with no line in the
// constructor is a nil the first page that needs it dereferences.
func TestNewAppServices(t *testing.T) {
	t.Parallel()
	services := newServices(t)

	assert.Empty(t, nilFields(services))
}

// The keystore is one, and the services built over it hold that one: a
// second would read the same folder behind its own lock.
func TestAppServices(t *testing.T) {
	t.Parallel()
	services := newServices(t)

	assert.Equal(t, services.Keystore, services.Git.Keys)
	assert.Equal(t, services.Artifacts, services.Cleaner.Store, "the cleaner prunes the tree the runs commit into")
	assert.Equal(t, "task-ui", services.Instance.ComposeProject)
}

// The instance's identifier is made once and kept: a second start over the
// same database answers the same one, so a restart and a deploy do not rename
// the instance. The start time is this process's and moves with it.
func TestInstanceIDKept(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfg := &app.Config{
		Listen: ":3000", Root: dir, DataDir: dir, SSHDir: filepath.Join(dir, "keys"),
		DB: filepath.Join(dir, "task-ui.db"), ComposeProject: "task-ui",
		SessionTTL: 7 * 24 * time.Hour, SessionMax: 30 * 24 * time.Hour,
		Keep: 15 * time.Minute, Grace: 10 * time.Second, ArtifactHistory: 50,
	}
	db := newDB(t)
	first, err := app.NewAppServices(cfg, app.NewAppStorage(db))
	assert.NoError(t, err)
	second, err := app.NewAppServices(cfg, app.NewAppStorage(db))
	assert.NoError(t, err)

	assert.True(t, model.ValidID(first.Instance.ID))
	assert.Equal(t, first.Instance.ID, second.Instance.ID)
	assert.NotEqual(t, "", first.Instance.Started)
}
