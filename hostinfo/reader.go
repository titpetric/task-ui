package hostinfo

import (
	"context"
	"sync"
	"time"
)

// Reader reads hosts' facts over ssh and keeps each answer for TTL.
// docker ps --size is slow on a host of many containers, so the reader
// runs it once in a while, in the background, and a page shows the last
// answer.
type Reader struct {
	// Run runs a command on a login and returns its output: the ssh
	// prober's Output.
	Run func(ctx context.Context, login, remote string) ([]byte, error)
	// Timeout bounds one host's answer; TTL is how long it stays good.
	Timeout time.Duration
	TTL     time.Duration

	mu      sync.Mutex
	facts   map[string]*Facts
	running map[string]chan struct{}
}

// Facts returns a login's facts, read again from the host when the last
// answer is older than TTL, and waiting at most wait for it; until the
// host answers it returns the last answer, or nil.
func (r *Reader) Facts(ctx context.Context, login string, wait time.Duration) *Facts {
	done := r.start(login)
	if done != nil && wait > 0 {
		select {
		case <-done:
		case <-time.After(wait):
		case <-ctx.Done():
		}
	}
	return r.Cached(login)
}

// Cached returns a login's last facts, or nil, and runs nothing.
func (r *Reader) Cached(login string) *Facts {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.facts[login]
}

// start reads the host unless its answer is fresh, and returns what
// closes once the host answered, the running read's if one runs; nil for
// a fresh answer.
func (r *Reader) start(login string) chan struct{} {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.facts == nil {
		r.facts, r.running = map[string]*Facts{}, map[string]chan struct{}{}
	}
	if facts, ok := r.facts[login]; ok && time.Since(facts.At) < r.TTL {
		return nil
	}
	if done, ok := r.running[login]; ok {
		return done
	}
	done := make(chan struct{})
	r.running[login] = done
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), r.Timeout)
		defer cancel()
		out, err := r.Run(ctx, login, script)
		facts := ParseFacts(string(out))
		if err != nil && len(facts.Filesystems) == 0 {
			facts = &Facts{Err: err}
		}
		facts.At = time.Now()
		r.mu.Lock()
		r.facts[login] = facts
		delete(r.running, login)
		r.mu.Unlock()
		close(done)
	}()
	return done
}
