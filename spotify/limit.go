package spotify

import (
	"cmp"
	"errors"
	"sync"
	"time"
)

// ErrLimited reports a user over the rate of write calls.
var ErrLimited = errors.New("too many Spotify changes in a minute; wait a little")

// Limiter caps how many write calls a user makes in a window.
type Limiter struct {
	Max    int
	Window time.Duration
	Now    func() time.Time

	mu   sync.Mutex
	seen map[string][]time.Time
}

// Allow records a call of the user and reports whether it is within the rate.
func (l *Limiter) Allow(user string) bool {
	max, window, now := cmp.Or(l.Max, 30), cmp.Or(l.Window, time.Minute), time.Now()
	if l.Now != nil {
		now = l.Now()
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.seen == nil {
		l.seen = map[string][]time.Time{}
	}
	kept := l.seen[user][:0]
	for _, at := range l.seen[user] {
		if now.Sub(at) < window {
			kept = append(kept, at)
		}
	}
	if len(kept) >= max {
		l.seen[user] = kept
		return false
	}
	l.seen[user] = append(kept, now)
	return true
}
