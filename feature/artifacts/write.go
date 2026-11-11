package artifactstore

import (
	"context"
	"os"
	"path/filepath"
	"strings"
)

// Write is WriteThen with no callback.
func (s *Store) Write(ctx context.Context, rel string, data []byte, author string) (string, error) {
	return s.WriteThen(ctx, rel, data, author, nil)
}

// WriteThen puts one file into the tree at a path, over the version it
// holds, and commits the change under the author's name, as IngestThen
// does for a run's files. It returns the commit that holds the file and
// calls then with it while it still holds the tree.
func (s *Store) WriteThen(ctx context.Context, rel string, data []byte, author string, then func(commit string) error) (string, error) {
	clean, err := writePath(rel)
	if err != nil {
		return "", err
	}
	unlock, err := s.lock()
	if err != nil {
		return "", err
	}
	defer unlock()
	if err := s.init(ctx); err != nil {
		return "", err
	}
	target := filepath.Join(s.Dir, filepath.FromSlash(clean))
	parts := strings.Split(clean, "/")
	for i := 1; i < len(parts); i++ {
		if info, err := os.Lstat(filepath.Join(s.Dir, filepath.FromSlash(strings.Join(parts[:i], "/")))); err == nil && !info.IsDir() {
			return "", ErrInTheWay
		}
	}
	if info, err := os.Lstat(target); err == nil && !info.Mode().IsRegular() {
		return "", ErrInTheWay
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return "", err
	}
	if err := writeFile(target, data); err != nil {
		return "", err
	}
	if _, err := s.git(ctx, "add", "--", clean); err != nil {
		return "", err
	}
	if _, err := s.git(ctx, "diff", "--cached", "--quiet"); err == nil {
		return s.then(ctx, then)
	}
	if author == "" {
		author = "task-ui"
	}
	if _, err := s.git(ctx, "commit", "-q", "--author", author+" <"+author+"@task-ui>", "-m", "Write "+clean); err != nil {
		return "", err
	}
	return s.then(ctx, then)
}

// writePath returns the slash path Write writes, or ErrPath: a
// relative path of names, none of them . or .., and no dot file or
// directory, which the tree lists as sealed.
func writePath(rel string) (string, error) {
	if rel == "" || strings.HasPrefix(rel, "/") || strings.Contains(rel, "\\") {
		return "", ErrPath
	}
	for _, part := range strings.Split(rel, "/") {
		if part == "" || strings.HasPrefix(part, ".") {
			return "", ErrPath
		}
	}
	return rel, nil
}

// writeFile writes data to dst, through a temporary file beside it.
func writeFile(dst string, data []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(dst), ".ingest-*")
	if err != nil {
		return err
	}
	if _, err := tmp.Write(data); err != nil {
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
