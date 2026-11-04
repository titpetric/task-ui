package app_test

import (
	"os"
	"testing"

	"github.com/titpetric/task-ui/internal/killswitch"
)

// TestMain bounds the binary. A test here builds a whole application:
// the catalogue, the runner, the scheduler and the database behind them,
// and the run that noticed was an OOM kill with no stack and no test
// name on it.
//
// killswitch panics the first reading over 8GiB instead, which prints
// the goroutines and so says which test was running. The limit and the
// per-second trace are environment knobs; the package doc names them.
func TestMain(m *testing.M) {
	stop := killswitch.Watch()
	code := m.Run()
	stop()
	os.Exit(code)
}
