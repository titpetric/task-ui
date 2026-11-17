//go:build integration

package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/app/api"
	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/integration/git"
	"github.com/titpetric/task-ui/internal/access"
)

// apiRepo is a checkout of the catalogue, with a bare remote it pushes to.
type apiRepo struct {
	t       *testing.T
	dir     string
	remote  string
	user    *model.User
	handler http.Handler
}

func apiGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %s: %s", strings.Join(args, " "), err, out)
	}
	return strings.TrimSpace(string(out))
}

// newAPIRepo writes a checkout with a commit, a remote, and the folder page's API in front of it.
func newAPIRepo(t *testing.T) *apiRepo {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, "demo")
	remote := filepath.Join(root, "remote.git")
	if out, err := exec.Command("git", "init", "-q", "--bare", remote).CombinedOutput(); err != nil {
		t.Fatalf("git init --bare: %s: %s", err, out)
	}
	apiGit(t, root, "init", "-q", "-b", "main", dir)
	apiGit(t, dir, "config", "user.name", "Task UI")
	apiGit(t, dir, "config", "user.email", "task-ui@example.com")
	apiGit(t, dir, "remote", "add", "origin", remote)
	for name, body := range map[string]string{"notes.txt": "one\n", ".gitignore": "out/\n", ".env": "TOKEN=1\n"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	apiGit(t, dir, "add", "-A")
	apiGit(t, dir, "commit", "-q", "-m", "notes")
	apiGit(t, dir, "push", "-q", "-u", "origin", "main")

	catalogue := catalogue{"demo": {ID: "demo", Title: "Demo", Dir: dir}}
	user := &model.User{Name: "root", Role: model.RoleAdmin}
	return &apiRepo{
		t:       t,
		dir:     dir,
		remote:  remote,
		user:    user,
		handler: api.New(api.Options{Catalogue: catalogue, Root: root, Git: &git.Git{}}).Routes(),
	}
}

func (r *apiRepo) call(method, path, body string) *httptest.ResponseRecorder {
	r.t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req = req.WithContext(access.WithUser(req.Context(), r.user))
	rec := httptest.NewRecorder()
	r.handler.ServeHTTP(rec, req)
	return rec
}

func TestFileWriteAndPush(t *testing.T) {
	r := newAPIRepo(t)

	body, _ := json.Marshal(map[string]any{"folder": "demo"})
	rec := r.call(http.MethodPost, "/files/push", string(body))
	if rec.Code != http.StatusOK {
		t.Fatalf("push: %d: %s", rec.Code, rec.Body.String())
	}

	rec = r.call(http.MethodPut, "/files/demo/notes.txt", "two\n")
	if rec.Code != http.StatusOK {
		t.Fatalf("write: %d: %s", rec.Code, rec.Body.String())
	}
	answer := map[string]any{}
	if err := json.Unmarshal(rec.Body.Bytes(), &answer); err != nil {
		t.Fatal(err)
	}
	if answer["committed"] != true || answer["commit"] == "" {
		t.Fatalf("the write did not commit: %s", rec.Body.String())
	}
	if data, err := os.ReadFile(filepath.Join(r.dir, "notes.txt")); err != nil || string(data) != "two\n" {
		t.Fatalf("the file holds %q: %v", data, err)
	}
	if subject := apiGit(t, r.dir, "log", "-1", "--format=%s"); subject != "edit: notes.txt" {
		t.Fatalf("the commit subject is %q", subject)
	}

	// A config dot file is editable, an env file is not.
	if rec := r.call(http.MethodPut, "/files/demo/.gitignore", "build/\n"); rec.Code != http.StatusOK {
		t.Fatalf(".gitignore: %d: %s", rec.Code, rec.Body.String())
	}
	if rec := r.call(http.MethodPut, "/files/demo/.env", "TOKEN=2\n"); rec.Code != http.StatusNotFound {
		t.Fatalf(".env: %d: %s", rec.Code, rec.Body.String())
	}

	// A raw read answers the file itself, for the editor to open.
	rec = r.call(http.MethodGet, "/files/demo/.gitignore?raw", "")
	if rec.Code != http.StatusOK || rec.Body.String() != "build/\n" {
		t.Fatalf("raw: %d: %q", rec.Code, rec.Body.String())
	}

	rec = r.call(http.MethodPost, "/files/push", string(body))
	if rec.Code != http.StatusOK {
		t.Fatalf("push after the edits: %d: %s", rec.Code, rec.Body.String())
	}
	if subject := apiGit(t, r.remote, "log", "-1", "--format=%s", "main"); subject != "edit: .gitignore" {
		t.Fatalf("the remote's subject is %q", subject)
	}
}

func TestFileWriteRefusesBinary(t *testing.T) {
	r := newAPIRepo(t)
	rec := r.call(http.MethodPut, "/files/demo/notes.txt", "one\x00two")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("binary: %d: %s", rec.Code, rec.Body.String())
	}
	if rec := r.call(http.MethodPut, "/files/demo/missing.txt", "x"); rec.Code != http.StatusNotFound {
		t.Fatalf("a file that does not exist: %d", rec.Code)
	}
	if rec := r.call(http.MethodPut, "/files/demo/../secret", "x"); rec.Code != http.StatusNotFound {
		t.Fatalf("a path out of the folder: %d", rec.Code)
	}
}
