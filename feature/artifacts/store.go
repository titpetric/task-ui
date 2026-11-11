// Package artifactstore keeps the files runs leave behind as one tree: the
// latest version of every path any run wrote, with the earlier versions in
// the history of a git repository around it.
//
// A run that writes job1.txt and a run that writes job2.txt leave both side
// by side; a later run that writes job1.txt again replaces it, and the
// repository keeps the version before. The history grows with every run, so
// the store squashes it to a single commit once it holds more than Keep.
package artifactstore

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/titpetric/task-ui/folder"
)

// ErrNotFound reports a path or a version the store does not hold.
var ErrNotFound = errors.New("not in the artifacts")

// ErrPath refuses a path Write would write: an absolute one, one that
// climbs out of the tree, or one that names a dot file or the repository.
var ErrPath = errors.New("not a path to write in the artifacts: a relative path, no .. and no dot file")

// ErrInTheWay refuses a write where the tree holds a directory at the
// path, or a file where the path needs a directory.
var ErrInTheWay = errors.New("the artifact tree holds a file or a directory in the way")

// commitID matches a full or abbreviated commit name.
var commitID = regexp.MustCompile(`^[0-9a-f]{7,64}$`)

// Store is the artifact tree under Dir, a git repository on branch main.
// The repository starts on the first Ingest.
type Store struct {
	// Dir is the root of the tree, and of the repository.
	Dir string
	// Keep is how many commits the history holds before Squash flattens
	// it; zero keeps every one.
	Keep int

	mu sync.Mutex
}

// Ingest is IngestThen with no callback.
func (s *Store) Ingest(ctx context.Context, src, message, author string) (string, error) {
	return s.IngestThen(ctx, src, message, author, nil)
}

// IngestThen copies every file under src into the tree, over the version a
// path holds, and commits the change under the author's name. It returns
// the commit that holds the files: a new one, or the current one when
// nothing changed. It skips a path element named .git, so a run cannot
// write into the repository. It calls then with the commit while it still
// holds the tree, so the records of the commit are in place before
// Expire, in this process or another, renames it.
func (s *Store) IngestThen(ctx context.Context, src, message, author string, then func(commit string) error) (string, error) {
	unlock, err := s.lock()
	if err != nil {
		return "", err
	}
	defer unlock()
	if err := s.init(ctx); err != nil {
		return "", err
	}
	if err := copyTree(src, s.Dir); err != nil {
		return "", err
	}
	if _, err := s.git(ctx, "add", "-A"); err != nil {
		return "", err
	}
	if _, err := s.git(ctx, "diff", "--cached", "--quiet"); err == nil {
		return s.then(ctx, then)
	}
	if author == "" {
		author = "task-ui"
	}
	if _, err := s.git(ctx, "commit", "-q", "--author", author+" <"+author+"@task-ui>", "-m", message); err != nil {
		return "", err
	}
	return s.then(ctx, then)
}

// then calls then, when given, with the current commit, and returns it.
func (s *Store) then(ctx context.Context, then func(commit string) error) (string, error) {
	head := s.head(ctx)
	if then != nil {
		if err := then(head); err != nil {
			return "", err
		}
	}
	return head, nil
}

// Remove takes a file or a directory out of the tree and commits that
// under the author's name; the history keeps its versions. It returns
// the new commit. A path the tree does not hold is ErrNotFound.
func (s *Store) Remove(ctx context.Context, rel, author string) (string, error) {
	clean, err := cleanPath(rel)
	if err != nil || clean == "" {
		return "", ErrNotFound
	}
	unlock, err := s.lock()
	if err != nil {
		return "", err
	}
	defer unlock()
	if !s.exists() {
		return "", ErrNotFound
	}
	if _, err := os.Lstat(filepath.Join(s.Dir, filepath.FromSlash(clean))); err != nil {
		return "", ErrNotFound
	}
	if err := os.RemoveAll(filepath.Join(s.Dir, filepath.FromSlash(clean))); err != nil {
		return "", err
	}
	if _, err := s.git(ctx, "add", "-A", "--", clean); err != nil {
		return "", err
	}
	if _, err := s.git(ctx, "diff", "--cached", "--quiet"); err == nil {
		return s.head(ctx), nil
	}
	if author == "" {
		author = "task-ui"
	}
	if _, err := s.git(ctx, "commit", "-q", "--author", author+" <"+author+"@task-ui>", "-m", "Delete "+clean); err != nil {
		return "", err
	}
	return s.head(ctx), nil
}

// Squash is SquashThen with no callback.
func (s *Store) Squash(ctx context.Context) (string, error) {
	return s.SquashThen(ctx, nil)
}

// SquashThen flattens the history to one commit of the current tree when it
// holds more than Keep commits, and drops the versions only the history
// held. It returns the new commit, or "" when the history is within Keep,
// and calls then with the new commit while it still holds the tree.
func (s *Store) SquashThen(ctx context.Context, then func(head string) error) (string, error) {
	unlock, err := s.lock()
	if err != nil {
		return "", err
	}
	defer unlock()
	if s.Keep <= 0 || !s.exists() {
		return "", nil
	}
	out, err := s.git(ctx, "rev-list", "--count", "HEAD")
	if err != nil {
		return "", nil
	}
	count, _ := strconv.Atoi(out)
	if count <= s.Keep {
		return "", nil
	}
	tree, err := s.git(ctx, "rev-parse", "HEAD^{tree}")
	if err != nil {
		return "", err
	}
	head, err := s.git(ctx, "commit-tree", tree, "-m", fmt.Sprintf("The artifacts, the history of %d versions squashed", count))
	if err != nil {
		return "", err
	}
	if err := s.replace(ctx, head, ""); err != nil {
		return "", err
	}
	if then != nil {
		if err := then(head); err != nil {
			return "", err
		}
	}
	return head, nil
}

// replace points main at a new head, when it still names old, if the
// caller names one, and drops what only the history before held.
func (s *Store) replace(ctx context.Context, head, old string) error {
	if err := s.point(ctx, head, old); err != nil {
		return err
	}
	return s.prune(ctx)
}

// point points main at a new head, when it still names old, if the
// caller names one.
func (s *Store) point(ctx context.Context, head, old string) error {
	ref := []string{"update-ref", "refs/heads/main", head}
	if old != "" {
		ref = append(ref, old)
	}
	_, err := s.git(ctx, ref...)
	return err
}

// prune drops what no commit of main holds.
func (s *Store) prune(ctx context.Context) error {
	for _, args := range [][]string{
		{"reflog", "expire", "--expire=now", "--all"},
		{"gc", "--prune=now", "--quiet"},
	} {
		if _, err := s.git(ctx, args...); err != nil {
			return err
		}
	}
	return nil
}

// Show returns a path's file as a commit holds it.
func (s *Store) Show(ctx context.Context, commit, rel string) ([]byte, error) {
	if !commitID.MatchString(commit) {
		return nil, ErrNotFound
	}
	clean, err := cleanPath(rel)
	if err != nil || clean == "" {
		return nil, ErrNotFound
	}
	unlock, err := s.lock()
	if err != nil {
		return nil, err
	}
	defer unlock()
	if !s.exists() {
		return nil, ErrNotFound
	}
	cmd := s.command(ctx, "cat-file", "blob", commit+":"+clean)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, ErrNotFound
	}
	return out.Bytes(), nil
}

// Path returns where a path of the tree is on disk. A path that leaves the
// tree or names the repository is ErrNotFound.
func (s *Store) Path(rel string) (string, error) {
	clean, err := cleanPath(rel)
	if err != nil {
		return "", err
	}
	return filepath.Join(s.Dir, filepath.FromSlash(clean)), nil
}

// List returns the latest versions a directory of the tree holds, directories first.
// It leaves out the repository itself and the other dot entries.
func (s *Store) List(rel string) []folder.Item {
	dir, err := s.Path(rel)
	if err != nil {
		return nil
	}
	items := []folder.Item{}
	for _, item := range folder.Listing(dir) {
		if !strings.HasPrefix(item.Name, ".") {
			items = append(items, item)
		}
	}
	return items
}

// Empty deletes the tree and its history; the next Ingest starts over.
func (s *Store) Empty() error {
	unlock, err := s.lock()
	if err != nil {
		return err
	}
	defer unlock()
	return os.RemoveAll(s.Dir)
}

// lock holds the tree for one change: against the other goroutines of
// this process, and with a lock on the file beside the tree against
// another process, such as task-ui cleanup rewriting the history while
// the server takes a run's files.
func (s *Store) lock() (func(), error) {
	s.mu.Lock()
	if err := os.MkdirAll(filepath.Dir(s.Dir), 0o755); err != nil {
		s.mu.Unlock()
		return nil, err
	}
	file, err := os.OpenFile(s.Dir+".lock", os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		s.mu.Unlock()
		return nil, err
	}
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX); err != nil {
		file.Close()
		s.mu.Unlock()
		return nil, err
	}
	return func() {
		_ = syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
		file.Close()
		s.mu.Unlock()
	}, nil
}

// init makes the repository, once.
func (s *Store) init(ctx context.Context) error {
	if s.exists() {
		return nil
	}
	if err := os.MkdirAll(s.Dir, 0o755); err != nil {
		return err
	}
	_, err := s.git(ctx, "init", "-q", "-b", "main")
	return err
}

// exists reports whether the repository is there.
func (s *Store) exists() bool {
	info, err := os.Stat(filepath.Join(s.Dir, ".git"))
	return err == nil && info.IsDir()
}

// head returns the current commit, or "" before the first one.
func (s *Store) head(ctx context.Context) string {
	out, _ := s.git(ctx, "rev-parse", "HEAD")
	return out
}

// git runs a git command in the tree and returns what it printed.
func (s *Store) git(ctx context.Context, args ...string) (string, error) {
	cmd := s.command(ctx, args...)
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s: %w: %s", args[0], err, strings.TrimSpace(errOut.String()))
	}
	return strings.TrimSpace(out.String()), nil
}

// command is git in the tree with a fixed identity, and none of the
// host's configuration: no hooks, no signing, no pager.
func (s *Store) command(ctx context.Context, args ...string) *exec.Cmd {
	full := append([]string{"-C", s.Dir, "-c", "user.name=task-ui", "-c", "user.email=task-ui@localhost", "-c", "commit.gpgsign=false", "-c", "core.hooksPath=/dev/null"}, args...)
	cmd := exec.CommandContext(ctx, "git", full...)
	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1", "GIT_TERMINAL_PROMPT=0", "GIT_PAGER=cat")
	return cmd
}

// cleanPath returns a slash path inside the tree, "" for its root. A path
// that climbs out of the tree, or has a .git element, is ErrNotFound.
func cleanPath(rel string) (string, error) {
	slashed := strings.ReplaceAll(rel, "\\", "/")
	for _, part := range strings.Split(slashed, "/") {
		if part == ".git" || part == ".." {
			return "", ErrNotFound
		}
	}
	return strings.TrimPrefix(path.Clean("/"+slashed), "/"), nil
}

// copyTree copies the files under src over dst, making the directories,
// replacing a file where a directory goes and the other way round.
func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(p string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.Name() == ".git" {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(src, p)
		if err != nil || rel == "." {
			return err
		}
		target := filepath.Join(dst, rel)
		if entry.IsDir() {
			if info, err := os.Lstat(target); err == nil && !info.IsDir() {
				_ = os.Remove(target)
			}
			return os.MkdirAll(target, 0o755)
		}
		if !entry.Type().IsRegular() {
			return nil
		}
		if info, err := os.Lstat(target); err == nil && info.IsDir() {
			if err := os.RemoveAll(target); err != nil {
				return err
			}
		}
		return copyFile(p, target)
	})
}

// copyFile writes src to dst, through a temporary file beside it.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	tmp, err := os.CreateTemp(filepath.Dir(dst), ".ingest-*")
	if err != nil {
		return err
	}
	if _, err := io.Copy(tmp, in); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmp.Name())
		return err
	}
	return os.Rename(tmp.Name(), dst)
}
