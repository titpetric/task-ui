package event_test

import (
	"testing"
	"time"

	"github.com/titpetric/task-ui/event"
	"github.com/titpetric/task-ui/internal/assert"
)

func next(t *testing.T, ch <-chan event.Event) event.Event {
	t.Helper()
	select {
	case e := <-ch:
		return e
	case <-time.After(2 * time.Second):
		t.Fatal("no event")
		return event.Event{}
	}
}

// A subscriber gets the events of its topic, in order, and none of
// another topic's; after it unsubscribes its channel is closed.
func TestBus(t *testing.T) {
	t.Parallel()
	bus := event.NewBus()
	ch, cancel := bus.Subscribe("workflow:1")
	other, cancelOther := bus.Subscribe("workflow:2")
	defer cancelOther()

	bus.Publish("workflow:1", "run", map[string]any{"id": "a"})
	bus.Publish("workflow:2", "run", map[string]any{"id": "b"})
	bus.Publish("workflow:1", "done", map[string]any{"id": "a"})

	first := next(t, ch)
	assert.Equal(t, "run", first.Name)
	assert.Equal(t, "workflow:1", first.Topic)
	assert.Equal(t, "a", first.Data.(map[string]any)["id"])
	assert.Equal(t, "done", next(t, ch).Name)
	assert.Equal(t, "b", next(t, other).Data.(map[string]any)["id"])

	cancel()
	_, open := <-ch
	assert.False(t, open, "closed after cancel")
	bus.Publish("workflow:1", "run", nil) // nobody listens: no panic, no block
}

// A subscriber that does not read loses events past its buffer; the
// publisher never waits for it.
func TestBus_SlowSubscriber(t *testing.T) {
	t.Parallel()
	bus := event.NewBus()
	ch, cancel := bus.Subscribe("t")
	defer cancel()
	done := make(chan struct{})
	go func() {
		for i := 0; i < event.Buffer*3; i++ {
			bus.Publish("t", "tick", i)
		}
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("publish blocked on a slow subscriber")
	}
	assert.Equal(t, 0, next(t, ch).Data) // the first events are kept
}
