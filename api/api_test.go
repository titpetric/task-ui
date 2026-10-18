package api_test

import (
	"bytes"
	"context"
	"encoding/json"
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
	"github.com/titpetric/task-ui/api"
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
		"artifact": "echo hello > \"$TASK_UI_ARTIFACTS/out.txt\"",
	})
	private := shellFolder(t, root, "private", map[string]string{"deploy": "echo deployed"})
	private.Allow = []string{"admin"}
	edge := shellFolder(t, root, "infra/edge", map[string]string{"deploy": "echo edge"})
	edge.EnvFiles = []string{".env"}

	infra.Job("echo").Interactive = true
	infra.EnvFiles = []string{".env"}

	runs := storage.NewRunStorage(db)
	artifacts := storage.NewArtifactStorage(db)
	recordings := &recording.Store{Root: t.TempDir()}
	manager := runner.New(runner.Options{
		Runs:         runs,
		Artifacts:    artifacts,
		Recordings:   recordings,
		ArtifactRoot: t.TempDir(),
		Grace:        time.Second,
		KillTimeout:  5 * time.Second,
		Keep:         time.Minute,
	})
	t.Cleanup(func() { manager.Shutdown(context.Background()) })

	users := storage.NewUserStorage(db)
	f := &fixture{root: root, users: map[string]*model.User{}, db: db}
	for name, role := range map[string]string{asAdmin: model.RoleAdmin, asOps: model.RoleRunner, asViewer: model.RoleViewer, asLimited: model.RoleRunner} {
		user := &model.User{Name: name, Role: role}
		if name == asLimited {
			user.Folders = model.EncodeList([]string{"infra"})
		}
		assert.NoError(t, users.Create(context.Background(), user))
		f.users[name] = user
	}

	keys := &sshkey.Store{Dir: t.TempDir()}
	secrets := secret.New(storage.NewSecretStorage(db), secret.NewVault(keys))
	cat := catalogue{"infra": infra, "private": private, "infra/edge": edge}
	workflows, workflowRuns := storage.NewWorkflowStorage(db), storage.NewWorkflowRunStorage(db)
	f.opts = api.Options{
		Workflows:    workflows,
		WorkflowRuns: workflowRuns,
		Flows:        workflow.New(workflows, workflowRuns, cat, manager),
		Users:        users,
		Projects:     storage.NewProjectStorage(db),
		Runs:         runs,
		Artifacts:    artifacts,
		Schedules:    storage.NewScheduleStorage(db),
		Secrets:      secrets,
		Runner:       manager,
		Catalogue:    cat,
		Git:          &git.Git{},
		Keys:         keys,
		Env:          envfile.New(storage.NewEnvFileStorage(db), secrets),
		Recordings:   recordings,
		Root:         root,
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

// response is a decoded reply.
type response struct {
	Status  int
	Headers http.Header
	Body    map[string]any
	Raw     []byte
}

// contains reports whether raw holds the text.
func contains(raw []byte, text string) bool {
	return bytes.Contains(raw, []byte(text))
}

// get returns a nested value of the body by path, or nil.
func (r response) get(path ...string) any {
	var v any = r.Body
	for _, key := range path {
		m, ok := v.(map[string]any)
		if !ok {
			return nil
		}
		v = m[key]
	}
	return v
}

// str returns a string value of the body, or "".
func (r response) str(path ...string) string {
	s, _ := r.get(path...).(string)
	return s
}

// list returns a list value of the body.
func (r response) list(key string) []any {
	l, _ := r.Body[key].([]any)
	return l
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

// withKey puts an ed25519 key in the SSH folder, so secrets can be kept.
func (f *fixture) withKey(t *testing.T) {
	t.Helper()
	_, err := f.opts.Keys.Generate("ed25519", "test")
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
