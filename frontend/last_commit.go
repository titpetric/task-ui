package frontend

import (
	"sync"
	"time"
)

// lastCommitTTL is how long a directory's last commit time stays good:
// the sidebar costs at most one git process per project per minute for
// any number of pages, and a pull shows within the minute.
const lastCommitTTL = time.Minute

// LastCommits answers the time of each project's last commit, from
// Read, git.Git.LastCommit on the server, kept per directory for a
// minute: the sidebar orders a workspace's projects by it on every page.
type LastCommits struct {
	// Read answers for a directory; nil answers zero for every one.
	Read func(dir string) time.Time
	// Now is the clock; tests move it. Nil is time.Now.
	Now func() time.Time

	mu      sync.Mutex
	entries map[string]lastCommitEntry
}

type lastCommitEntry struct {
	when      time.Time
	checkedAt time.Time
}

// Get returns the time of the last commit in dir, from the cache while
// that holds, zero for no checkout or no commit.
func (c *LastCommits) Get(dir string) time.Time {
	now := time.Now
	if c.Now != nil {
		now = c.Now
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if entry, ok := c.entries[dir]; ok && now().Sub(entry.checkedAt) < lastCommitTTL {
		return entry.when
	}
	var when time.Time
	if c.Read != nil {
		when = c.Read(dir)
	}
	if c.entries == nil {
		c.entries = map[string]lastCommitEntry{}
	}
	c.entries[dir] = lastCommitEntry{when: when, checkedAt: now()}
	return when
}
