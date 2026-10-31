package runner_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/runner"
)

func titleSession() *runner.Session {
	run := &model.Run{ID: model.NewID(), Folder: "f", Job: "build", TermCols: 80, TermRows: 24}
	return runner.NewSession(run, &folder.Folder{ID: "f"}, &folder.Job{ID: "build"})
}

// The terminal names the run with OSC 0 or 2, ended with BEL or ST,
// whole or split across writes; OSC 1, the icon name, does not.
func TestSession_Title(t *testing.T) {
	t.Parallel()
	session := titleSession()
	write := func(s string) {
		_, err := session.Write([]byte(s))
		assert.NoError(t, err)
	}

	write("hello\x1b]0;building images\x07 more")
	assert.Equal(t, "building images", session.Title())
	write("\x1b]2;pushing\x1b\\")
	assert.Equal(t, "pushing", session.Title())
	write("\x1b]1;icon\x07")
	assert.Equal(t, "pushing", session.Title())
	write("\x1b]2;dep")
	assert.Equal(t, "pushing", session.Title(), "nothing until the sequence ends")
	write("loying\x1b")
	write("\\")
	assert.Equal(t, "deploying", session.Title())
	write("\x1b]0;\ttabs\x01 and  \x07")
	assert.Equal(t, "tabs and", session.Title())
	write("\x1b]0;\x07")
	assert.Equal(t, "", session.Title(), "an empty title clears it")
	assert.Equal(t, "", session.Snapshot().Title)

	write("\x1b]0;row\x07")
	assert.Equal(t, "row", session.Snapshot().Title, "the run row carries it")
}

// A window hears the title on attach and on every change.
func TestSession_Title_attach(t *testing.T) {
	t.Parallel()
	session := titleSession()
	_, _ = session.Write([]byte("\x1b]0;first\x07"))
	sub := session.Attach(true)
	defer session.Detach(sub)

	var titles []string
	_, _ = session.Write([]byte("\x1b]0;second\x07"))
	for len(titles) < 2 {
		select {
		case event := <-sub.C:
			if event.Title != nil {
				titles = append(titles, *event.Title)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("no title event")
		}
	}
	assert.Equal(t, []string{"first", "second"}, titles)
}

// Titled hears the first change at once and, of a title that keeps
// changing, the last one once TitleEvery passed.
func TestSession_Titled(t *testing.T) {
	t.Parallel()
	session := titleSession()
	var mu sync.Mutex
	var heard []string
	got := make(chan struct{}, 16)
	session.Titled = func(run model.Run) {
		mu.Lock()
		heard = append(heard, run.Title)
		mu.Unlock()
		got <- struct{}{}
	}
	for i := 1; i <= 5; i++ {
		_, _ = session.Write(fmt.Appendf(nil, "\x1b]0;step %d\x07", i))
	}
	for range 2 {
		select {
		case <-got:
		case <-time.After(2 * time.Second):
			t.Fatal("Titled not called")
		}
	}
	// Two calls and no third: the first title at once, the last of the
	// run of them when TitleEvery passed, and the three in between
	// coalesced. Waiting out another TitleEvery to prove the absence
	// would be half a second of sleeping for a channel that is already
	// empty.
	select {
	case <-got:
		t.Fatal("a third title: the changes between did not coalesce")
	default:
	}
	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, []string{"step 1", "step 5"}, heard)
}
