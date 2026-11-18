package note

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/internal/assert"
)

// fakeSink keeps the records a tail writes and reports each one.
type fakeSink struct {
	mu     sync.Mutex
	events []*model.NoteEvent
	seen   chan *model.NoteEvent
}

// Append keeps the event and sends it to the test.
func (f *fakeSink) Append(_ context.Context, event *model.NoteEvent) error {
	f.mu.Lock()
	f.events = append(f.events, event)
	f.mu.Unlock()
	f.seen <- event
	return nil
}

// Last returns the newest kept event, or an error when there is none.
func (f *fakeSink) Last(_ context.Context, _ string) (*model.NoteEvent, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.events) == 0 {
		return nil, errors.New("no events")
	}
	return f.events[len(f.events)-1], nil
}

// TestThreadTopic asserts a thread is named by its root ID.
func TestThreadTopic(t *testing.T) {
	assert.Equal(t, "thread:01ABC", ThreadTopic("01ABC"))
}

// TestTranscriptTail_ReadsLines asserts complete lines are kept and a partial one waits.
func TestTranscriptTail_ReadsLines(t *testing.T) {
	path := transcriptFile(t, `{"seq":1,"kind":"text","body":"one"}`+"\n"+`{"seq":2,"kind":"tool_use","body":"two","tool_id":"t","data":{"name":"Bash"}}`+"\n"+`{"seq":3,"kind":"text","bo`)
	sink := &fakeSink{seen: make(chan *model.NoteEvent, 4)}
	tail := &transcriptTail{path: path, noteID: "n", runID: "r", questionID: "q", events: sink}
	tick, done := make(chan time.Time), make(chan struct{})
	finished := make(chan struct{})
	go func() {
		tail.follow(tick, done)
		close(finished)
	}()

	tick <- time.Now()
	assert.Equal(t, "one", receive(t, sink.seen).Body)
	second := receive(t, sink.seen)
	assert.Equal(t, "tool_use", second.Kind)
	assert.Equal(t, "t", second.ToolID)
	assert.JSONEq(t, `{"name":"Bash"}`, second.Data)

	close(done)
	<-finished
	assert.Len(t, sink.events, 2)
}

// TestTranscriptTail_SkipsNonJSON asserts a line that is not the agent's is ignored.
func TestTranscriptTail_SkipsNonJSON(t *testing.T) {
	path := transcriptFile(t, "not json\n"+`{"seq":1,"kind":"text","body":"kept"}`+"\n")
	sink := &fakeSink{seen: make(chan *model.NoteEvent, 2)}
	tail := &transcriptTail{path: path, noteID: "n", runID: "r", events: sink}
	tick, done := make(chan time.Time), make(chan struct{})
	finished := make(chan struct{})
	go func() {
		tail.follow(tick, done)
		close(finished)
	}()
	tick <- time.Now()
	assert.Equal(t, "kept", receive(t, sink.seen).Body)
	close(done)
	<-finished
}

// TestTranscriptTail_DropsLongLine asserts an unterminated line over the cap becomes one error.
func TestTranscriptTail_DropsLongLine(t *testing.T) {
	path := transcriptFile(t, strings.Repeat("x", transcriptMaxLine+1))
	sink := &fakeSink{seen: make(chan *model.NoteEvent, 2)}
	tail := &transcriptTail{path: path, noteID: "n", runID: "r", events: sink}
	tick, done := make(chan time.Time), make(chan struct{})
	finished := make(chan struct{})
	go func() {
		tail.follow(tick, done)
		close(finished)
	}()
	tick <- time.Now()
	dropped := receive(t, sink.seen)
	assert.Equal(t, "error", dropped.Kind)
	assert.Equal(t, int64(1), dropped.Seq)
	close(done)
	<-finished
	assert.Len(t, sink.events, 1)
}

// TestTranscriptTail_MissingFile asserts a tail over a file that is not there writes nothing.
func TestTranscriptTail_MissingFile(t *testing.T) {
	sink := &fakeSink{seen: make(chan *model.NoteEvent, 1)}
	tail := &transcriptTail{path: filepath.Join(t.TempDir(), "none"), noteID: "n", runID: "r", events: sink}
	tick, done := make(chan time.Time), make(chan struct{})
	finished := make(chan struct{})
	go func() {
		tail.follow(tick, done)
		close(finished)
	}()
	close(done)
	<-finished
	assert.Empty(t, sink.events)
}

// transcriptFile writes body to a temp file and returns its path.
func transcriptFile(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "events.jsonl")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// receive takes the next event, failing the test when none arrives.
func receive(t *testing.T, seen <-chan *model.NoteEvent) *model.NoteEvent {
	t.Helper()
	select {
	case event := <-seen:
		return event
	case <-time.After(time.Second):
		t.Fatal("no transcript event")
		return nil
	}
}
