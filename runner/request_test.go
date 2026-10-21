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
// Workspace, or WorkspaceOf's, wins over the folder's.
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
	// An internal job run for a folder takes that folder's workspace.
	req = f.shellJob("internal/prompt", script, false)
	req.WorkspaceOf = "ws/project"
	assert.Contains(t, run(req), "ws=[ws] dir=["+filepath.Join(f.root, "ws")+"]")
}

// A request's Dir is where the process starts, and its job's Dir what
// the row and the lock know: the same job in two directories runs
// twice at once, in one directory once. Args follow the argv after --
// as they are, for a job that takes them, and a job that does not
// refuses them.
func TestManager_DirArgs(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()

	req := f.shellJob("site", `pwd -P; printf '[%s]' "$@"; sleep 1`, false)
	sub := func(rel string) runner.Request {
		dir := filepath.Join(req.Folder.Dir, rel)
		assert.NoError(t, os.MkdirAll(dir, 0o755))
		job := *req.Job
		job.Dir = rel
		next := req
		next.Job, next.Dir = &job, dir
		return next
	}
	images, docs := sub("images"), sub("docs")

	first, err := f.manager.Start(ctx, images)
	assert.NoError(t, err)
	second, err := f.manager.Start(ctx, docs)
	assert.NoError(t, err, "another directory runs at once")
	_, err = f.manager.Start(ctx, images)
	assert.ErrorIs(t, err, runner.ErrBusy)

	out, status := collect(t, first.Attach(true))
	assert.Equal(t, model.StateOK, status.State, out)
	real, err := filepath.EvalSymlinks(images.Dir)
	assert.NoError(t, err)
	assert.Contains(t, out, real)
	_, _ = collect(t, second.Attach(true))
	row, err := f.runs.Get(ctx, first.ID())
	assert.NoError(t, err)
	assert.Equal(t, "images", row.Dir)
	assert.Equal(t, "site/images/run", row.Key())
	assert.Equal(t, "[]", row.Args)

	withArgs := sub("args")
	withArgs.Job.TakesArgs = true
	withArgs.Args = []string{"ps", "-a", "; id", "$HOME"}
	session, err := f.manager.Start(ctx, withArgs)
	assert.NoError(t, err)
	out, status = collect(t, session.Attach(true))
	assert.Equal(t, model.StateOK, status.State, out)
	assert.Contains(t, out, "[ps][-a][; id][$HOME]")
	row, err = f.runs.Get(ctx, session.ID())
	assert.NoError(t, err)
	assert.Equal(t, []string{"ps", "-a", "; id", "$HOME"}, row.ArgList())

	refused := f.shellJob("plain", "true", false)
	refused.Args = []string{"x"}
	_, err = f.manager.Start(ctx, refused)
	assert.ErrorIs(t, err, runner.ErrNoArgs)
}
