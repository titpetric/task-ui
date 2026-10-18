package frontend_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	_ "modernc.org/sqlite" // the driver storage opens

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/detector"
	"github.com/titpetric/task-ui/envfile"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/frontend"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/git"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/recording"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/secret"
	"github.com/titpetric/task-ui/sshkey"
	"github.com/titpetric/task-ui/storage"
	"github.com/titpetric/task-ui/workflow"
)

// catalogue is a fixed set of folders.
type catalogue map[string]*folder.Folder

func (c catalogue) Find(id string) *folder.Folder { return c[id] }

func (c catalogue) List() []*folder.Folder {
	ids := make([]string, 0, len(c))
	for id := range c {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	result := make([]*folder.Folder, 0, len(ids))
	for _, id := range ids {
		result = append(result, c[id])
	}
	return result
}

// The users of every fixture: an admin, a runner, a viewer, and a runner
// limited to the infra folder.
const (
	asAdmin   = "admin"
	asOps     = "ops"
	asViewer  = "viewer"
	asLimited = "limited"
)

// fixture is the front end over a private database, served with the real
// guard: a request signs in with the session cookie of a user.
type fixture struct {
	opts   frontend.Options
	server *httptest.Server
	users  map[string]*model.User
	infra  *folder.Folder
	root   string

	passwords *storage.UserAuthPasswordStorage
	emails    *storage.UserEmailStorage
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	db, err := storage.OpenAndMigrate(context.Background(), ":memory:")
	assert.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	root := t.TempDir()
	infra := shellFolder(t, root, "infra", map[string]string{"plan": "echo planned", "slow": "sleep 30", "many": "seq 1 200"})
	infra.Description = "The servers"
	infra.EnvFiles = []string{".env"}
	private := shellFolder(t, root, "private", map[string]string{"deploy": "echo deployed"})
	private.Allow = []string{"admin"}
	edge := shellFolder(t, root, "infra/edge", map[string]string{"deploy": "echo edge"})
	edge.EnvFiles = []string{".env"}

	users := storage.NewUserStorage(db)
	f := &fixture{root: root, users: map[string]*model.User{}, infra: infra, passwords: storage.NewUserAuthPasswordStorage(db), emails: storage.NewUserEmailStorage(db)}
	for name, role := range map[string]string{asAdmin: model.RoleAdmin, asOps: model.RoleRunner, asViewer: model.RoleViewer, asLimited: model.RoleRunner} {
		user := &model.User{Name: name, Role: role}
		if name == asLimited {
			user.Folders = model.EncodeList([]string{"infra"})
		}
		assert.NoError(t, users.Create(context.Background(), user))
		f.users[name] = user
	}

	runs := storage.NewRunStorage(db)
	keys := &sshkey.Store{Dir: t.TempDir()}
	secrets := secret.New(storage.NewSecretStorage(db), secret.NewVault(keys))
	recordings := &recording.Store{Root: t.TempDir()}
	manager := runner.New(runner.Options{Runs: runs, Recordings: recordings, Grace: time.Second, Keep: time.Minute})
	t.Cleanup(func() { manager.Shutdown(context.Background()) })

	cat := catalogue{"infra": infra, "private": private, "infra/edge": edge}
	workflows, workflowRuns := storage.NewWorkflowStorage(db), storage.NewWorkflowRunStorage(db)
	f.opts = frontend.Options{
		Recordings:   recordings,
		Workflows:    workflows,
		WorkflowRuns: workflowRuns,
		Flows:        workflow.New(workflows, workflowRuns, cat, manager),
		Guard:        access.NewGuard(users, access.NewSessions("test-secret", time.Hour)),
		Auth:         access.NewAuthenticator(users, storage.NewUserEmailStorage(db), storage.NewUserAuthPasswordStorage(db)),
		Users:        users,
		Projects:     storage.NewProjectStorage(db),
		Runs:         runs,
		Artifacts:    storage.NewArtifactStorage(db),
		Schedules:    storage.NewScheduleStorage(db),
		Secrets:      secrets,
		Runner:       manager,
		Catalogue:    cat,
		Git:          &git.Git{},
		Keys:         keys,
		Env:          envfile.New(storage.NewEnvFileStorage(db), secrets),
		Root:         root,
		Version:      "test",
		Now:          func() time.Time { return time.Date(2026, time.October, 16, 22, 0, 0, 0, time.Local) },
	}
	h, err := frontend.New(f.opts)
	assert.NoError(t, err)
	f.server = httptest.NewServer(h.Routes())
	t.Cleanup(f.server.Close)
	return f
}

// shellFolder makes a folder under root with one shell job per script.
func shellFolder(t *testing.T, root, id string, scripts map[string]string) *folder.Folder {
	t.Helper()
	dir := filepath.Join(root, id)
	assert.NoError(t, os.MkdirAll(dir, 0o755))
	f := &folder.Folder{ID: id, Title: id, Dir: dir, Sources: []*detector.Source{{Kind: "script", File: "scripts"}}}
	names := make([]string, 0, len(scripts))
	for name := range scripts {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		job := &folder.Job{Folder: id, ID: name, Kind: "script", Source: "scripts", Command: []string{"sh", "-c", scripts[name]}}
		job.Name = name
		f.Jobs = append(f.Jobs, job)
	}
	return f
}

// client returns a client that follows no redirects.
func client() *http.Client {
	return &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
}

// response is a page as a test reads it.
type response struct {
	Status   int
	Body     string
	Location string
	Headers  http.Header
}

// get fetches a page as a user; an empty user is nobody.
func (f *fixture) get(t *testing.T, user, path string) response {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, f.server.URL+path, nil)
	assert.NoError(t, err)
	if user != "" {
		req.AddCookie(&http.Cookie{Name: access.Cookie, Value: f.opts.Guard.Sessions.Issue(user)})
	}
	res, err := client().Do(req)
	assert.NoError(t, err)
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	assert.NoError(t, err)
	return response{Status: res.StatusCode, Body: string(body), Location: res.Header.Get("Location"), Headers: res.Header}
}

// page fetches a page that must render, and returns its body.
func (f *fixture) page(t *testing.T, user, path string) string {
	t.Helper()
	res := f.get(t, user, path)
	assert.Equal(t, http.StatusOK, res.Status, path)
	assert.Contains(t, res.Body, "<!DOCTYPE html>")
	return res.Body
}

// withKey puts an ed25519 key in the SSH folder, so secrets can be kept.
func (f *fixture) withKey(t *testing.T) {
	t.Helper()
	_, err := f.opts.Keys.Generate("ed25519", "test")
	assert.NoError(t, err)
}

// run records a finished run of a job.
func (f *fixture) run(t *testing.T, folderID, job, user, state string) *model.Run {
	t.Helper()
	now := time.Now()
	run := &model.Run{Folder: folderID, Job: job, UserName: user, State: state, StartedAt: &now, FinishedAt: &now}
	assert.NoError(t, f.opts.Runs.Create(context.Background(), run))
	return run
}

// start starts a job through the runner and returns its run.
func (f *fixture) start(t *testing.T, fold *folder.Folder, job string) *runner.Session {
	t.Helper()
	session, err := f.opts.Runner.Start(context.Background(), runner.Request{Folder: fold, Job: fold.Job(job), User: f.users[asOps], Cols: 80, Rows: 24})
	assert.NoError(t, err)
	return session
}

func TestNew(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	var h *frontend.Handlers
	h, err := frontend.New(f.opts)
	assert.NoError(t, err)
	assert.NotNil(t, h.Routes())
}

func TestRoutes_Assets(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	css := f.get(t, "", "/static/task-ui.css")
	assert.Equal(t, http.StatusOK, css.Status)
	assert.True(t, strings.HasPrefix(css.Headers.Get("Content-Type"), "text/css"))
	assert.Equal(t, "public, max-age=3600", css.Headers.Get("Cache-Control"))
	assert.Contains(t, css.Body, "--signal")
	// The URLs carry a hash of the assets, so a deploy that changes them
	// changes the URLs; the build version is "dev" for every go install.
	assert.Contains(t, f.page(t, asViewer, "/"), `href="/static/task-ui.css?v=`+view.AssetVersion()+`"`)
	assert.Contains(t, f.get(t, "", "/login").Body, `src="/static/task-ui.js?v=`+view.AssetVersion()+`"`)

	health := f.get(t, "", "/healthz")
	assert.Equal(t, http.StatusOK, health.Status)
	assert.Equal(t, "ok\n", health.Body)
}

func TestRoutes_Guard(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	// A visitor without a session is sent to sign in, with the way back.
	res := f.get(t, "", "/f/infra")
	assert.Equal(t, http.StatusSeeOther, res.Status)
	assert.Equal(t, "/login?back=%2Ff%2Finfra", res.Location)

	missing := f.get(t, "", "/nothing")
	assert.Equal(t, http.StatusSeeOther, missing.Status)

	// A signed in user gets the pages, and the not found page for the rest.
	assert.Equal(t, http.StatusOK, f.get(t, asViewer, "/").Status)
	gone := f.get(t, asViewer, "/nothing")
	assert.Equal(t, http.StatusNotFound, gone.Status)
	assert.Contains(t, gone.Body, "Not found")
	assert.Contains(t, gone.Body, "/nothing")
}
