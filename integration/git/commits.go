package git

import (
	"sync"
	"time"
)

// commitTTL is how long a directory's last commit time stays good: the
// sidebar costs at most one git process per project per minute for any
// number of pages, and a pull shows within the minute.
const commitTTL = time.Minute

// Commits answers the time of each project's last commit, from Read,
// Git.LastCommit on the server, kept per directory for a minute: both
// front ends order a workspace's projects by it on every page.
type Commits struct {
	// Read answers for a directory; nil answers zero for every one.
	Read func(dir string) time.Time
	// Now is the clock; tests move it. Nil is time.Now.
	Now func() time.Time

	mu      sync.Mutex
	entries map[string]commitEntry
}

type commitEntry struct {
	when      time.Time
	checkedAt time.Time
}

// commits is the cache Git.Commit reads, one for the process the way
// statuses is, so a project costs one git process a minute however many
// pages ask and whichever front end they came from.
var commits = &Commits{}

// Get returns the time of the last commit in dir, from the cache while
// that holds, zero for no checkout or no commit.
func (c *Commits) Get(dir string) time.Time {
	return c.read(dir, c.Read)
}

// read answers dir from the cache while that holds, and otherwise from
// from, which is Read for a cache of its own and Git.LastCommit for the
// process's.
func (c *Commits) read(dir string, from func(dir string) time.Time) time.Time {
	now := time.Now
	if c.Now != nil {
		now = c.Now
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if entry, ok := c.entries[dir]; ok && now().Sub(entry.checkedAt) < commitTTL {
		return entry.when
	}
	var when time.Time
	if from != nil {
		when = from(dir)
	}
	if c.entries == nil {
		c.entries = map[string]commitEntry{}
	}
	c.entries[dir] = commitEntry{when: when, checkedAt: now()}
	return when
}
