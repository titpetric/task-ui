package meeting

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
)

// A hub sends a room's frames to every pipe but the sender, and ends them all when the room does.
func TestHub(t *testing.T) {
	t.Parallel()
	hub := NewHub()
	id := "room"
	ana := &Pipe{MeetingID: id, UserID: "u1", UserName: "ana", Frames: make(chan Frame, Buffer)}
	bob := &Pipe{MeetingID: id, UserID: "u2", UserName: "bob", Frames: make(chan Frame, Buffer)}
	hub.Join(ana)
	hub.Join(bob)
	assert.Equal(t, 2, hub.Count(id))

	frame := Frame{Text: []byte("hello")}
	assert.Equal(t, 0, hub.Broadcast(id, ana, frame))
	assert.Equal(t, 0, len(ana.Frames))
	got := <-bob.Frames
	assert.Equal(t, "hello", string(got.Text))

	hub.Leave(bob)
	assert.Equal(t, 1, hub.Count(id))
	_, open := <-bob.Frames
	assert.False(t, open)

	hub.End(id)
	assert.Equal(t, 0, hub.Count(id))
	_, open = <-ana.Frames
	assert.False(t, open)
}

// A hub counts a user's pipes, and a pipe closes its queue once however often it is closed.
func TestHub_PipesOf(t *testing.T) {
	t.Parallel()
	hub := NewHub()
	id := "room"
	first := &Pipe{MeetingID: id, UserID: "u1", Frames: make(chan Frame, Buffer)}
	second := &Pipe{MeetingID: id, UserID: "u1", Frames: make(chan Frame, Buffer)}
	hub.Join(first)
	hub.Join(second)
	assert.Equal(t, 2, hub.PipesOf(id, "u1"))
	assert.Equal(t, 0, hub.PipesOf(id, "u2"))

	first.Close()
	first.Close()
	_, open := <-first.Frames
	assert.False(t, open)
	assert.Equal(t, 2, hub.PipesOf(id, "u1"))
	hub.Leave(first)
	assert.Equal(t, 1, hub.PipesOf(id, "u1"))
}

// A pipe that stops reading drops frames, and the hub counts the drop.
func TestHub_Drop(t *testing.T) {
	t.Parallel()
	hub := NewHub()
	id := "room"
	frame := Frame{Data: []byte{1}}
	slow := &Pipe{MeetingID: id, UserID: "u1", Frames: make(chan Frame)}
	hub.Join(slow)
	assert.Equal(t, 1, hub.Broadcast(id, nil, frame))
	hub.Leave(slow)

	full := &Pipe{MeetingID: id, UserID: "u2", Frames: make(chan Frame, 1)}
	hub.Join(full)
	assert.Equal(t, 0, hub.Broadcast(id, nil, frame))
	assert.Equal(t, 1, hub.Broadcast(id, nil, frame))
	assert.Equal(t, 1, len(full.Frames))
}
