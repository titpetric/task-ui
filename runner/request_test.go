package runner_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/runner"
)

// A run of a folder inside a workspace names the workspace and its
// directory; HOME stays the server's. A run outside a named one gets the
// default workspace, _ at the root, once there is one, and a request's
// Workspace wins over the folder's.
func TestManager_Workspace(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	assert.NoError(t, f.workspaces.Create(ctx, &model.Workspace{Folder: "ws"}))
	script := `echo "home=[$HOME] ws=[$TASK_UI_WORKSPACE] dir=[$TASK_UI_WORKSPACE_DIR] set=[${TASK_UI_WORKSPACE+y}]"`
	home, err := os.UserHomeDir()
	assert.NoError(t, err)

	run := func(req runner.Request) string {
		t.Helper()
		session, err := f.manager.Start(ctx, req)
		assert.NoError(t, err)
		out, status := collect(t, session.Attach(true))
		assert.Equal(t, model.StateOK, status.State, out)
		return out
	}

	inside := run(f.shellJob("ws/project", script, false))
	assert.Contains(t, inside, "home=["+home+"] ws=[ws] dir=["+filepath.Join(f.root, "ws")+"] set=[y]")

	outside := run(f.shellJob("plain", script, false))
	assert.Contains(t, outside, "home=["+home+"] ws=[] dir=[] set=[]")

	assert.NoError(t, f.workspaces.EnsureDefault(ctx))
	root := run(f.shellJob("plain", script, false))
	assert.Contains(t, root, "home=["+home+"] ws=[_] dir=["+f.root+"] set=[y]")
	assert.Contains(t, run(f.shellJob("ws/project", script, false)), "ws=[ws] dir=["+filepath.Join(f.root, "ws")+"]")
	internal := run(f.shellJob("internal/tools", script, false))
	assert.Contains(t, internal, "ws=[] dir=[] set=[]")

	other := filepath.Join(f.root, "gozo")
	req := f.shellJob("ws/other", script, false)
	req.Workspace = other
	assert.Contains(t, run(req), "ws=[gozo] dir=["+other+"]")
	req = f.shellJob("internal/prompt", script, false)
	req.Workspace = f.root
	assert.Contains(t, run(req), "ws=[_] dir=["+f.root+"]")
}
