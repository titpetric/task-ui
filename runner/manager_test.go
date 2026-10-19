package runner_test

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "modernc.org/sqlite" // the driver storage opens

	"github.com/titpetric/task-ui/envfile"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/git"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/recording"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/secret"
	"github.com/titpetric/task-ui/sshkey"
	"github.com/titpetric/task-ui/storage"
)

// fixture is a manager over a private database and temporary folders.
type fixture struct {
	manager *runner.Manager
	runs    *storage.RunStorage
	files   *storage.ArtifactStorage
	secrets *secret.Store
	root    string
	// policies give the instance key, id_ed25519, to roles' runs.
	policies *storage.SSHKeyPolicyStorage
	// now is what the manager's clock reads; zero is the real time.
	now time.Time
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	db, err := storage.OpenAndMigrate(context.Background(), ":memory:")
	assert.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	runs := storage.NewRunStorage(db)
	artifacts := storage.NewArtifactStorage(db)
	keys := &sshkey.Store{Dir: t.TempDir()}
	_, err = keys.Generate("ed25519", "test")
	assert.NoError(t, err)
	secrets := secret.New(storage.NewSecretStorage(db), secret.NewVault(keys))
	f := &fixture{runs: runs, files: artifacts, secrets: secrets, root: t.TempDir(), policies: storage.NewSSHKeyPolicyStorage(db)}
	manager := runner.New(runner.Options{
		Now: func() time.Time {
			if f.now.IsZero() {
				return time.Now()
			}
			return f.now
		},
		Runs:         runs,
		Artifacts:    artifacts,
		Secrets:      secrets,
		Env:          envfile.New(storage.NewEnvFileStorage(db), secrets),
		Git:          &git.Git{},
		Recordings:   &recording.Store{Root: t.TempDir()},
		ArtifactRoot: t.TempDir(),
		Grace:        time.Second,
		KillTimeout:  5 * time.Second,
		Keep:         time.Minute,
		KeyDir:       keys.Dir,
		KeyPolicies:  f.policies,
	})
	f.manager = manager
	return f
}

// shellJob is a folder with one job that runs a shell script.
func (f *fixture) shellJob(id string, script string, interactive bool) runner.Request {
	dir := filepath.Join(f.root, id)
	_ = os.MkdirAll(dir, 0o755)
	fold := &folder.Folder{ID: id, Dir: dir}
	job := &folder.Job{Folder: id, ID: "run", Kind: "script", Command: []string{"sh", "-c", script}}
	job.Name = "run"
	job.Interactive = interactive
	fold.Jobs = []*folder.Job{job}
	return runner.Request{Folder: fold, Job: job, User: &model.User{ID: "u1", Name: "ops"}, Cols: 80, Rows: 24}
}

// collect reads a subscriber until its status, returning the output seen.
func collect(t *testing.T, sub *runner.Subscriber) (string, runner.Status) {
	t.Helper()
	var out strings.Builder
	deadline := time.After(10 * time.Second)
	for {
		select {
		case event, ok := <-sub.C:
			if !ok {
				t.Fatalf("subscriber closed without a status, output %q", out.String())
			}
			out.Write(event.Output)
			if event.Status != nil {
				return out.String(), *event.Status
			}
		case <-deadline:
			t.Fatalf("no status after 10s, output %q", out.String())
		}
	}
}

func TestNew(t *testing.T) {
	t.Parallel()
	m := runner.New(runner.Options{})
	assert.NotNil(t, m)
	assert.Len(t, m.List(), 0)
	_, ok := m.Get("missing")
	assert.False(t, ok)
	_, ok = m.Active("missing")
	assert.False(t, ok)
	assert.ErrorIs(t, m.Stop("missing"), runner.ErrNotFound)
	assert.ErrorIs(t, m.Kill("missing"), runner.ErrNotFound)
}

func TestManager_Start(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()

	session, err := f.manager.Start(ctx, f.shellJob("echo", "echo hello; exit 0", false))
	assert.NoError(t, err)
	assert.Equal(t, "echo/run", session.Key())
	assert.True(t, model.ValidID(session.ID()))

	sub := session.Attach(true)
	out, status := collect(t, sub)
	assert.Contains(t, out, "hello")
	assert.Equal(t, model.StateOK, status.State)
	assert.True(t, status.Finished())
	<-session.Done()

	run, err := f.runs.Get(ctx, session.ID())
	assert.NoError(t, err)
	assert.Equal(t, model.StateOK, run.State)
	assert.True(t, run.Finished())
	assert.NotEmpty(t, run.CastPath)
	assert.Equal(t, "ops", run.UserName)
	assert.Len(t, f.manager.List(), 0, "released after the run")
	kept, ok := f.manager.Get(session.ID())
	assert.True(t, ok, "a finished session stays attachable")
	assert.Equal(t, session, kept)

	// Attaching after the end replays the output and the status at once.
	late, status := collect(t, kept.Attach(true))
	assert.Contains(t, late, "hello")
	assert.Equal(t, model.StateOK, status.State)

	_, err = f.manager.Start(ctx, runner.Request{})
	assert.Error(t, err, "a request without a job")
}

func TestManager_StartFailed(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	session, err := f.manager.Start(context.Background(), f.shellJob("fail", "exit 3", false))
	assert.NoError(t, err)
	_, status := collect(t, session.Attach(true))
	assert.Equal(t, model.StateFailed, status.State)
	assert.Equal(t, 3, status.ExitCode)

	req := f.shellJob("nocmd", "true", false)
	req.Job.Command = nil
	session, err = f.manager.Start(context.Background(), req)
	assert.NoError(t, err)
	out, status := collect(t, session.Attach(true))
	assert.Equal(t, model.StateFailed, status.State)
	assert.Contains(t, out, "no command")
	assert.Contains(t, status.Message, "no command")
}

func TestManager_StartBusy(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	req := f.shellJob("sleep", "sleep 30", false)

	first, err := f.manager.Start(ctx, req)
	assert.NoError(t, err)
	active, ok := f.manager.Active("sleep/run")
	assert.True(t, ok)
	assert.Equal(t, first, active)
	assert.Len(t, f.manager.List(), 1)

	current, err := f.manager.Start(ctx, req)
	assert.ErrorIs(t, err, runner.ErrBusy)
	assert.Equal(t, first, current)

	// Run force kills the first run and takes over.
	req.Force = true
	second, err := f.manager.Start(ctx, req)
	assert.NoError(t, err)
	assert.NotEqual(t, first.ID(), second.ID())
	_, status := collect(t, first.Attach(true))
	assert.Equal(t, model.StateKilled, status.State)
	assert.Equal(t, int64(1), second.Snapshot().Forced)

	assert.NoError(t, f.manager.Kill(second.ID()))
	_, status = collect(t, second.Attach(true))
	assert.Equal(t, model.StateKilled, status.State)
	<-second.Done()
	assert.Len(t, f.manager.List(), 0)
}

func TestManager_Stop(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	session, err := f.manager.Start(context.Background(), f.shellJob("sleep", "sleep 30", false))
	assert.NoError(t, err)
	sub := session.Attach(true)
	time.Sleep(100 * time.Millisecond)
	assert.NoError(t, f.manager.Stop(session.ID()))
	_, status := collect(t, sub)
	assert.Equal(t, model.StateStopped, status.State)
	assert.Equal(t, 129, status.ExitCode, "128 + SIGHUP, the terminal closing")
}

func TestManager_Shutdown(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	a, err := f.manager.Start(ctx, f.shellJob("a", "sleep 30", false))
	assert.NoError(t, err)
	b, err := f.manager.Start(ctx, f.shellJob("b", "sleep 30", false))
	assert.NoError(t, err)
	f.manager.Shutdown(ctx)
	assert.Equal(t, model.StateKilled, a.Status().State)
	assert.Equal(t, model.StateKilled, b.Status().State)
}

func TestManager_Secrets(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	assert.NoError(t, f.secrets.Set(ctx, &model.Secret{Scope: storage.ScopeAll, Name: "GREETING", Content: "hello from all"}))
	assert.NoError(t, f.secrets.Set(ctx, &model.Secret{Scope: "env", Name: "GREETING", Content: "hello from the folder"}))

	req := f.shellJob("env", `echo "$GREETING" "$TASK_UI_FOLDER" "$TASK_UI_JOB" "$TASK_UI_USER"; test -d "$TASK_UI_ARTIFACTS"`, false)
	req.Folder.EnvFiles = []string{".env"}
	session, err := f.manager.Start(ctx, req)
	assert.NoError(t, err)
	out, status := collect(t, session.Attach(true))
	assert.Equal(t, model.StateOK, status.State)
	assert.Contains(t, out, "hello from the folder env run ops")

	data, err := os.ReadFile(filepath.Join(req.Folder.Dir, ".env"))
	assert.NoError(t, err)
	assert.Contains(t, string(data), "GREETING=\"hello from the folder\"", "the env file was generated")
}

func TestManager_Artifacts(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	session, err := f.manager.Start(ctx, f.shellJob("art", `mkdir -p "$TASK_UI_ARTIFACTS/sub"; echo hi > "$TASK_UI_ARTIFACTS/out.txt"; echo x > "$TASK_UI_ARTIFACTS/sub/data.json"`, false))
	assert.NoError(t, err)
	_, status := collect(t, session.Attach(true))
	assert.Equal(t, model.StateOK, status.State)

	artifacts, err := f.files.List(ctx, storage.ArtifactFilter{RunID: session.ID()})
	assert.NoError(t, err)
	assert.Len(t, artifacts, 2)
	byName := map[string]model.Artifact{}
	for _, artifact := range artifacts {
		byName[artifact.Name] = artifact
	}
	out := byName["out.txt"]
	assert.Equal(t, int64(3), out.ByteSize)
	assert.Contains(t, out.ContentType, "text/plain")
	assert.Equal(t, "application/json", byName["sub/data.json"].ContentType)

	file, err := f.manager.OpenArtifact(&out)
	assert.NoError(t, err)
	defer file.Close()
	assert.FileExists(t, f.manager.ArtifactPath(&out))

	// A run that leaves nothing has no directory.
	empty, err := f.manager.Start(ctx, f.shellJob("none", "true", false))
	assert.NoError(t, err)
	collect(t, empty.Attach(true))
	assert.NoDirExists(t, filepath.Join(f.manager.ArtifactPath(&model.Artifact{Path: empty.ID()})))
}

func TestManager_GitPull(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()

	// A folder that is a checkout of a local repository.
	origin := filepath.Join(f.root, "origin")
	assert.NoError(t, os.MkdirAll(origin, 0o755))
	gitRun(t, origin, "init", "-q")
	assert.NoError(t, os.WriteFile(filepath.Join(origin, "note.txt"), []byte("v1\n"), 0o644))
	gitRun(t, origin, "add", ".")
	gitRun(t, origin, "commit", "-q", "-m", "first")
	bare := filepath.Join(f.root, "origin.git")
	gitRun(t, f.root, "clone", "-q", "--bare", origin, bare)
	gitRun(t, origin, "remote", "add", "origin", bare)
	gitRun(t, f.root, "clone", "-q", bare, filepath.Join(f.root, "notes"))

	req := f.shellJob("notes", "cat note.txt", false)
	assert.NoError(t, os.WriteFile(filepath.Join(origin, "note.txt"), []byte("v2\n"), 0o644))
	gitRun(t, origin, "commit", "-q", "-am", "second")
	gitRun(t, origin, "push", "-q", "origin", "HEAD")

	session, err := f.manager.Start(ctx, req)
	assert.NoError(t, err)
	out, status := collect(t, session.Attach(true))
	assert.Equal(t, model.StateOK, status.State)
	assert.Contains(t, out, "git pull")
	assert.Contains(t, out, "v2", "the run sees what was pushed")
}

func gitRun(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-c", "user.name=t", "-c", "user.email=t@t", "-c", "init.defaultBranch=main"}, args...)...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func TestErrors(t *testing.T) {
	t.Parallel()
	for _, err := range []error{runner.ErrBusy, runner.ErrNotRunning, runner.ErrNotInteractive, runner.ErrNotFound} {
		assert.Error(t, err)
		assert.False(t, errors.Is(err, context.Canceled))
	}
}
