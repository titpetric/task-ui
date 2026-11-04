package storage_test

import (
	"os"
	"testing"

	"github.com/titpetric/task-ui/internal/killswitch"
)

// TestMain bounds the binary. This package holds the integration tests,
// and each of them copies a migrated database; a fixture that holds on
// to one is paid for by every test after it.
//
// killswitch panics the first reading over 8GiB instead of leaving the
// kernel to end the run with no stack and no test name. The limit and
// the per-second trace are environment knobs; the package doc names
// them, and docs/testing.md says what to do with a reading that grows.
func TestMain(m *testing.M) {
	stop := killswitch.Watch()
	code := m.Run()
	stop()
	os.Exit(code)
}
