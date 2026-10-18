package runner_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/runner"
)

// waitOutput reads a subscriber until the output holds want.
func waitOutput(t *testing.T, sub *runner.Subscriber, want string) string {
	t.Helper()
	var out strings.Builder
	deadline := time.After(10 * time.Second)
	for !strings.Contains(out.String(), want) {
		select {
		case event, ok := <-sub.C:
			if !ok {
				t.Fatalf("closed before %q, output %q", want, out.String())
			}
			out.Write(event.Output)
			if event.Status != nil {
				t.Fatalf("finished before %q, output %q", want, out.String())
			}
		case <-deadline:
			t.Fatalf("no %q after 10s, output %q", want, out.String())
		}
	}
	return out.String()
}

func TestSession_Attach(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	session, err := f.manager.Start(context.Background(), f.shellJob("late", "echo first; sleep 30", false))
	assert.NoError(t, err)
	first := session.Attach(true)
	waitOutput(t, first, "first")
	assert.Equal(t, 1, session.Attached())

	// A window attaching later sees what it missed.
	second := session.Attach(false)
	assert.Equal(t, 2, session.Attached())
	assert.Contains(t, waitOutput(t, second, "first"), "first")
	assert.True(t, first.ReadOnly)
	assert.False(t, second.ReadOnly)

	// Detaching leaves the run going.
	session.Detach(first)
	session.Detach(first)
	assert.Equal(t, 1, session.Attached())
	assert.Equal(t, model.StateRunning, session.Status().State)

	session.Kill(time.Second)
	_, status := collect(t, second)
	assert.Equal(t, model.StateKilled, status.State)
	assert.Equal(t, 0, session.Attached())
}

func TestSession_Input(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	session, err := f.manager.Start(context.Background(), f.shellJob("cat", "cat", true))
	assert.NoError(t, err)
	sub := session.Attach(false)
	time.Sleep(100 * time.Millisecond)
	assert.NoError(t, session.Input([]byte("hi there\n")))
	waitOutput(t, sub, "hi there")

	session.Stop(time.Second)
	_, status := collect(t, sub)
	assert.Equal(t, model.StateStopped, status.State)
	assert.ErrorIs(t, session.Input([]byte("x")), runner.ErrNotRunning)

	plain, err := f.manager.Start(context.Background(), f.shellJob("plain", "sleep 30", false))
	assert.NoError(t, err)
	time.Sleep(100 * time.Millisecond)
	assert.ErrorIs(t, plain.Input([]byte("x")), runner.ErrNotInteractive)
	plain.Kill(time.Second)
	<-plain.Done()
}

func TestSession_Resize(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	session, err := f.manager.Start(context.Background(), f.shellJob("size", "sleep 0.3; stty size", false))
	assert.NoError(t, err)
	time.Sleep(100 * time.Millisecond)
	session.Resize(100, 40)
	session.Resize(0, 0)
	out, status := collect(t, session.Attach(true))
	assert.Equal(t, model.StateOK, status.State)
	assert.Contains(t, out, "40 100")
	frame := session.Frame()
	assert.Equal(t, 100, frame.Cols)
	assert.Equal(t, 40, frame.Rows)
}

func TestSession_Frame(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	session, err := f.manager.Start(context.Background(), f.shellJob("frame", `printf 'plain \033[31mred\033[0m \033[1mbold\033[0m\n'`, false))
	assert.NoError(t, err)
	_, status := collect(t, session.Attach(true))
	assert.Equal(t, model.StateOK, status.State)

	frame := session.Frame()
	assert.Equal(t, 80, frame.Cols)
	assert.Equal(t, 24, frame.Rows)
	text := frame.Text()
	assert.Contains(t, text, "plain red bold")
	assert.Equal(t, 24, strings.Count(text, "\n"))

	svg := frame.SVG(runner.DarkPalette)
	assert.True(t, strings.HasPrefix(svg, "<svg"))
	assert.Contains(t, svg, "red")
	assert.Contains(t, svg, runner.DarkPalette.ANSI[1])
	assert.Contains(t, svg, `font-weight="bold"`)
}

func TestSession_Write(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	session, err := f.manager.Start(context.Background(), f.shellJob("note", "sleep 30", false))
	assert.NoError(t, err)
	n, err := session.Write([]byte("a note from the manager\r\n"))
	assert.NoError(t, err)
	assert.Equal(t, 25, n)
	assert.Contains(t, waitOutput(t, session.Attach(true), "a note"), "a note from the manager")
	assert.Contains(t, session.Frame().Text(), "a note from the manager")
	session.Kill(time.Second)
	<-session.Done()
}

func TestSession_Stop(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	session, err := f.manager.Start(context.Background(), f.shellJob("stop", "sleep 30", false))
	assert.NoError(t, err)
	time.Sleep(50 * time.Millisecond)
	session.Stop(time.Second)
	<-session.Done()
	assert.Equal(t, model.StateStopped, session.Status().State)
	snapshot := session.Snapshot()
	assert.Equal(t, snapshot.ID, session.ID())
	assert.Equal(t, snapshot.Key(), session.Key())
	assert.Equal(t, model.StateStopped, snapshot.State)
	assert.NotNil(t, snapshot.FinishedAt)
}

// Input typed before the process is up reaches it once it starts, and a
// window may size the terminal before the process exists.
func TestSession_InputBeforeStart(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	req := f.shellJob("early", `read line; echo "got $line"; stty size`, true)

	session, err := f.manager.Start(context.Background(), req)
	assert.NoError(t, err)
	session.Resize(100, 40)
	assert.NoError(t, session.Input([]byte("hello\n")))

	out, status := collect(t, session.Attach(false))
	assert.Contains(t, out, "got hello")
	assert.Contains(t, out, "40 100")
	assert.Equal(t, model.StateOK, status.State)
	assert.ErrorIs(t, session.Input([]byte("late\n")), runner.ErrNotRunning)
}
