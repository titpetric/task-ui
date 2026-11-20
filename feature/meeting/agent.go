package meeting

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/integration/claude"
)

// ErrAgent reports a name no agent user holds.
var ErrAgent = errors.New("no such agent")

// AgentHooks is the note board side of a room's agent.
type AgentHooks struct {
	// Valid returns the user an agent name resolves to.
	Valid func(ctx context.Context, name string) (*model.User, error)
	// Ask puts a discussion line to the thread and answers the thread's ID.
	Ask func(ctx context.Context, host *model.User, thread, agent, origin, body string) (string, error)
	// Watch delivers the notes written to a thread.
	Watch func(thread string) (<-chan model.Note, func())
}

// agentRoom is the discussion a room's agent has not read yet.
type agentRoom struct {
	mu      sync.Mutex
	pending []string
	busy    bool
	cancel  func()
}

// agentQueues keeps a room's pending lines, by meeting.
type agentQueues struct {
	mu   sync.Mutex
	room map[string]*agentRoom
}

// get returns the queue of a room, making it when it is not there.
func (q *agentQueues) get(id string) *agentRoom {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.room == nil {
		q.room = map[string]*agentRoom{}
	}
	if q.room[id] == nil {
		q.room[id] = &agentRoom{}
	}
	return q.room[id]
}

// SetAgent names the agent a room talks to, as its creator or an admin, or clears it.
func (s *Service) SetAgent(ctx context.Context, user *model.User, id, name string) (*model.Meeting, error) {
	m, err := s.Store.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if m.State != model.MeetingOpen {
		return nil, ErrEnded
	}
	if user.ID != m.UserID && !user.CanAdmin() {
		return nil, ErrNotHost
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return s.clearAgent(ctx, m)
	}
	if s.Agent == nil || s.Agent.Valid == nil {
		return nil, errors.New("agents aren't available on this instance")
	}
	agent, err := s.Agent.Valid(ctx, name)
	if err != nil {
		return nil, err
	}
	if m.Agent == agent.Name {
		return m, nil
	}
	m.Agent, m.AgentID, m.AgentNoteID = agent.Name, agent.ID, ""
	if err := s.Store.Update(ctx, m); err != nil {
		return nil, err
	}
	s.stopAgent(m.ID)
	_, _ = s.systemLine(ctx, m.ID, agent, model.LineJoin, agent.Name+" is the room's agent")
	return m, nil
}

// clearAgent takes the room's agent out and ends its watcher.
func (s *Service) clearAgent(ctx context.Context, m *model.Meeting) (*model.Meeting, error) {
	if m.Agent == "" {
		return m, nil
	}
	agent := &model.User{ID: m.AgentID, Name: m.Agent}
	m.Agent, m.AgentID, m.AgentNoteID = "", "", ""
	if err := s.Store.Update(ctx, m); err != nil {
		return nil, err
	}
	s.stopAgent(m.ID)
	_, _ = s.systemLine(ctx, m.ID, agent, model.LineLeave, agent.Name+" left as the agent")
	return m, nil
}

// stopAgent empties a room's queue and ends its watcher.
func (s *Service) stopAgent(id string) {
	room := s.agents.get(id)
	room.mu.Lock()
	defer room.mu.Unlock()
	room.pending = nil
	room.busy = false
	if room.cancel != nil {
		room.cancel()
		room.cancel = nil
	}
}

// discuss queues what a speaker said for the room's agent, and asks when it is idle.
func (s *Service) discuss(m *model.Meeting, user *model.User, text string) {
	if m.Agent == "" || s.Agent == nil || user.Name == m.Agent {
		return
	}
	room := s.agents.get(m.ID)
	room.mu.Lock()
	room.pending = append(room.pending, user.Name+": "+text)
	start := !room.busy
	if start {
		room.busy = true
	}
	room.mu.Unlock()
	if start {
		go s.askAgent(m.ID, room)
	}
}

// askAgent sends the room's queued lines to its agent, one run at a time.
func (s *Service) askAgent(id string, room *agentRoom) {
	ctx := context.Background()
	room.mu.Lock()
	lines := room.pending
	room.pending = nil
	if len(lines) == 0 {
		room.busy = false
		room.mu.Unlock()
		return
	}
	room.mu.Unlock()

	m, err := s.Store.Get(ctx, id)
	if err != nil || m.State != model.MeetingOpen || m.Agent == "" {
		s.releaseAgent(id, room)
		return
	}
	host, err := s.Users.Get(ctx, m.UserID)
	if err != nil {
		s.agentLine(ctx, m, "The agent could not start: the creator is not there.")
		s.releaseAgent(id, room)
		return
	}
	body := strings.Join(lines, "\n")
	thread := m.AgentNoteID
	if thread == "" {
		body = claude.MeetingDispatcher + "\n\nThe discussion:\n\n" + body
	}
	opened, err := s.Agent.Ask(ctx, host, thread, m.Agent, "meeting:"+id, body)
	if err != nil {
		s.agentLine(ctx, m, "The agent could not start: "+err.Error())
		s.releaseAgent(id, room)
		return
	}
	if thread == "" && opened != "" {
		m.AgentNoteID = opened
		if err := s.Store.Update(ctx, m); err != nil {
			s.logger().Warn("meetings: agent thread not stored", "meeting", id, "err", err)
		}
		s.watchAgent(m, opened, room)
	}
}

// releaseAgent marks the room idle, and starts the next batch if one arrived.
func (s *Service) releaseAgent(id string, room *agentRoom) {
	room.mu.Lock()
	room.busy = false
	more := len(room.pending) > 0
	if more {
		room.busy = true
	}
	room.mu.Unlock()
	if more {
		go s.askAgent(id, room)
	}
}

// watchAgent relays the replies an agent writes to a room's transcript.
func (s *Service) watchAgent(m *model.Meeting, thread string, room *agentRoom) {
	if s.Agent.Watch == nil {
		return
	}
	replies, cancel := s.Agent.Watch(thread)
	room.mu.Lock()
	if room.cancel != nil {
		room.cancel()
	}
	room.cancel = cancel
	room.mu.Unlock()
	go func() {
		ctx := context.Background()
		for note := range replies {
			if note.ParentID == "" || note.UserName != m.Agent {
				continue
			}
			line := &model.MeetingLine{
				ID: model.NewID(), MeetingID: m.ID, UserID: note.UserID,
				UserName: note.UserName, Kind: model.LineSaid, Body: note.Body,
			}
			if err := s.Store.AddLine(ctx, line); err == nil {
				s.Hub.Broadcast(m.ID, nil, TextFrame(Message{Type: "line", Line: line}))
			}
			room.mu.Lock()
			room.busy = false
			more := len(room.pending) > 0
			if more {
				room.busy = true
			}
			room.mu.Unlock()
			if more {
				go s.askAgent(m.ID, room)
			}
		}
	}()
}

// agentLine adds a system line about the agent to a room.
func (s *Service) agentLine(ctx context.Context, m *model.Meeting, body string) {
	agent := &model.User{ID: m.AgentID, Name: m.Agent}
	_, _ = s.systemLine(ctx, m.ID, agent, model.LineJoin, body)
}
