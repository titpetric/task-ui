package view

import (
	"time"
)

// RecentGroup is the recents that started in one hour, or for the older
// ones one day, under its heading.
type RecentGroup struct {
	Label   string
	Entries []Recent
}

// GroupRecents puts recents, newest first, under headings of how long ago
// they started: in the last hour, 1 hour ago, 2 hours ago, and past a day,
// 1 day ago and on.
func GroupRecents(now time.Time, recents []Recent) []RecentGroup {
	var groups []RecentGroup
	for _, recent := range recents {
		label := HoursAgo(now.Sub(recent.StartedAt()))
		if n := len(groups); n > 0 && groups[n-1].Label == label {
			groups[n-1].Entries = append(groups[n-1].Entries, recent)
			continue
		}
		groups = append(groups, RecentGroup{Label: label, Entries: []Recent{recent}})
	}
	return groups
}

// HoursAgo is the heading of a recent that started a while ago: whole hours,
// then whole days.
func HoursAgo(since time.Duration) string {
	hours := int(since / time.Hour)
	switch {
	case hours < 1:
		return "In the last hour"
	case hours < 24:
		return plural(hours, "hour") + " ago"
	default:
		return plural(hours/24, "day") + " ago"
	}
}
