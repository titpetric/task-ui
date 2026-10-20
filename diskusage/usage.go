// Package diskusage measures how much space a directory takes, split into
// its files and the git history beside them in .git.
package diskusage

import (
	"io/fs"
	"path/filepath"
)

// Usage is the space a directory takes: Files for everything but .git,
// History for the .git directory at its top.
type Usage struct {
	Files   int64
	History int64
}

// Total is the space the directory takes in all.
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

// Measure walks a directory and adds up the sizes of its files. Entries
// it cannot read count as nothing.
func Measure(dir string) Usage {
	var u Usage
	git := filepath.Join(dir, ".git")
	_ = filepath.WalkDir(dir, func(p string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil || !info.Mode().IsRegular() {
			return nil
		}
		if p == git || len(p) > len(git) && p[:len(git)+1] == git+string(filepath.Separator) {
			u.History += info.Size()
			return nil
		}
		u.Files += info.Size()
		return nil
	})
	return u
}
