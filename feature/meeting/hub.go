package meeting

import (
	"sync"
)

// Buffer bounds the frames a pipe may hold before the ones after it are dropped.
const Buffer = 64

// Frame holds one message to a participant: text control, or binary audio.
type Frame struct {
	Text []byte
	Data []byte
}

// Pipe holds one participant's live connection: the frames to write, and who they are.
type Pipe struct {
	MeetingID string
	UserID    string
	UserName  string
	Frames    chan Frame

	once sync.Once
}

// Close ends the pipe's queue once, so a late leave does not close it twice.
func (p *Pipe) Close() {
	p.once.Do(func() { close(p.Frames) })
}

// Hub keeps every room's pipes in memory.
type Hub struct {
	mu    sync.Mutex
	rooms map[string]map[*Pipe]struct{}
}

// NewHub returns an empty hub.
func NewHub() *Hub {
	return &Hub{rooms: map[string]map[*Pipe]struct{}{}}
}

// Join adds a pipe to its room.
func (h *Hub) Join(p *Pipe) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.rooms[p.MeetingID] == nil {
		h.rooms[p.MeetingID] = map[*Pipe]struct{}{}
	}
	h.rooms[p.MeetingID][p] = struct{}{}
}

// Leave removes a pipe and ends its queue.
func (h *Hub) Leave(p *Pipe) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.drop(p)
}

// drop removes a pipe under the lock the caller holds.
func (h *Hub) drop(p *Pipe) {
	room := h.rooms[p.MeetingID]
	if room == nil {
		return
	}
	delete(room, p)
	if len(room) == 0 {
		delete(h.rooms, p.MeetingID)
	}
	p.Close()
}

// Broadcast sends a frame to every pipe of a room except skip, and answers the count of pipes that dropped it.
func (h *Hub) Broadcast(meetingID string, skip *Pipe, f Frame) int {
	h.mu.Lock()
	defer h.mu.Unlock()
	dropped := 0
	for p := range h.rooms[meetingID] {
		if p == skip {
			continue
		}
		select {
		case p.Frames <- f:
		default:
			dropped++
		}
	}
	return dropped
}

// End ends every pipe of a room, for a meeting that is over.
func (h *Hub) End(meetingID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for p := range h.rooms[meetingID] {
		delete(h.rooms[meetingID], p)
		p.Close()
	}
	delete(h.rooms, meetingID)
}

// PipesOf counts a room's pipes that belong to one user.
func (h *Hub) PipesOf(meetingID, userID string) int {
	h.mu.Lock()
	defer h.mu.Unlock()
	count := 0
	for p := range h.rooms[meetingID] {
		if p.UserID == userID {
			count++
		}
	}
	return count
}

// Count returns the pipes a room holds.
func (h *Hub) Count(meetingID string) int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.rooms[meetingID])
}
