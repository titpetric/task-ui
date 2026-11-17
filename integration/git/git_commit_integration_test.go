//go:build integration

package git

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// gitlab is a checkout with an identity, and a bare remote it pushes to.
type gitlab struct {
	t    *testing.T
	dir  string
	head string
}

func (l *gitlab) git(dir string, args ...string) string {
	l.t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		l.t.Fatalf("git %s: %s: %s", strings.Join(args, " "), err, out)
	}
	return strings.TrimSpace(string(out))
}

func newGitlab(t *testing.T) *gitlab {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, "tree")
	remote := filepath.Join(root, "remote.git")
	if out, err := exec.Command("git", "init", "-q", "--bare", remote).CombinedOutput(); err != nil {
		t.Fatalf("git init --bare: %s: %s", err, out)
	}
	l := &gitlab{t: t, dir: dir, head: "main"}
	l.git(root, "init", "-q", "-b", l.head, dir)
	l.git(dir, "config", "user.name", "Task UI")
	l.git(dir, "config", "user.email", "task-ui@example.com")
	l.git(dir, "remote", "add", "origin", remote)
	if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("one\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	l.git(dir, "add", "notes.txt")
	l.git(dir, "commit", "-q", "-m", "notes")
	l.git(dir, "push", "-q", "-u", "origin", l.head)
	return l
}

func (l *gitlab) save(name, content string) {
	l.t.Helper()
	if err := os.WriteFile(filepath.Join(l.dir, name), []byte(content), 0o644); err != nil {
		l.t.Fatal(err)
	}
}

// CommitPath writes a commit for the path and says so; a path that holds no
// change writes none.
func TestGit_CommitPath(t *testing.T) {
	t.Parallel()
	l := newGitlab(t)
	g := &Git{}
	ctx := context.Background()

	l.save("notes.txt", "two\n")
	commit, made, err := g.CommitPath(ctx, l.dir, "notes.txt", "edit: notes.txt")
	if err != nil || !made || commit == "" {
		t.Fatalf("CommitPath: commit=%q made=%v err=%v", commit, made, err)
	}
	if head := l.git(l.dir, "rev-parse", "HEAD"); head != commit {
		t.Fatalf("HEAD is %s, want %s", head, commit)
	}
	if subject := l.git(l.dir, "log", "-1", "--format=%s"); subject != "edit: notes.txt" {
		t.Fatalf("subject is %q", subject)
	}

	_, made, err = g.CommitPath(ctx, l.dir, "notes.txt", "edit: notes.txt")
	if err != nil || made {
		t.Fatalf("a path with no change: made=%v err=%v", made, err)
	}

	if _, _, err := g.CommitPath(ctx, t.TempDir(), "notes.txt", "edit: notes.txt"); err != ErrNotRepo {
		t.Fatalf("outside a checkout: %v", err)
	}
}

// Push sends the commit on the branch to the remote the checkout came from.
func TestGit_Push(t *testing.T) {
	t.Parallel()
	l := newGitlab(t)
	g := &Git{}
	ctx := context.Background()

	l.save("notes.txt", "three\n")
	if _, made, err := g.CommitPath(ctx, l.dir, "notes.txt", "edit: notes.txt"); err != nil || !made {
		t.Fatalf("CommitPath: made=%v err=%v", made, err)
	}
	if err := g.Push(ctx, l.dir, io.Discard); err != nil {
		t.Fatalf("Push: %v", err)
	}
	remote := l.git(l.dir, "remote", "get-url", "origin")
	if subject := l.git(remote, "log", "-1", "--format=%s", l.head); subject != "edit: notes.txt" {
		t.Fatalf("the remote's subject is %q", subject)
	}
	if err := g.Push(ctx, t.TempDir(), io.Discard); err != ErrNotRepo {
		t.Fatalf("outside a checkout: %v", err)
	}
}
