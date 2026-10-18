package runner

import (
	"sync"
)

// Subscriber is one window attached to a session. Its channel carries the
// output as it happens, after a replay of what the window missed. A
// subscriber that does not keep up loses output rather than slowing the
// run: the channel is bounded, and when it is full the oldest output is
// dropped to make room. The status is never dropped: it is the last word.
type Subscriber struct {
	C <-chan Event

	// ReadOnly subscribers may not send input.
	ReadOnly bool

	mu     sync.Mutex
	c      chan Event
	closed bool
}

func newSubscriber(buffer int, readOnly bool) *Subscriber {
	c := make(chan Event, buffer)
	return &Subscriber{C: c, c: c, ReadOnly: readOnly}
}

// send delivers an event without ever blocking the run. Output that does
// not fit is dropped; a status makes room for itself by dropping the
// oldest output waiting.
func (s *Subscriber) send(event Event) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	for {
		select {
		case s.c <- event:
			return
		default:
		}
		if event.Status == nil {
			return
		}
		select {
		case <-s.c:
		default:
		}
	}
}

func (s *Subscriber) close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.closed {
		s.closed = true
		close(s.c)
	}
}
