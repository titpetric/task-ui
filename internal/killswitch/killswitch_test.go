package killswitch_test

import (
	"bytes"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/internal/killswitch"
)

func TestParseSize(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		in   string
		want int64
	}{
		{"0", 0},
		{"512", 512},
		{"512B", 512},
		{"8GiB", 8 << 30},
		{"2gib", 2 << 30},
		{" 64MiB ", 64 << 20},
		{"1KiB", 1 << 10},
		{"1TiB", 1 << 40},
	} {
		got, err := killswitch.ParseSize(tc.in)
		assert.NoError(t, err)
		assert.EqualValues(t, tc.want, got, tc.in)
	}
	for _, in := range []string{"", "GiB", "8GB", "8 apples", "99999999999999999999"} {
		_, err := killswitch.ParseSize(in)
		assert.Error(t, err, in)
	}
}

// TestStart_Reports is the watcher doing its two jobs: a reading every
// interval under trace, and the peak when it stops.
func TestStart_Reports(t *testing.T) {
	t.Parallel()
	out := &buffer{}
	stop := killswitch.Start(killswitch.Options{
		Limit:  -1,
		Every:  time.Millisecond,
		Report: time.Millisecond,
		Trace:  true,
		Out:    out,
		Kill:   func(string) { t.Error("killed under no limit") },
	})
	// lifetime: long enough for a second reading, short enough to stay
	// inside the budget.
	time.Sleep(20 * time.Millisecond)
	stop()
	stop()

	got := out.String()
	assert.True(t, strings.Count(got, "killswitch: rss ") >= 2, got)
	assert.Contains(t, got, "goroutines")
	assert.Contains(t, got, "killswitch: peak rss ")
}

// TestStart_Quiet is the report interval holding the trace to one line:
// the watcher reads far more often than it prints, so the peak is close
// to the truth without the output being a wall.
func TestStart_Quiet(t *testing.T) {
	t.Parallel()
	out := &buffer{}
	stop := killswitch.Start(killswitch.Options{
		Limit:  -1,
		Every:  time.Millisecond,
		Report: time.Hour,
		Trace:  true,
		Out:    out,
	})
	// lifetime: twenty readings at a millisecond, of which the hour-long
	// report interval lets through the first.
	time.Sleep(20 * time.Millisecond)
	stop()

	got := out.String()
	assert.EqualValues(t, 1, strings.Count(got, "killswitch: rss "), got)
	assert.Contains(t, got, "killswitch: peak rss ")
}

// TestStart_Kills is the limit being passed: one byte is a limit every
// process is over, so the first reading kills.
func TestStart_Kills(t *testing.T) {
	t.Parallel()
	out := &buffer{}
	killed := make(chan string, 1)
	stop := killswitch.Start(killswitch.Options{
		Limit: 1,
		Every: time.Millisecond,
		Out:   out,
		Kill:  func(line string) { killed <- line },
	})
	defer stop()

	select {
	case line := <-killed:
		assert.Contains(t, line, "over the 1B limit")
		assert.Contains(t, out.String(), "killswitch: rss ")
	case <-time.After(5 * time.Second):
		t.Fatal("no reading over a one byte limit")
	}
}

// TestWatch_Off is TASK_UI_TEST_MEM_KILL=off: a stopper that reports
// nothing, because a run that asked for no watcher wants no output.
func TestWatch_Off(t *testing.T) {
	t.Setenv(killswitch.KillEnv, "off")
	killswitch.Watch()()
}

// buffer is a bytes.Buffer a watcher's goroutine can write while the
// test reads it.
type buffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *buffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *buffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}
