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

	"github.com/titpetric/task-ui/folder"
)

// ErrNotFound reports a path or a version the store does not hold.
var ErrNotFound = errors.New("not in the artifacts")

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

// Ingest copies every file under src into the tree, over the version a
// path holds, and commits the change under the author's name. It returns
// the commit that holds the files: a new one, or the current one when
// nothing changed. Ingest skips a path element named .git, so a run cannot
// write into the repository.
func (s *Store) Ingest(ctx context.Context, src, message, author string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
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
		return s.head(ctx), nil
	}
	if author == "" {
		author = "task-ui"
	}
	if _, err := s.git(ctx, "commit", "-q", "--author", author+" <"+author+"@task-ui>", "-m", message); err != nil {
		return "", err
	}
	return s.head(ctx), nil
}

// Squash flattens the history to one commit of the current tree when it
// holds more than Keep commits, and drops the versions only the history
// held. It returns the new commit, or "" when the history is within Keep.
func (s *Store) Squash(ctx context.Context) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
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
	for _, args := range [][]string{
		{"update-ref", "refs/heads/main", head},
		{"reflog", "expire", "--expire=now", "--all"},
		{"gc", "--prune=now", "--quiet"},
	} {
		if _, err := s.git(ctx, args...); err != nil {
			return "", err
		}
	}
	return head, nil
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
	s.mu.Lock()
	defer s.mu.Unlock()
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

// List returns what a directory of the tree holds, the latest versions,
// directories first. The repository itself is not listed.
func (s *Store) List(rel string) []folder.Item {
	dir, err := s.Path(rel)
	if err != nil {
		return nil
	}
	return folder.Listing(dir)
}

// Empty deletes the tree and its history; the next Ingest starts over.
func (s *Store) Empty() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return os.RemoveAll(s.Dir)
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
