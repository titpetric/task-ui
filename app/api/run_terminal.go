package api

import (
	"time"
)

// RunTerminal is what a window drives of a run in progress: a
// runner.Session.
type RunTerminal interface {
	Input(p []byte) error
	TakeInput(user string) bool
	Resize(cols, rows int)
	Stop(grace time.Duration)
	Kill(grace time.Duration)
}
