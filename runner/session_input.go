package runner

import (
	"log/slog"
)

// TakeInput opens the pty of a job that takes no input to the keystrokes Input sends, to answer a prompt it waits on.
// It reports whether user came first; the recording and the log name that first user.
func (s *Session) TakeInput(user string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.inputBy != "" || s.status.Finished() {
		return false
	}
	s.inputBy = user
	for _, sink := range s.sinks {
		if m, ok := sink.(marker); ok {
			_ = m.Mark("input taken by " + user)
		}
	}
	slog.Info("run input taken", "run", s.Run.ID, "user", user)
	return true
}

// InputBy returns who took input on the run, "" for nobody.
func (s *Session) InputBy() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.inputBy
}

// marker is a sink that records an event besides the output: the
// recording, which keeps it as a marker.
type marker interface {
	Mark(label string) error
}
