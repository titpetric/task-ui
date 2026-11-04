package git

import (
	"strings"
)

// Change is one file a checkout changed or added, as git status
// --porcelain=v2 lists it.
type Change struct {
	// Index and Worktree are the file's status in the index and in the
	// work tree: "M" modified, "A" added, "D" deleted, "R" renamed, "C"
	// copied, "T" type changed, "U" in conflict, "?" untracked, and "."
	// unchanged.
	Index    string `json:"index"`
	Worktree string `json:"worktree"`
	Path     string `json:"path"`
	// From is the path a renamed or copied file came from.
	From string `json:"from,omitempty"`
	// Conflict says the file is unmerged.
	Conflict bool `json:"conflict,omitempty"`
}

// Code returns the file's two status letters, the index's and the work
// tree's, as git status --porcelain=v2 prints them.
func (c Change) Code() string {
	return c.Index + c.Worktree
}

// State names in a word what became of the file, for the table: a
// rename or a copy first, then a deletion, an addition, a change of
// type or a modification, in the index or in the work tree.
func (c Change) State() string {
	switch {
	case c.Conflict:
		return "conflicted"
	case c.Index == "?":
		return "untracked"
	}
	for _, letter := range []byte("RCDATM") {
		if strings.IndexByte(c.Index+c.Worktree, letter) < 0 {
			continue
		}
		switch letter {
		case 'R':
			return "renamed"
		case 'C':
			return "copied"
		case 'D':
			return "deleted"
		case 'A':
			return "added"
		case 'T':
			return "type changed"
		case 'M':
			return "modified"
		}
	}
	return "changed"
}

// parseChange reads a line of git status --porcelain=v2 that is not a
// branch header into the file it names.
func parseChange(line string) (Change, bool) {
	switch line[0] {
	case '?':
		path, ok := strings.CutPrefix(line, "? ")
		return Change{Index: "?", Worktree: "?", Path: path}, ok
	case '1':
		fields := strings.SplitN(line, " ", 9)
		if len(fields) != 9 || len(fields[1]) != 2 {
			return Change{}, false
		}
		return Change{Index: fields[1][:1], Worktree: fields[1][1:], Path: fields[8]}, true
	case '2':
		fields := strings.SplitN(line, " ", 10)
		if len(fields) != 10 || len(fields[1]) != 2 {
			return Change{}, false
		}
		path, from, _ := strings.Cut(fields[9], "\t")
		return Change{Index: fields[1][:1], Worktree: fields[1][1:], Path: path, From: from}, true
	case 'u':
		fields := strings.SplitN(line, " ", 11)
		if len(fields) != 11 || len(fields[1]) != 2 {
			return Change{}, false
		}
		return Change{Index: fields[1][:1], Worktree: fields[1][1:], Path: fields[10], Conflict: true}, true
	}
	return Change{}, false
}
