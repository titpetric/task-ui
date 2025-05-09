package internal

import (
	"sync/atomic"
)

// Semaphore provides a run-once type of flag (atomic)
type Semaphore struct {
	semaphore int32
}

// NewSemaphore creates a *Semaphore.
func NewSemaphore() *Semaphore {
	return &Semaphore{
		semaphore: 0,
	}
}

// CanRun will allow a caller to proceed
func (l *Semaphore) CanRun() bool {
	return atomic.CompareAndSwapInt32(&l.semaphore, 0, 1)
}

// Done finishes the run
func (l *Semaphore) Done() {
	atomic.CompareAndSwapInt32(&l.semaphore, 1, 0)
}
