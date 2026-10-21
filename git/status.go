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
// the commits ahead of and behind it, and how many files changed.
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
	// Dirty counts the changed and untracked files.
	Dirty int `json:"dirty"`
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
	out, err := g.output(checkout, "-c", "safe.directory=*", "status", "--porcelain=v2", "--branch")
	if err != nil {
		return nil
	}
	return ParseStatus(out)
}

// ParseStatus reads the output of git status --porcelain=v2 --branch.
func ParseStatus(out string) *Status {
	s := &Status{}
	oid := ""
	for _, line := range strings.Split(out, "\n") {
		if line == "" {
			continue
		}
		header, ok := strings.CutPrefix(line, "# ")
		if !ok {
			s.Dirty++
			continue
		}
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
	}
	if s.Branch == "" && oid != "" && oid != "(initial)" {
		s.Detached = oid[:min(len(oid), 7)]
	}
	return s
}
