package git

import (
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Status is the state of the checkout a directory is in, for the git
// line of a folder page: its branch or detached commit, its upstream,
// the commits ahead of and behind it, and the files it changed.
type Status struct {
	// Checkout is the checkout's directory, a slash path from the root;
	// "" is the root itself.
	Checkout string `json:"checkout"`
	Branch   string `json:"branch,omitempty"`
	// Detached names the commit a checkout off any branch is at.
	Detached string `json:"detached,omitempty"`
	Upstream string `json:"upstream,omitempty"`
	Ahead    int    `json:"ahead"`
	Behind   int    `json:"behind"`
	// Dirty counts the changed and untracked files, the length of
	// Changes.
	Dirty int `json:"dirty"`
	// Changes are the changed and untracked files, as git lists them.
	Changes []Change `json:"changes"`
}

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

// statusTTL is how long a checkout's status stays good.
const statusTTL = 5 * time.Second

// statusCache keeps each checkout's status for statusTTL.
type statusCache struct {
	mu   sync.Mutex
	byID map[string]statusEntry
}

type statusEntry struct {
	status *Status
	at     time.Time
}

var statuses = &statusCache{byID: map[string]statusEntry{}}

// Status returns the status of the checkout dir is in: dir itself, or
// its nearest parent with a .git, no higher than root. Nil when there is
// none, or git cannot read it. It keeps a checkout's answer 5 seconds.
func (g *Git) Status(dir, root string) *Status {
	root = filepath.Clean(root)
	checkout := ""
	for at := filepath.Clean(dir); ; at = filepath.Dir(at) {
		if rel, err := filepath.Rel(root, at); err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			break
		}
		if g.IsRepo(at) {
			checkout = at
			break
		}
		if at == root || filepath.Dir(at) == at {
			break
		}
	}
	if checkout == "" {
		return nil
	}
	key := root + "\x00" + checkout
	statuses.mu.Lock()
	entry, ok := statuses.byID[key]
	statuses.mu.Unlock()
	if ok && time.Since(entry.at) < statusTTL {
		return entry.status
	}
	status := g.status(checkout)
	if status != nil {
		rel, _ := filepath.Rel(root, checkout)
		status.Checkout = strings.TrimPrefix(filepath.ToSlash(rel), ".")
	}
	statuses.mu.Lock()
	statuses.byID[key] = statusEntry{status: status, at: time.Now()}
	statuses.mu.Unlock()
	return status
}

// status reads git status --porcelain=v2 --branch of a checkout.
func (g *Git) status(checkout string) *Status {
	out, err := g.output(checkout, "-c", "safe.directory=*", "-c", "core.quotePath=false", "status", "--porcelain=v2", "--branch")
	if err != nil {
		return nil
	}
	return ParseStatus(out)
}

// ParseStatus reads the output of git status --porcelain=v2 --branch.
func ParseStatus(out string) *Status {
	s := &Status{Changes: []Change{}}
	oid := ""
	for _, line := range strings.Split(out, "\n") {
		if line == "" {
			continue
		}
		if header, ok := strings.CutPrefix(line, "# "); ok {
			key, value, _ := strings.Cut(header, " ")
			switch key {
			case "branch.oid":
				oid = value
			case "branch.head":
				if value != "(detached)" {
					s.Branch = value
				}
			case "branch.upstream":
				s.Upstream = value
			case "branch.ab":
				for _, field := range strings.Fields(value) {
					n, _ := strconv.Atoi(field[1:])
					switch field[0] {
					case '+':
						s.Ahead = n
					case '-':
						s.Behind = n
					}
				}
			}
			continue
		}
		s.Dirty++
		if change, ok := parseChange(line); ok {
			s.Changes = append(s.Changes, change)
		}
	}
	if s.Branch == "" && oid != "" && oid != "(initial)" {
		s.Detached = oid[:min(len(oid), 7)]
	}
	return s
}
