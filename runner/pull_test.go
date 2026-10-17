package runner_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/runner"
)

// checkout makes a repository with one commit and returns a clone of it.
func checkout(t *testing.T) string {
	t.Helper()
	base := t.TempDir()
	origin := filepath.Join(base, "origin")
	assert.NoError(t, os.MkdirAll(origin, 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(origin, "README"), []byte("hello\n"), 0o644))
	run := func(dir string, args ...string) {
		cmd := exec.Command("git", append([]string{"-c", "user.name=test", "-c", "user.email=test@example.com"}, args...)...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		assert.NoError(t, err, string(out))
	}
	run(origin, "init", "-q", "-b", "main")
	run(origin, "add", ".")
	run(origin, "commit", "-q", "-m", "init")
	run(base, "clone", "-q", origin, "clone")
	return filepath.Join(base, "clone")
}

// Two jobs of one checkout started together both pull, one after the
// other, and both run.
func TestManager_PullOneAtATime(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	dir := checkout(t)
	fold := &folder.Folder{ID: "repo", Dir: dir}
	user := &model.User{ID: "u1", Name: "ops"}
	var jobs []*folder.Job
	for _, name := range []string{"one", "two", "three"} {
		job := &folder.Job{Folder: "repo", ID: name, Kind: "script", Command: []string{"sh", "-c", "cat README"}}
		job.Name = name
		jobs = append(jobs, job)
	}
	fold.Jobs = jobs

	var sessions []*runner.Session
	for _, job := range jobs {
		session, err := f.manager.Start(context.Background(), runner.Request{Folder: fold, Job: job, User: user, Cols: 80, Rows: 24})
		assert.NoError(t, err)
		sessions = append(sessions, session)
	}
	for _, session := range sessions {
		out, status := collect(t, session.Attach(false))
		assert.Equal(t, model.StateOK, status.State, out)
		assert.Contains(t, out, "git pull --ff-only")
		assert.Contains(t, out, "hello")
	}
}

// A folder that is not a checkout is not pulled.
func TestManager_NoPull(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	session, err := f.manager.Start(context.Background(), f.shellJob("plain", "echo plain", false))
	assert.NoError(t, err)
	out, status := collect(t, session.Attach(false))
	assert.Equal(t, model.StateOK, status.State)
	assert.NotContains(t, out, "git pull")
}
