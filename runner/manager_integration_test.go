//go:build integration

package runner_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/titpetric/task-ui/envfile"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/git"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/keystore"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/recording"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/secret"
	"github.com/titpetric/task-ui/storage"
	_ "modernc.org/sqlite"
)

// fixture is a manager over a private database and temporary folders.
type fixture struct {
	manager *runner.Manager
	runs    *storage.RunStorage
	files   *storage.ArtifactStorage
	secrets *secret.Store
	// reads log which run got which secret.
	reads *storage.SecretReadStorage
	root  string
	// now is what the manager's clock reads; zero is the real time.
	now time.Time
	// workspaces are the workspaces under root the manager knows, and
	// keys the key folder, with an instance SSH key, task_ssh_ed25519.
	workspaces *storage.WorkspaceStorage
	keys       *keystore.Local
	// summaries are Claude's summaries of the runs, recordings their casts.
	summaries  *storage.RunSummaryStorage
	recordings *recording.Store
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	db, err := storage.OpenAndMigrate(context.Background(), ":memory:")
	assert.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	runs := storage.NewRunStorage(db)
	artifacts := storage.NewArtifactStorage(db)
	keys := &keystore.Local{Dir: t.TempDir(), Rows: storage.NewKeyInfoStorage(db), Settings: storage.NewSettingStorage(db), Tmp: t.TempDir()}
	_, err = keys.Generate(context.Background(), keystore.Spec{Name: "task_ssh", Comment: "test"})
	assert.NoError(t, err)
	_, err = keys.Generate(context.Background(), keystore.Spec{Name: "task_encryption", Purpose: keystore.PurposeEncrypt})
	assert.NoError(t, err)
	secrets := secret.New(storage.NewSecretStorage(db), secret.NewVault(keys.Vault(storage.ScopeInstance)))
	f := &fixture{
		runs: runs, files: artifacts, secrets: secrets, reads: storage.NewSecretReadStorage(db), root: t.TempDir(),
		workspaces: storage.NewWorkspaceStorage(db), keys: keys,
		summaries: storage.NewRunSummaryStorage(db), recordings: &recording.Store{Root: t.TempDir()},
	}
	manager := runner.New(runner.Options{
		Now: func() time.Time {
			if f.now.IsZero() {
				return time.Now()
			}
			return f.now
		},
		Runs:         runs,
		Artifacts:    artifacts,
		Summaries:    f.summaries,
		Secrets:      secrets,
		SecretReads:  f.reads,
		Env:          envfile.New(storage.NewEnvFileStorage(db), secrets),
		Git:          &git.Git{},
		Recordings:   f.recordings,
		ArtifactRoot: t.TempDir(),
		Grace:        time.Second,
		KillTimeout:  5 * time.Second,
		Keep:         time.Minute,
		Keys:         keys,
		Root:         f.root,
		Workspaces:   f.workspaces,
		Users:        storage.NewUserStorage(db),
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
	assert.NoError(t, f.secrets.Set(ctx, &model.Secret{Scope: storage.ScopeAll, Name: "REGION", Content: "eu"}))

	req := f.shellJob("env", `echo "$GREETING" "$TASK_UI_FOLDER" "$TASK_UI_JOB" "$TASK_UI_USER"; test -d "$TASK_UI_ARTIFACTS"`, false)
	req.Folder.EnvFiles = []string{".env"}
	session, err := f.manager.Start(ctx, req)
	assert.NoError(t, err)
	out, status := collect(t, session.Attach(true))
	assert.Equal(t, model.StateOK, status.State)
	assert.Contains(t, out, "hello from the folder env run ops")

	// Each secret the run got logs a read: the folder's own GREETING and
	// REGION. Names, never values.
	reads, err := f.reads.ForRun(ctx, session.ID())
	assert.NoError(t, err)
	assert.Len(t, reads, 2)
	assert.Equal(t, "GREETING", reads[0].Name)
	assert.Equal(t, "env", reads[0].Folder)
	assert.Equal(t, "run", reads[0].Job)
	assert.Equal(t, "ops", reads[0].UserName)
	assert.Equal(t, "REGION", reads[1].Name)
	own, err := f.secrets.Get(ctx, model.Secret{Scope: "env", Name: "GREETING"})
	assert.NoError(t, err)
	assert.Equal(t, own.ID, reads[0].SecretID)

	data, err := os.ReadFile(filepath.Join(req.Folder.Dir, ".env"))
	assert.NoError(t, err)
	assert.Contains(t, string(data), "GREETING=\"hello from the folder\"", "the env file was generated")
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

// A request's Env reaches the run; the TASK_UI_ variables still win.
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

// Runs of a host job lock per host, its job, so one runs on every host at
// once; the args are the row's and not the command's, and a timeout
// kills a run that goes past it.
func TestManager_HostTargets(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	root := &folder.Folder{ID: model.Default, Dir: f.root}
	start := func(target, script string, timeout time.Duration) *runner.Session {
		job := root.HostJob(target)
		job.Command = []string{"sh", "-c", script, "sh"}
		session, err := f.manager.Start(ctx, runner.Request{Folder: root, Job: job, User: &model.User{ID: "u1", Name: "ops"}, Target: target, BatchID: "b1", Args: []string{"uptime"}, Timeout: timeout})
		assert.NoError(t, err)
		return session
	}
	slow := start("web1", "sleep 5", 300*time.Millisecond)
	fast := start("web2", `echo "on web2 with $# args"`, 0)

	out, status := collect(t, fast.Attach(true))
	assert.Equal(t, model.StateOK, status.State)
	assert.Contains(t, out, "on web2 with 0 args")
	out, status = collect(t, slow.Attach(true))
	assert.Equal(t, model.StateKilled, status.State)
	assert.Contains(t, out, "the run went past 300ms")

	row, err := f.runs.Get(ctx, fast.ID())
	assert.NoError(t, err)
	assert.Equal(t, "web2", row.Target)
	assert.Equal(t, "b1", row.BatchID)
	assert.Equal(t, []string{"uptime"}, row.ArgList())
}
