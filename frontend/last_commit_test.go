package frontend_test

import (
	"testing"
	"time"

	"github.com/titpetric/task-ui/frontend"
	"github.com/titpetric/task-ui/internal/assert"
)

// The cache reads a directory once a minute: two asks within the
// window cost one read, and the next minute reads again; without a
// reader every directory has no commit.
func TestLastCommits_Get(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, time.October, 25, 12, 0, 0, 0, time.UTC)
	reads := map[string]int{}
	when := map[string]time.Time{"/src/api": now.Add(-time.Hour)}
	c := &frontend.LastCommits{
		Read: func(dir string) time.Time { reads[dir]++; return when[dir] },
		Now:  func() time.Time { return now },
	}
	assert.Equal(t, now.Add(-time.Hour), c.Get("/src/api"))
	assert.Equal(t, now.Add(-time.Hour), c.Get("/src/api"))
	assert.True(t, c.Get("/src/web").IsZero(), "no commit")
	assert.Equal(t, map[string]int{"/src/api": 1, "/src/web": 1}, reads)

	when["/src/web"] = now
	now = now.Add(59 * time.Second)
	assert.True(t, c.Get("/src/web").IsZero(), "still from the cache")
	now = now.Add(2 * time.Second)
	assert.Equal(t, now.Add(-61*time.Second), c.Get("/src/web"), "read again after the minute")
	assert.Equal(t, 2, reads["/src/web"])

	var none frontend.LastCommits
	assert.True(t, none.Get("/anywhere").IsZero())
}
