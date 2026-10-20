package view_test

import (
	"testing"
	"time"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// Recents fall under the hour, then the day, they started in.
func TestGroupRecents(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)
	at := func(ago time.Duration) view.Recent {
		started := now.Add(-ago)
		return view.Recent{Run: &view.Run{Run: model.Run{StartedAt: &started}}}
	}
	groups := view.GroupRecents(now, []view.Recent{at(time.Minute), at(50 * time.Minute), at(70 * time.Minute), at(3 * time.Hour), at(50 * time.Hour)})
	assert.Len(t, groups, 4)
	assert.Equal(t, "In the last hour", groups[0].Label)
	assert.Len(t, groups[0].Entries, 2)
	assert.Equal(t, "1 hour ago", groups[1].Label)
	assert.Equal(t, "3 hours ago", groups[2].Label)
	assert.Equal(t, "2 days ago", groups[3].Label)
	assert.Len(t, view.GroupRecents(now, nil), 0)
}

func TestHoursAgo(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "In the last hour", view.HoursAgo(59*time.Minute))
	assert.Equal(t, "23 hours ago", view.HoursAgo(23*time.Hour+59*time.Minute))
	assert.Equal(t, "1 day ago", view.HoursAgo(24*time.Hour))
}
