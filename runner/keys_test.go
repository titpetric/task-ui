package runner_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/sshkey"
)

// A folder inside a checkout pulls the checkout, and the keys of the
// checkout's .ssh/ reach the job through an agent.
func TestManager_CheckoutKeys(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	dir := checkout(t)
	sub := filepath.Join(dir, "terraform")
	assert.NoError(t, os.MkdirAll(sub, 0o755))
	_, err := (&sshkey.Store{Dir: filepath.Join(dir, ".ssh")}).Generate("ed25519", "deploy")
	assert.NoError(t, err)

	fold := &folder.Folder{ID: "repo/terraform", Dir: sub}
	job := &folder.Job{Folder: fold.ID, ID: "plan", Kind: "script", Command: []string{"sh", "-c", `test -S "$SSH_AUTH_SOCK" && echo agent-ok; cat ../README`}}
	job.Name = "plan"
	fold.Jobs = []*folder.Job{job}

	session, err := f.manager.Start(context.Background(), runner.Request{Folder: fold, Job: job, User: &model.User{ID: "u1", Name: "ops"}, Cols: 80, Rows: 24})
	assert.NoError(t, err)
	out, status := collect(t, session.Attach(false))
	assert.Equal(t, model.StateOK, status.State, out)
	assert.Contains(t, out, "git pull --ff-only")
	assert.Contains(t, out, "ssh-agent: id_ed25519 from .ssh/")
	assert.Contains(t, out, "agent-ok")
	assert.Contains(t, out, "hello")
}
