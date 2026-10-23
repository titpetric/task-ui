package runner_test

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/titpetric/task-ui/hostpath"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/recording"
	"github.com/titpetric/task-ui/runner"
)

// A run gets its directory and the data directory as the host sees
// them, for a docker run -v the host's daemon resolves; the server's
// own TASK_UI_WORKSPACE, its workspace folder, stays out.
func TestManager_HostPaths(t *testing.T) {
	t.Setenv("TASK_UI_WORKSPACE", "/srv/server-workspaces")
	f := newFixture(t)
	req := f.shellJob("hostpaths", `echo "pwd=[$TASK_UI_HOST_PWD] data=[$TASK_UI_HOST_DATA] ws=[${TASK_UI_WORKSPACE-unset}]"`, false)
	mounts, err := hostpath.Parse(strings.NewReader("564 530 8:1 /mnt/host" + filepath.Dir(req.Folder.Dir) + " " + filepath.Dir(req.Folder.Dir) + " rw - ext4 /dev/sda1 rw\n"))
	assert.NoError(t, err)
	m := runner.New(runner.Options{Runs: f.runs, Recordings: &recording.Store{Root: t.TempDir()}, Grace: time.Second, KillTimeout: 5 * time.Second, Keep: time.Minute, HostPaths: mounts, HostData: "/mnt/host/data"})
	t.Cleanup(func() { m.Shutdown(context.Background()) })
	session, err := m.Start(context.Background(), req)
	assert.NoError(t, err)
	out, status := collect(t, session.Attach(true))
	assert.Equal(t, model.StateOK, status.State)
	assert.Contains(t, out, "pwd=[/mnt/host"+req.Folder.Dir+"] data=[/mnt/host/data] ws=[unset]")
}
