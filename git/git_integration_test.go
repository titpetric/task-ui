//go:build integration

package git

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func gitCmd(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-c", "user.name=t", "-c", "user.email=t@t", "-c", "init.defaultBranch=main"}, args...)...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

// repo makes a bare repository with one file, and returns its file URL and
// a working clone to push changes from.
func repo(t *testing.T) (string, string) {
	t.Helper()
	work := filepath.Join(t.TempDir(), "work")
	if err := os.MkdirAll(work, 0o755); err != nil {
		t.Fatal(err)
	}
	gitCmd(t, work, "init", "-q")
	if err := os.WriteFile(filepath.Join(work, "note.txt"), []byte("v1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitCmd(t, work, "add", ".")
	gitCmd(t, work, "commit", "-q", "-m", "first")
	bare := filepath.Join(t.TempDir(), "notes.git")
	gitCmd(t, filepath.Dir(bare), "clone", "-q", "--bare", work, bare)
	gitCmd(t, work, "remote", "add", "origin", bare)
	return "file://" + bare, work
}

func TestGit_Clone(t *testing.T) {
	g := &Git{Keys: &fakeKeys{t: t}}
	url, _ := repo(t)
	dir := filepath.Join(t.TempDir(), "notes")

	if _, err := g.Clone(context.Background(), url, dir, ""); err != nil {
		t.Fatal(err)
	}
	if !g.IsRepo(dir) {
		t.Fatal("clone is not a repo")
	}
	if _, err := g.Clone(context.Background(), url, dir, ""); err == nil {
		t.Fatal("cloned over an existing folder")
	}

	broken := filepath.Join(t.TempDir(), "broken")
	if _, err := g.Clone(context.Background(), "file:///nonexistent/repo.git", broken, ""); err == nil {
		t.Fatal("bad clone succeeded")
	}
	if _, err := os.Stat(broken); err == nil {
		t.Fatal("failed clone left a folder")
	}
	if _, err := g.Clone(context.Background(), url, filepath.Join(t.TempDir(), "x"), "-oops"); err == nil {
		t.Fatal("option-like branch accepted")
	}
	if _, err := g.Clone(context.Background(), "nope", filepath.Join(t.TempDir(), "y"), ""); err == nil {
		t.Fatal("invalid url accepted")
	}
}

func TestGit_Pull(t *testing.T) {
	g := &Git{}
	url, work := repo(t)
	dir := filepath.Join(t.TempDir(), "notes")
	if _, err := g.Clone(context.Background(), url, dir, "main"); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(work, "note.txt"), []byte("v2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitCmd(t, work, "commit", "-q", "-am", "second")
	gitCmd(t, work, "push", "-q", "origin", "HEAD:main")

	var out bytes.Buffer
	if err := g.Pull(context.Background(), dir, &out); err != nil {
		t.Fatalf("pull: %v\n%s", err, out.String())
	}
	data, _ := os.ReadFile(filepath.Join(dir, "note.txt"))
	if string(data) != "v2\n" {
		t.Fatalf("pulled content: %q", data)
	}
	if err := g.Pull(context.Background(), t.TempDir(), &out); err == nil {
		t.Fatal("pulled a folder that is not a repo")
	}
}

func TestGit_Info(t *testing.T) {
	g := &Git{}
	url, _ := repo(t)
	dir := filepath.Join(t.TempDir(), "notes")
	if _, err := g.Clone(context.Background(), url, dir, ""); err != nil {
		t.Fatal(err)
	}
	info := g.Info(dir)
	if info == nil || info.Remote != url || info.Branch != "main" || info.Subject != "first" || info.Commit == "" || info.When.IsZero() {
		t.Fatalf("info: %+v", info)
	}
	if g.Info(t.TempDir()) != nil {
		t.Fatal("info of a plain folder")
	}
}

func TestGit_Remove(t *testing.T) {
	g := &Git{}
	url, _ := repo(t)
	dir := filepath.Join(t.TempDir(), "notes")
	if _, err := g.Clone(context.Background(), url, dir, ""); err != nil {
		t.Fatal(err)
	}
	plain := t.TempDir()
	if err := g.Remove(plain); err == nil {
		t.Fatal("removed a folder that is not a checkout")
	}
	if err := g.Remove(dir); err != nil {
		t.Fatal(err)
	}
	if g.IsRepo(dir) {
		t.Fatal("checkout still there")
	}
}

// Checkout finds the checkout a folder is in, no higher than the levels allow.
func TestGit_Checkout(t *testing.T) {
	t.Parallel()
	g := &Git{}
	root := t.TempDir()
	repo := filepath.Join(root, "deploy")
	deep := filepath.Join(repo, "terraform", "apps")
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(deep, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		dir    string
		levels int
		want   string
	}{
		{repo, 0, repo},
		{deep, 2, repo},
		{deep, 1, ""},
		{root, 0, ""},
		{"/", 3, ""},
	} {
		if got := g.Checkout(c.dir, c.levels); got != c.want {
			t.Errorf("Checkout(%q, %d) = %q, want %q", c.dir, c.levels, got, c.want)
		}
	}
}

// A clone or a pull runs ssh with the git key in an agent, named by its
// public half, and the key folder's known_hosts; nothing else does, and
// with no git key ssh is its own.
func TestGit_env(t *testing.T) {
	t.Parallel()
	keys := &fakeKeys{t: t, key: true}
	g := &Git{Keys: keys}
	env, cleanup, err := g.env(context.Background(), true)
	if err != nil {
		t.Fatal(err)
	}
	command := ""
	for _, kv := range env {
		if v, ok := strings.CutPrefix(kv, "GIT_SSH_COMMAND="); ok {
			command = v
		}
	}
	for _, want := range []string{"IdentitiesOnly=yes", "IdentityAgent='" + keys.agent.Socket + "'", "-i '/keys/task_ssh_ed25519.pub'", "UserKnownHostsFile='/keys/known_hosts'"} {
		if !strings.Contains(command, want) {
			t.Fatalf("%q lacks %q", command, want)
		}
	}
	cleanup()
	if _, err := os.Stat(keys.agent.Socket); !os.IsNotExist(err) {
		t.Fatal("the agent outlives the command")
	}
	local, cleanup, _ := g.env(context.Background(), false)
	cleanup()
	for _, kv := range local {
		if strings.HasPrefix(kv, "GIT_SSH_COMMAND=") {
			t.Fatal("a local command gets ssh")
		}
	}
	none, cleanup, _ := (&Git{Keys: &fakeKeys{t: t}}).env(context.Background(), true)
	cleanup()
	for _, kv := range none {
		if strings.Contains(kv, "IdentityAgent") {
			t.Fatal("an agent without a git key")
		}
	}
}
