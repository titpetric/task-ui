//go:build integration

package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/artifactstore"
	"github.com/titpetric/task-ui/beszel"
	"github.com/titpetric/task-ui/claude"
	"github.com/titpetric/task-ui/envfile"
	"github.com/titpetric/task-ui/event"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/git"
	"github.com/titpetric/task-ui/hostrun"
	"github.com/titpetric/task-ui/integration"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/keystore"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/note"
	"github.com/titpetric/task-ui/recording"
	"github.com/titpetric/task-ui/retention"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/secret"
	"github.com/titpetric/task-ui/sshprobe"
	"github.com/titpetric/task-ui/storage"
	"github.com/titpetric/task-ui/workflow"
)

// fixture is the API over a private database, with a server in front
// that puts the user a request names on the context, as the guard does.
type fixture struct {
	opts   api.Options
	server *httptest.Server
	users  map[string]*model.User
	root   string
	db     storage.DB
}

// The users of every fixture. The admin and ops see everything; the
// viewer sees everything and runs nothing; the limited runner sees
// only infra.
const (
	asAdmin   = "admin"
	asOps     = "ops"
	asViewer  = "viewer"
	asLimited = "limited"
)

func newFixture(t *testing.T) *fixture {
	t.Helper()
	db, err := storage.OpenAndMigrate(context.Background(), ":memory:")
	assert.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	root := t.TempDir()
	infra := shellFolder(t, root, "infra", map[string]string{
		"plan":     "echo planned",
		"slow":     "sleep 30",
		"fail":     "exit 3",
		"echo":     "read line; echo \"got $line\"",
		"ask":      "read line; echo \"answered $line\"",
		"artifact": "echo hello > \"$TASK_UI_ARTIFACTS/out.txt\"",
	})
	private := shellFolder(t, root, "private", map[string]string{"deploy": "echo deployed"})
	private.Allow = []string{"admin"}
	edge := shellFolder(t, root, "infra/edge", map[string]string{"deploy": "echo edge"})

	infra.Job("echo").Interactive = true

	runs := storage.NewRunStorage(db)
	artifacts := storage.NewArtifactStorage(db)
	recordings := &recording.Store{Root: t.TempDir()}
	tree := &artifactstore.Store{Dir: filepath.Join(t.TempDir(), "tree")}
	manager := runner.New(runner.Options{
		Runs:          runs,
		Artifacts:     artifacts,
		ArtifactStore: tree,
		Recordings:    recordings,
		ArtifactRoot:  t.TempDir(),
		Grace:         time.Second,
		KillTimeout:   5 * time.Second,
		Keep:          time.Minute,
	})
	t.Cleanup(func() { manager.Shutdown(context.Background()) })

	users := storage.NewUserStorage(db)
	f := &fixture{root: root, users: map[string]*model.User{}, db: db}
	for name, role := range map[string]string{asAdmin: model.RoleAdmin, asOps: model.RoleRunner, asViewer: model.RoleViewer, asLimited: model.RoleRunner} {
		user := &model.User{Name: name, Role: role, Mcp: true}
		if name == asLimited {
			user.Folders = model.EncodeList([]string{"infra"})
		}
		assert.NoError(t, users.Create(context.Background(), user))
		f.users[name] = user
	}

	keys := &keystore.Local{Dir: t.TempDir(), Rows: storage.NewKeyInfoStorage(db), Settings: storage.NewSettingStorage(db), Tmp: t.TempDir()}
	_, err = keys.Generate(context.Background(), keystore.Spec{Name: "task_encryption", Purpose: keystore.PurposeEncrypt, Comment: "task-ui instance"})
	assert.NoError(t, err)
	secrets := secret.New(storage.NewSecretStorage(db), secret.NewVault(keys.Vault(storage.ScopeInstance)))
	cat := catalogue{"infra": infra, "private": private, "infra/edge": edge, model.Default: {ID: model.Default, Dir: root}}
	workflows, workflowRuns := storage.NewWorkflowStorage(db), storage.NewWorkflowRunStorage(db)
	flows := workflow.New(workflows, workflowRuns, cat, manager)
	flows.Events = event.NewBus()
	manager.Events = flows.Events
	f.opts = api.Options{
		Workflows:     workflows,
		WorkflowRuns:  workflowRuns,
		Flows:         flows,
		Events:        flows.Events,
		Users:         users,
		Projects:      storage.NewProjectStorage(db),
		Workspaces:    storage.NewWorkspaceStorage(db),
		Runs:          runs,
		Artifacts:     artifacts,
		ArtifactStore: tree,
		Pins:          storage.NewPinStorage(db),
		Schedules:     storage.NewScheduleStorage(db),
		Secrets:       secrets,
		FolderRefs:    storage.NewFolderRefStorage(db),
		Runner:        manager,
		Catalogue:     cat,
		Git:           &git.Git{},
		Keys:          keys,
		KeyReads:      storage.NewKeyReadStorage(db),
		Env:           envfile.New(storage.NewEnvFileStorage(db), secrets),
		Recordings:    recordings,
		Root:          root,
		Integrations:  &integration.Store{Secrets: secrets},
		// A cheap scrypt keeps the exports quick under -race.
		ExportWorkFactor: 10,
		Board:            &note.Board{Notes: storage.NewNoteStorage(db), Users: users, Runs: runs, WorkflowRuns: workflowRuns, Catalogue: cat, Runner: manager, Recordings: recordings, Events: flows.Events, Root: root},
	}
	flows.Prompter = f.opts.Board
	f.opts.Hosts = &beszel.Inventory{TTL: time.Minute, Settings: func(ctx context.Context) (string, string, error) {
		values, err := f.opts.Integrations.Values(ctx, integration.Beszel)
		return values["url"], values["token"], err
	}}
	f.opts.HostRuns = &hostrun.Starter{
		Catalogue: cat, Hosts: f.opts.Hosts, Runner: manager, Timeout: 10 * time.Second,
		SSH: &sshprobe.Prober{Users: []string{"debian"}, Timeout: 2 * time.Second, TTL: time.Hour, Command: fakeSSH(t), Terminal: true},
	}
	flows.Hosts = f.opts.HostRuns
	f.opts.Summarizer = &claude.Summarizer{Command: fakeClaude(t), Timeout: 10 * time.Second}
	f.opts.Summaries = storage.NewRunSummaryStorage(db)
	f.opts.Cleaner = &retention.Cleaner{Values: storage.NewSettingStorage(db), Runs: runs, Artifacts: artifacts, Store: tree, SecretReads: storage.NewSecretReadStorage(db), RecordingsDir: recordings.Root}
	f.opts.SecretReads = storage.NewSecretReadStorage(db)
	f.opts.Tokens = storage.NewUserAuthTokenStorage(db)
	// AddUser stands in for the password file: a new name becomes a runner.
	f.opts.AddUser = func(ctx context.Context, name, password string) error {
		if _, err := users.GetByName(ctx, name); err == nil {
			return api.ErrUserExists
		}
		if len(password) < 8 {
			return errors.New("a password needs at least 8 characters")
		}
		return users.Create(ctx, &model.User{Name: name, Role: model.RoleRunner})
	}
	// RemoveUser stands in for the password file too.
	f.opts.RemoveUser = func(ctx context.Context, name string) error {
		user, err := users.GetByName(ctx, name)
		if err != nil {
			return err
		}
		return users.Delete(ctx, user.ID)
	}
	routes := api.New(f.opts).Routes()
	f.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := f.users[r.Header.Get("X-User")]
		routes.ServeHTTP(w, r.WithContext(access.WithUser(r.Context(), user)))
	}))
	t.Cleanup(f.server.Close)
	return f
}

// shellFolder makes a folder under root with one shell job per script.
func shellFolder(t *testing.T, root, id string, scripts map[string]string) *folder.Folder {
	t.Helper()
	dir := filepath.Join(root, id)
	assert.NoError(t, os.MkdirAll(dir, 0o755))
	f := &folder.Folder{ID: id, Title: id, Dir: dir}
	names := make([]string, 0, len(scripts))
	for name := range scripts {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		job := &folder.Job{Folder: id, ID: name, Kind: "script", Command: []string{"sh", "-c", scripts[name]}}
		job.Name = name
		f.Jobs = append(f.Jobs, job)
	}
	return f
}

// contains reports whether raw holds the text.
func contains(raw []byte, text string) bool {
	return bytes.Contains(raw, []byte(text))
}

// call sends a request as a user and decodes the reply.
func (f *fixture) call(t *testing.T, user, method, path string, body any) response {
	t.Helper()
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		assert.NoError(t, err)
		reader = bytes.NewReader(data)
	}
	req, err := http.NewRequest(method, f.server.URL+path, reader)
	assert.NoError(t, err)
	req.Header.Set("X-User", user)
	return f.do(t, req)
}

// do sends a request and decodes the reply.
func (f *fixture) do(t *testing.T, req *http.Request) response {
	t.Helper()
	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	assert.NoError(t, err)
	result := response{Status: res.StatusCode, Headers: res.Header, Raw: raw}
	if strings.HasPrefix(res.Header.Get("Content-Type"), "application/json") {
		_ = json.Unmarshal(raw, &result.Body)
	}
	return result
}

// withKey puts an SSH key, task_ssh_ed25519, in the key folder beside
// the instance's encryption key.
func (f *fixture) withKey(t *testing.T) {
	t.Helper()
	_, err := f.opts.Keys.Generate(context.Background(), keystore.Spec{Name: "task_ssh", Comment: "test"})
	assert.NoError(t, err)
}

// start starts a job as a user and returns the run ID.
func (f *fixture) start(t *testing.T, user, folderID, job string) string {
	t.Helper()
	res := f.call(t, user, http.MethodPost, "/runs", api.RunRequest{Folder: folderID, Job: job, Cols: 80, Rows: 24})
	assert.Equal(t, http.StatusCreated, res.Status, string(res.Raw))
	return res.str("id")
}

// finished waits for a run to leave the running state and returns it.
func (f *fixture) finished(t *testing.T, id string) *model.Run {
	t.Helper()
	var run *model.Run
	assert.Eventually(t, func() bool {
		var err error
		run, err = f.opts.Runs.Get(context.Background(), id)
		return err == nil && run.Finished()
	}, 10*time.Second, 20*time.Millisecond)
	return run
}

func TestNew(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	var h *api.Handlers = api.New(f.opts)
	assert.NotNil(t, h)
	assert.NotNil(t, h.Routes())
}

func TestRoutes_UnknownPath(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	res := f.call(t, asAdmin, http.MethodGet, "/nothing", nil)
	assert.Equal(t, http.StatusNotFound, res.Status)
}

func TestRoutes_BadJSON(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	req, err := http.NewRequest(http.MethodPost, f.server.URL+"/runs", strings.NewReader("{"))
	assert.NoError(t, err)
	req.Header.Set("X-User", asAdmin)
	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	empty := f.call(t, asAdmin, http.MethodPost, "/runs", nil)
	assert.Equal(t, http.StatusBadRequest, empty.Status)
	assert.Contains(t, empty.str("error"), "JSON body")
}

// fakeSSH is an ssh that signs in as debian to dev1, running the remote
// command locally, and reaches nothing else.
// fakeClaude answers with the token it signed in with and the report's
// first line.
func fakeClaude(t *testing.T) string {
	t.Helper()
	script := filepath.Join(t.TempDir(), "claude")
	body := "#!/bin/sh\necho \"summary by $CLAUDE_CODE_OAUTH_TOKEN of $(head -n1)\"\n"
	assert.NoError(t, os.WriteFile(script, []byte(body), 0o755))
	return script
}

func fakeSSH(t *testing.T) string {
	t.Helper()
	script := filepath.Join(t.TempDir(), "ssh")
	body := "#!/bin/sh\nlogin=$(eval echo \\${$(($# - 1))})\nremote=$(eval echo \\${$#})\ncase \"$login\" in debian@dev1|debian@dev2) exec sh -c \"$remote\";; esac\nexit 255\n"
	assert.NoError(t, os.WriteFile(script, []byte(body), 0o755))
	return script
}
