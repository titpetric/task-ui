package app_test

import (
	"os"
	"testing"

	"github.com/titpetric/task-ui/internal/killswitch"
)

// TestMain bounds the binary. A test here builds a whole application: the
// catalogue, the runner, the scheduler and the database behind them. An
// unbounded run ends as an OOM kill, which prints no stack and no test name.
//
// killswitch panics on the first reading over the limit, which prints the
// goroutines and names the running test. The limit and the per-second trace are
// environment knobs; the package doc names them.
func TestMain(m *testing.M) {
	stop := killswitch.Watch()
	code := m.Run()
	stop()
	os.Exit(code)
}
