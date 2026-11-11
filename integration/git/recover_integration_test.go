//go:build integration

package git

import (
	"context"
	"io"
	"io/fs"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestGit_RecoverDirty(t *testing.T) {
	l := newLab(t)
	l.commit("a.txt", "two\n", "second")
	l.save(l.work, "a.txt", "mine\n")
	if err := l.pull(); err == nil {
		t.Fatal("the pull of a checkout with a change in the way worked")
	}
	rec := l.recover("local changes would be overwritten by merge")
	if rec.Case != CaseDirty || !rec.Mended {
		t.Fatalf("case %q mended %v, want %q mended", rec.Case, rec.Mended, CaseDirty)
	}
	if len(rec.Parked) != 1 {
		t.Fatalf("parked %v, want one ref", rec.Parked)
	}
	if got := l.run(l.work, "show", rec.Parked[0]+":a.txt"); got != "mine" {
		t.Errorf("the parked ref holds %q, want the change", got)
	}
	if err := l.pull(); err != nil {
		t.Fatalf("the pull after the mend: %v", err)
	}
	if got := l.files(l.work)["a.txt"]; got != "two\n" {
		t.Errorf("a.txt is %q, want the upstream's", got)
	}
}

func TestGit_RecoverUntracked(t *testing.T) {
	l := newLab(t)
	l.commit("c.txt", "theirs\n", "adds c")
	l.save(l.work, "c.txt", "mine\n")
	if err := l.pull(); err == nil {
		t.Fatal("the pull over an untracked file worked")
	}
	rec := l.recover("untracked working tree files would be overwritten by merge")
	if rec.Case != CaseDirty || !rec.Mended {
		t.Fatalf("case %q mended %v, want %q mended", rec.Case, rec.Mended, CaseDirty)
	}
	if len(rec.Parked) != 1 {
		t.Fatalf("parked %v, want one ref", rec.Parked)
	}
	if got := l.run(l.work, "show", rec.Parked[0]+":c.txt"); got != "mine" {
		t.Errorf("the parked ref holds %q, want the untracked file", got)
	}
	if err := l.pull(); err != nil {
		t.Fatalf("the pull after the mend: %v", err)
	}
}

func TestGit_RecoverDiverged(t *testing.T) {
	l := newLab(t)
	l.save(l.work, "mine.txt", "mine\n")
	l.run(l.work, "add", "mine.txt")
	l.run(l.work, "commit", "-m", "local only")
	l.commit("theirs.txt", "theirs\n", "second")
	if err := l.pull(); err == nil {
		t.Fatal("the pull of a diverged checkout worked")
	}
	rec := l.recover("Not possible to fast-forward, aborting")
	if rec.Case != CaseDiverged || !rec.Mended {
		t.Fatalf("case %q mended %v, want %q mended", rec.Case, rec.Mended, CaseDiverged)
	}
	if len(rec.Parked) != 1 {
		t.Fatalf("parked %v, want one ref", rec.Parked)
	}
	if got := l.run(l.work, "log", "-1", "--format=%s", rec.Parked[0]); got != "local only" {
		t.Errorf("the parked ref is at %q, want the local commit", got)
	}
	if err := l.pull(); err != nil {
		t.Fatalf("the pull after the mend: %v", err)
	}
	if _, err := l.try(l.work, "cat-file", "-e", "HEAD:mine.txt"); err == nil {
		t.Error("the local commit is still in the branch")
	}
}

func TestGit_RecoverUpstreamGone(t *testing.T) {
	l := newLab(t)
	l.run(l.remote, "checkout", "-b", "feature")
	l.save(l.remote, "f.txt", "feature\n")
	l.run(l.remote, "add", "f.txt")
	l.run(l.remote, "commit", "-m", "feature work")
	l.run(l.remote, "push", "-u", "origin", "feature")
	l.run(l.work, "fetch", "origin")
	l.run(l.work, "checkout", "feature")
	l.run(l.remote, "checkout", "main")
	l.run(l.remote, "push", "origin", "--delete", "feature")
	if err := l.pull(); err == nil {
		t.Fatal("the pull of a branch the remote dropped worked")
	}
	rec := l.recover("no such ref was fetched")
	if rec.Case != CaseGone || !rec.Mended {
		t.Fatalf("case %q mended %v, want %q mended", rec.Case, rec.Mended, CaseGone)
	}
	if rec.Branch != "main" {
		t.Errorf("the checkout is on %q, want the remote's default", rec.Branch)
	}
	if len(rec.Parked) != 1 {
		t.Fatalf("parked %v, want one ref", rec.Parked)
	}
	if got := l.run(l.work, "log", "-1", "--format=%s", rec.Parked[0]); got != "feature work" {
		t.Errorf("the parked ref is at %q, want the branch's tip", got)
	}
	if err := l.pull(); err != nil {
		t.Fatalf("the pull after the mend: %v", err)
	}
}

func TestGit_RecoverDetached(t *testing.T) {
	l := newLab(t)
	l.commit("a.txt", "two\n", "second")
	l.run(l.work, "checkout", "--detach", "HEAD")
	l.save(l.work, "b.txt", "mine\n")
	if err := l.pull(); err == nil {
		t.Fatal("the pull of a detached checkout worked")
	}
	rec := l.recover("You are not currently on a branch")
	if rec.Case != CaseDetached || !rec.Mended {
		t.Fatalf("case %q mended %v, want %q mended", rec.Case, rec.Mended, CaseDetached)
	}
	if rec.Branch != "main" {
		t.Errorf("the checkout is on %q, want the remote's default", rec.Branch)
	}
	if len(rec.Parked) != 1 {
		t.Fatalf("parked %v, want one ref", rec.Parked)
	}
	if got := l.run(l.work, "show", rec.Parked[0]+":b.txt"); got != "mine" {
		t.Errorf("the parked ref holds %q, want the change", got)
	}
	if err := l.pull(); err != nil {
		t.Fatalf("the pull after the mend: %v", err)
	}
}

func TestGit_RecoverIndexLock(t *testing.T) {
	l := newLab(t)
	l.commit("a.txt", "two\n", "second")
	lock := filepath.Join(l.work, ".git", "index.lock")
	l.save(l.work, ".git/index.lock", "")
	if err := l.pull(); err == nil {
		t.Fatal("the pull under a stale index lock worked")
	}
	rec := l.recover("Unable to create index.lock: File exists")
	if rec.Case != CaseLock || !rec.Mended {
		t.Fatalf("case %q mended %v, want %q mended", rec.Case, rec.Mended, CaseLock)
	}
	if len(rec.Parked) != 0 {
		t.Errorf("parked %v, want nothing", rec.Parked)
	}
	if _, err := os.Stat(lock); err == nil {
		t.Error("the stale lock is still there")
	}
	if err := l.pull(); err != nil {
		t.Fatalf("the pull after the mend: %v", err)
	}
	if got := l.files(l.work)["a.txt"]; got != "two\n" {
		t.Errorf("a.txt is %q, want the upstream's", got)
	}
}

func TestGit_RecoverLockHeld(t *testing.T) {
	l := newLab(t)
	l.commit("a.txt", "two\n", "second")
	lock := filepath.Join(l.work, ".git", "index.lock")
	l.save(l.work, ".git/index.lock", strconv.Itoa(os.Getpid())+"\n")
	rec := l.recover("Unable to create index.lock: File exists")
	if rec.Case != CaseLock || rec.Mended {
		t.Fatalf("case %q mended %v, want %q unmended", rec.Case, rec.Mended, CaseLock)
	}
	if _, err := os.Stat(lock); err != nil {
		t.Error("a held lock was removed")
	}
}

func TestGit_RecoverAuth(t *testing.T) {
	l := newLab(t)
	l.save(l.work, "a.txt", "mine\n")
	l.run(l.work, "remote", "set-url", "origin", "file:///nowhere/"+filepath.Base(l.origin))
	before := l.files(l.work)
	head := l.run(l.work, "rev-parse", "HEAD")
	if err := l.pull(); err == nil {
		t.Fatal("the pull of a remote that is not there worked")
	}
	rec := l.recover("Could not read from remote repository")
	if rec.Case != CaseAuth || rec.Mended {
		t.Fatalf("case %q mended %v, want %q unmended", rec.Case, rec.Mended, CaseAuth)
	}
	if !maps.Equal(before, l.files(l.work)) {
		t.Error("the tree changed under an unreachable remote")
	}
	if got := l.run(l.work, "rev-parse", "HEAD"); got != head {
		t.Errorf("HEAD moved to %s, want %s", got, head)
	}
	if refs := l.parked(); len(refs) != 0 {
		t.Errorf("parked %v, want nothing", refs)
	}
}

func TestGit_RecoverBroken(t *testing.T) {
	l := newLab(t)
	l.save(l.work, "mine.txt", "mine\n")
	if err := os.RemoveAll(filepath.Join(l.work, ".git", "objects")); err != nil {
		t.Fatal(err)
	}
	rec := l.recover("fatal: not a git repository")
	if rec.Case != CaseBroken || !rec.Mended {
		t.Fatalf("case %q mended %v, want %q mended", rec.Case, rec.Mended, CaseBroken)
	}
	if rec.Moved == "" {
		t.Fatal("the broken checkout was not kept")
	}
	if got := l.files(rec.Moved)["mine.txt"]; got != "mine\n" {
		t.Errorf("the kept checkout holds %q of mine.txt, want the file", got)
	}
	if got := l.files(l.work)["a.txt"]; got != "one\n" {
		t.Errorf("the fresh clone holds %q of a.txt, want the remote's", got)
	}
	if err := l.pull(); err != nil {
		t.Fatalf("the pull after the clone: %v", err)
	}
}

// A caller that refuses the last resort gets a checkout left alone and a
// recovery that says it was not mended: a workspace home is where every
// run of it lives, and a fresh clone beside it moves the ground.
func TestGit_RecoverBrokenKept(t *testing.T) {
	l := newLab(t)
	l.save(l.work, "mine.txt", "mine\n")
	if err := os.RemoveAll(filepath.Join(l.work, ".git", "objects")); err != nil {
		t.Fatal(err)
	}
	rec := l.recoverWith("fatal: not a git repository", false)
	if rec.Mended || rec.Moved != "" {
		t.Fatalf("mended %v moved %q, want neither", rec.Mended, rec.Moved)
	}
	if got := l.files(l.work)["mine.txt"]; got != "mine\n" {
		t.Errorf("the checkout holds %q of mine.txt, want it untouched", got)
	}
	if _, err := os.Stat(filepath.Join(l.work, ".git")); err != nil {
		t.Errorf("the checkout lost its .git: %v", err)
	}
}

func TestGit_RecoverTriage(t *testing.T) {
	// A pull that failed over a checkout with nothing wrong in it: the
	// rules classify nothing, which is what a triage answers.
	t.Run("a case the model names", func(t *testing.T) {
		l := newLab(t)
		asked := ""
		l.git.Triage = func(_ context.Context, _, failure, status string) (string, error) {
			asked = failure + status
			return CaseAuth, nil
		}
		before := l.files(l.work)
		rec := l.recover("error: RPC failed")
		if rec.Case != CaseAuth || !rec.Triaged || rec.Mended {
			t.Fatalf("case %q triaged %v mended %v, want %q triaged and unmended", rec.Case, rec.Triaged, rec.Mended, CaseAuth)
		}
		if !strings.Contains(asked, "RPC failed") || !strings.Contains(asked, "branch.head main") {
			t.Errorf("the triage was asked %q, want the failure and the status", asked)
		}
		if !maps.Equal(before, l.files(l.work)) {
			t.Error("the tree changed on a triage that names no mend")
		}
	})
	t.Run("an answer in no case", func(t *testing.T) {
		l := newLab(t)
		l.git.Triage = func(context.Context, string, string, string) (string, error) {
			return "I had a look at the reflog", nil
		}
		rec := l.recover("error: RPC failed")
		if rec.Triaged || rec.Moved == "" {
			t.Fatalf("triaged %v moved %q, want the last resort", rec.Triaged, rec.Moved)
		}
		if rec.Case != CaseUnknown {
			t.Errorf("case %q, want %q", rec.Case, CaseUnknown)
		}
	})
	t.Run("no triage where the rules classify", func(t *testing.T) {
		l := newLab(t)
		l.git.Triage = func(context.Context, string, string, string) (string, error) {
			t.Error("the triage was asked about a case the rules know")
			return "", nil
		}
		l.commit("a.txt", "two\n", "second")
		l.save(l.work, "a.txt", "mine\n")
		if err := l.pull(); err == nil {
			t.Fatal("the pull of a checkout with a change in the way worked")
		}
		if rec := l.recover("local changes would be overwritten"); rec.Case != CaseDirty {
			t.Errorf("case %q, want %q", rec.Case, CaseDirty)
		}
	})
}

// lab is a bare origin, the checkout under test and the checkout the origin's own commits come from.
type lab struct {
	t      *testing.T
	git    *Git
	root   string
	origin string
	url    string
	work   string
	remote string
}

// newLab builds the origin with one commit on main, a checkout of it and the checkout that pushes to it.
func newLab(t *testing.T) *lab {
	t.Helper()
	root := t.TempDir()
	// The machine's own git configuration decides nothing here, neither for
	// the fixture's commands nor for the ones the package runs.
	t.Setenv("HOME", root)
	t.Setenv("GIT_CONFIG_GLOBAL", filepath.Join(root, "gitconfig"))
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("GIT_AUTHOR_NAME", "task-ui")
	t.Setenv("GIT_AUTHOR_EMAIL", "task-ui@test")
	t.Setenv("GIT_COMMITTER_NAME", "task-ui")
	t.Setenv("GIT_COMMITTER_EMAIL", "task-ui@test")
	l := &lab{t: t, git: &Git{}, root: root}
	l.origin = filepath.Join(root, "origin.git")
	l.url = "file://" + l.origin
	l.run(root, "init", "--bare", "-b", "main", l.origin)
	l.remote = filepath.Join(root, "remote")
	l.run(root, "clone", l.url, l.remote)
	l.save(l.remote, "a.txt", "one\n")
	l.save(l.remote, "b.txt", "keep\n")
	l.run(l.remote, "add", ".")
	l.run(l.remote, "commit", "-m", "first")
	l.run(l.remote, "push", "origin", "main")
	l.work = filepath.Join(root, "work")
	if _, err := l.git.Clone(context.Background(), l.url, l.work, ""); err != nil {
		t.Fatal(err)
	}
	return l
}

// run runs git in a directory of the lab and fails the test when git does.
func (l *lab) run(dir string, args ...string) string {
	l.t.Helper()
	out, err := l.try(dir, args...)
	if err != nil {
		l.t.Fatalf("git %s: %v: %s", strings.Join(args, " "), err, out)
	}
	return out
}

// try returns the output of a git command in a directory of the lab, and its error.
func (l *lab) try(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

// save writes a file of a checkout.
func (l *lab) save(dir, name, content string) {
	l.t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		l.t.Fatal(err)
	}
}

// commit adds a commit on the origin's main, what a checkout falls behind by.
func (l *lab) commit(name, content, message string) {
	l.t.Helper()
	l.run(l.remote, "checkout", "main")
	l.run(l.remote, "pull", "--ff-only")
	l.save(l.remote, name, content)
	l.run(l.remote, "add", "--", name)
	l.run(l.remote, "commit", "-m", message)
	l.run(l.remote, "push", "origin", "main")
}

// pull fast forwards the checkout under test the way a run does.
func (l *lab) pull() error {
	return l.git.Pull(context.Background(), l.work, io.Discard)
}

// recover mends the checkout under test, the last resort allowed, and fails the test when the mend itself errors.
func (l *lab) recover(failure string) Recovery {
	l.t.Helper()
	return l.recoverWith(failure, true)
}

// recoverWith mends the checkout under test, clone saying whether the last resort is allowed.
func (l *lab) recoverWith(failure string, clone bool) Recovery {
	l.t.Helper()
	rec, err := l.git.Recover(context.Background(), l.work, failure, clone, io.Discard)
	if err != nil {
		l.t.Fatalf("recover: %v", err)
	}
	return rec
}

// parked returns the refs the recoveries of the checkout left behind.
func (l *lab) parked() []string {
	l.t.Helper()
	out := l.run(l.work, "for-each-ref", "--format=%(refname)", ParkPrefix)
	if out == "" {
		return nil
	}
	return strings.Split(out, "\n")
}

// files returns every file of the checkout outside .git, by path and content.
func (l *lab) files(dir string) map[string]string {
	l.t.Helper()
	tree := map[string]string{}
	err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		if rel == ".git" {
			return fs.SkipDir
		}
		if entry.IsDir() {
			return nil
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		tree[rel] = string(body)
		return nil
	})
	if err != nil {
		l.t.Fatal(err)
	}
	return tree
}
