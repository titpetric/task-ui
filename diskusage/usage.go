// Package diskusage measures how much space a directory takes, split into
// its files and the git history beside them in .git.
package diskusage

import (
	"bytes"
	"io/fs"
	"os/exec"
	"path/filepath"
	"strings"
)

// Usage is the space a directory takes: Files for everything but .git and
// what git ignores, History for the .git directory at its top, and Ignored
// for what the .gitignore files leave out, such as build output, caches
// and coverage. Sizes are the files' lengths; .git holds its objects
// compressed, so the history can take less than the files it records.
type Usage struct {
	Files   int64
	History int64
	Ignored int64
}

// Total is the space the files and the history take, leaving out the ignored files.
func (u Usage) Total() int64 {
	return u.Files + u.History
}

// FilesPercent is the share of the total the files take, 0 to 100.
func (u Usage) FilesPercent() float64 {
	if u.Total() == 0 {
		return 0
	}
	return float64(u.Files) * 100 / float64(u.Total())
}

// HistoryPercent is the share of the total the history takes, 0 to 100.
func (u Usage) HistoryPercent() float64 {
	if u.Total() == 0 {
		return 0
	}
	return 100 - u.FilesPercent()
}

// Measure walks a directory and adds up the sizes of its files, counting entries it cannot read as nothing.
func Measure(dir string) Usage {
	var u Usage
	git := filepath.Join(dir, ".git")
	ignored := ignoredPaths(dir)
	_ = filepath.WalkDir(dir, func(p string, entry fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if p == git {
			u.History += sum(p)
			return filepath.SkipDir
		}
		if rel, err := filepath.Rel(dir, p); err == nil && ignored[filepath.ToSlash(rel)] {
			if entry.IsDir() {
				u.Ignored += sum(p)
				return filepath.SkipDir
			}
			u.Ignored += size(entry)
			return nil
		}
		u.Files += size(entry)
		return nil
	})
	return u
}

// sum adds up the sizes of the regular files under dir.
func sum(dir string) int64 {
	var total int64
	_ = filepath.WalkDir(dir, func(_ string, entry fs.DirEntry, err error) error {
		if err == nil {
			total += size(entry)
		}
		return nil
	})
	return total
}

// size is the length of a regular file, and nothing for anything else.
func size(entry fs.DirEntry) int64 {
	if entry.IsDir() {
		return 0
	}
	info, err := entry.Info()
	if err != nil || !info.Mode().IsRegular() {
		return 0
	}
	return info.Size()
}

// ignoredPaths asks git for the untracked paths a checkout ignores, a
// whole directory as one entry, relative and slash separated. A directory
// that is no checkout, or git that fails, ignores nothing.
func ignoredPaths(dir string) map[string]bool {
	paths := map[string]bool{}
	if _, err := exec.LookPath("git"); err != nil {
		return paths
	}
	cmd := exec.Command("git", "-c", "safe.directory=*", "-C", dir, "ls-files", "-z", "--others", "--ignored", "--exclude-standard", "--directory")
	out, err := cmd.Output()
	if err != nil {
		return paths
	}
	for _, path := range bytes.Split(out, []byte{0}) {
		if len(path) > 0 {
			paths[strings.TrimSuffix(string(path), "/")] = true
		}
	}
	return paths
}
