package runner

import (
	"slices"
	"sync"

	"github.com/titpetric/task-ui/folder"
)

// Slots counts the runs that do work against the instance's limit, and
// keeps the runs that wait for a slot in the order they came.
type Slots struct {
	mu      sync.Mutex
	limit   int
	going   map[string]bool
	waiting []string
}

// NewSlots returns slots for at most limit runs at once, one at the least.
func NewSlots(limit int) *Slots {
	if limit < 1 {
		limit = 1
	}
	return &Slots{limit: limit, going: map[string]bool{}}
}

// TakesSlot reports whether a run of the job takes a slot: no terminal.
func TakesSlot(job *folder.Job) bool {
	return job != nil && !job.Interactive && !job.IsShell()
}

// Limit is how many runs go at once.
func (s *Slots) Limit() int {
	return s.limit
}

// Take reserves a slot for the run when one is free and no run waits before
// it, and reports true; else the run waits, last in line.
func (s *Slots) Take(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.going[id] {
		return true
	}
	if slices.Contains(s.waiting, id) {
		return false
	}
	if len(s.waiting) == 0 && s.free() {
		s.going[id] = true
		return true
	}
	s.waiting = append(s.waiting, id)
	return false
}

// Full reports whether a run that comes now waits.
func (s *Slots) Full() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.waiting) > 0 || !s.free()
}

// Room reports whether a run that comes now finds a slot or a place in
// a line that holds fewer than backlog runs.
func (s *Slots) Room(backlog int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.waiting) == 0 && s.free() {
		return true
	}
	return len(s.waiting) < backlog
}

// free reports whether a slot is free. The caller holds mu.
func (s *Slots) free() bool {
	return len(s.going) < s.limit
}

// Free releases a finished run's slot and returns the waiting runs, in order, that take the slots free now.
func (s *Slots) Free(id string) []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.going, id)
	var next []string
	for len(s.waiting) > 0 && s.free() {
		first := s.waiting[0]
		s.waiting = s.waiting[1:]
		s.going[first] = true
		next = append(next, first)
	}
	return next
}

// Drop takes a waiting run out of the line, and reports whether it
// waited.
func (s *Slots) Drop(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	at := slices.Index(s.waiting, id)
	if at < 0 {
		return false
	}
	s.waiting = slices.Delete(s.waiting, at, at+1)
	return true
}

// Clear empties the line and returns the runs that waited, in order.
func (s *Slots) Clear() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	waiting := s.waiting
	s.waiting = nil
	return waiting
}

// Position is where a waiting run stands in line, 1 for the next one, 0 for a run that does not wait.
func (s *Slots) Position(id string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return slices.Index(s.waiting, id) + 1
}

// Going is how many runs hold a slot.
func (s *Slots) Going() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.going)
}

// Waiting is how many runs wait for a slot.
func (s *Slots) Waiting() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.waiting)
}
