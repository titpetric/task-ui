package server

import (
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func gitRun(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-c", "user.name=t", "-c", "user.email=t@t", "-c", "init.defaultBranch=main"}, args...)...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

// gitRepo makes a bare repository holding a pipeline file and a note, and
// returns its file URL and a working clone to push changes from.
func gitRepo(t *testing.T) (string, string) {
	t.Helper()
	work := filepath.Join(t.TempDir(), "work")
	if err := os.MkdirAll(work, 0o755); err != nil {
		t.Fatal(err)
	}
	gitRun(t, work, "init", "-q")
	mkfile(t, filepath.Join(work, "atkins.yml"), "name: Notes\njobs:\n  note.txt:\n    desc: \"Print the note\"\n")
	mkfile(t, filepath.Join(work, "note.txt"), "v1\n")
	gitRun(t, work, "add", ".")
	gitRun(t, work, "commit", "-q", "-m", "first")
	bare := filepath.Join(t.TempDir(), "notes.git")
	gitRun(t, filepath.Dir(bare), "clone", "-q", "--bare", work, bare)
	gitRun(t, work, "remote", "add", "origin", bare)
	return "file://" + bare, work
}

func TestProjectsPage(t *testing.T) {
	ts, _ := testServer(t)
	admin := login(t, ts, "admin")
	code, body := get(t, admin, ts.URL+"/settings/projects")
	if code != http.StatusOK || !strings.Contains(body, "echo") || !strings.Contains(body, "local folder") {
		t.Fatalf("projects page: %d", code)
	}
	code, body = get(t, admin, ts.URL+"/settings/projects/new?url=x&error=bad")
	if code != http.StatusOK || !strings.Contains(body, "bad") || !strings.Contains(body, `value="x"`) {
		t.Fatalf("new form keeps the input: %d", code)
	}
}

func TestProjectClone(t *testing.T) {
	ts, svc := testServer(t)
	admin := login(t, ts, "admin")
	ops := login(t, ts, "ops")
	repo, work := gitRepo(t)

	if code, _ := post(t, ops, ts.URL+"/settings/projects/clone", url.Values{"url": {repo}}); code != http.StatusForbidden {
		t.Fatalf("runner cloned: %d", code)
	}

	// A clone that cannot work reports git's output and keeps nothing.
	if _, location := post(t, admin, ts.URL+"/settings/projects/clone", url.Values{"url": {"file:///nonexistent/repo.git"}, "id": {"broken"}}); !strings.Contains(location, "error=") {
		t.Fatalf("bad clone: %s", location)
	}
	if _, err := os.Stat(filepath.Join(svc.config.Root, "broken")); err == nil {
		t.Fatal("failed clone left a folder")
	}
	if _, location := post(t, admin, ts.URL+"/settings/projects/clone", url.Values{"url": {"-oops"}}); !strings.Contains(location, "error=") {
		t.Fatal("option-like URL accepted")
	}
	if _, location := post(t, admin, ts.URL+"/settings/projects/clone", url.Values{"url": {repo}, "id": {"bad name"}}); !strings.Contains(location, "error=") {
		t.Fatal("bad folder name accepted")
	}

	// The folder name comes from the URL.
	code, location := post(t, admin, ts.URL+"/settings/projects/clone", url.Values{"url": {repo}})
	if code != http.StatusSeeOther || !strings.Contains(location, "notice=") {
		t.Fatalf("clone: %d %s", code, location)
	}
	code, body := get(t, admin, ts.URL+"/settings/projects")
	if code != http.StatusOK || !strings.Contains(body, "notes") || !strings.Contains(body, "atkins.yml") {
		t.Fatalf("projects page: %d", code)
	}
	code, body = get(t, admin, ts.URL+"/f/notes")
	if code != http.StatusOK || !strings.Contains(body, "Source") || !strings.Contains(body, "first") {
		t.Fatalf("folder page source: %d", code)
	}

	// The run pulls first, so a change pushed upstream is what runs.
	ws, err := dial(t, ts, ops, "/ws/run/notes/note.txt", ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	out, statuses := collect(t, ws, final)
	ws.Close()
	if !strings.Contains(out, "git pull") || !strings.Contains(out, "v1") || statuses[len(statuses)-1].State != "ok" {
		t.Fatalf("first run: %q %+v", out, statuses)
	}
	mkfile(t, filepath.Join(work, "note.txt"), "v2\n")
	gitRun(t, work, "commit", "-q", "-am", "second")
	gitRun(t, work, "push", "-q", "origin", "HEAD:main")
	ws, err = dial(t, ts, ops, "/ws/run/notes/note.txt", ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	out, _ = collect(t, ws, final)
	ws.Close()
	if !strings.Contains(out, "v2") {
		t.Fatalf("run did not pull: %q", out)
	}
}

func TestProjectDelete(t *testing.T) {
	ts, svc := testServer(t)
	admin := login(t, ts, "admin")
	repo, _ := gitRepo(t)
	if _, location := post(t, admin, ts.URL+"/settings/projects/clone", url.Values{"url": {repo}, "id": {"notes"}}); !strings.Contains(location, "notice=") {
		t.Fatalf("clone: %s", location)
	}
	if _, location := post(t, admin, ts.URL+"/settings/projects/delete", url.Values{"id": {"echo"}}); !strings.Contains(location, "error=") {
		t.Fatal("removed a local folder")
	}
	if _, location := post(t, admin, ts.URL+"/settings/projects/delete", url.Values{"id": {"../x"}}); !strings.Contains(location, "error=") {
		t.Fatal("bad name accepted")
	}
	if _, location := post(t, admin, ts.URL+"/settings/projects/delete", url.Values{"id": {"notes"}}); !strings.Contains(location, "notice=") {
		t.Fatalf("delete: %s", location)
	}
	if _, err := os.Stat(filepath.Join(svc.config.Root, "notes")); err == nil {
		t.Fatal("checkout still there")
	}
}

func TestFolderIDFromURL(t *testing.T) {
	for in, want := range map[string]string{
		"git@github.com:org/infra.git": "infra",
		"https://x/o/web-site/":        "web-site",
		"file:///tmp/notes.git":        "notes",
		"plain":                        "plain",
	} {
		if got := folderIDFromURL(in); got != want {
			t.Errorf("%q: got %q, want %q", in, got, want)
		}
	}
}
