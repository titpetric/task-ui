package meeting

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/storage"
)

// The reasons a room refuses an action.
var (
	ErrNotFound = errors.New("the meeting is not there")
	ErrEnded    = errors.New("the meeting has ended")
	ErrOpen     = errors.New("the meeting is still open")
	ErrNotHost  = errors.New("only the creator or an admin ends the meeting")
	ErrTitle    = errors.New("a meeting title is at most 120 characters")
)

// TitleLimit is how long a meeting title may run, in characters.
const TitleLimit = 120

// The times a room lives by.
const (
	// StaleAfter is how long a participant may go unheard from before they count as gone.
	StaleAfter = 30 * time.Second
	// EmptyAfter is how long a room sits with nobody in it before it ends.
	EmptyAfter = 60 * time.Second
	// SweepEvery is how often the sweeper looks at rooms and participants.
	SweepEvery = 15 * time.Second
)

// Service runs rooms: their rows, their audio files and the pipes between participants.
type Service struct {
	Store    *storage.MeetingStorage
	Users    *storage.UserStorage
	Channels *Channels
	Hub      *Hub
	// Env returns the integration variables a user's runs get.
	Env func(ctx context.Context, user, role string) (map[string]string, error)
	// Summarize asks Claude about a meeting's transcript.
	Summarize func(ctx context.Context, env map[string]string, system, prompt, document string) (string, error)
	// Now reads the clock, time.Now unless set.
	Now func() time.Time
	// Logger reports the summaries that failed.
	Logger *slog.Logger
}

// now reads the service's clock.
func (s *Service) now() time.Time {
	if s.Now != nil {
		return s.Now()
	}
	return time.Now()
}

// logger returns the logger a failed summary goes to.
func (s *Service) logger() *slog.Logger {
	if s.Logger != nil {
		return s.Logger
	}
	return slog.Default()
}

// Create opens a room as user, answering it.
func (s *Service) Create(ctx context.Context, user *model.User, title string) (*model.Meeting, error) {
	now := s.now()
	title = strings.TrimSpace(title)
	if len([]rune(title)) > TitleLimit {
		return nil, ErrTitle
	}
	if title == "" {
		title = "Meeting by " + user.Name
	}
	m := &model.Meeting{
		ID: model.NewID(), Title: title, UserID: user.ID, UserName: user.Name,
		State: model.MeetingOpen, StartedAt: &now, LimitSeconds: model.MeetingLimit,
	}
	if err := s.Store.Create(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

// Get returns a room, its participants and the whole log.
func (s *Service) Get(ctx context.Context, id string) (*View, error) {
	m, err := s.Store.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	people, err := s.Store.Participants(ctx, id)
	if err != nil {
		return nil, err
	}
	lines, err := s.Store.Lines(ctx, id, 0)
	if err != nil {
		return nil, err
	}
	return &View{Meeting: m, Participants: people, Lines: lines}, nil
}

// List returns rooms, the open ones above the ended, at most limit.
func (s *Service) List(ctx context.Context, limit int) ([]model.Meeting, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.Store.List(ctx, limit)
}

// Join puts user in a room, logs the visit and answers the line it logged, nil for a second tab.
func (s *Service) Join(ctx context.Context, user *model.User, id string) (*model.Meeting, *model.MeetingParticipant, *model.MeetingLine, error) {
	m, err := s.Store.Get(ctx, id)
	if err != nil {
		return nil, nil, nil, err
	}
	if m.State != model.MeetingOpen {
		return nil, nil, nil, ErrEnded
	}
	now := s.now()
	p, err := s.Store.Participant(ctx, id, user.ID)
	if errors.Is(err, storage.ErrNotFound) {
		p = &model.MeetingParticipant{MeetingID: id, UserID: user.ID, UserName: user.Name, Joins: 1}
		if err := s.Store.Join(ctx, p); err != nil {
			return nil, nil, nil, err
		}
		line, err := s.systemLine(ctx, id, user, model.LineJoin, user.Name+" joined")
		return m, p, line, err
	}
	if err != nil {
		return nil, nil, nil, err
	}
	if p.LeftAt != nil {
		p.UserName = user.Name
		if err := s.Store.Join(ctx, p); err != nil {
			return nil, nil, nil, err
		}
		line, err := s.systemLine(ctx, id, user, model.LineRejoin, user.Name+" rejoined")
		return m, p, line, err
	}
	if err := s.Store.Seen(ctx, id, user.ID, &now); err != nil {
		return nil, nil, nil, err
	}
	return m, p, nil, nil
}

// Leave takes user out of a room, logs the part and starts the empty timer when the room is empty.
func (s *Service) Leave(ctx context.Context, user *model.User, id string) (*model.MeetingLine, error) {
	m, err := s.Store.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if m.State != model.MeetingOpen {
		return nil, nil
	}
	now := s.now()
	present, err := s.Store.Left(ctx, id, user.ID, &now)
	if err != nil || !present {
		return nil, err
	}
	line, err := s.systemLine(ctx, id, user, model.LineLeave, user.Name+" left")
	if err != nil {
		return nil, err
	}
	count, err := s.Store.Present(ctx, id)
	if err != nil {
		return line, err
	}
	if count == 0 {
		m.EmptySince = &now
		err = s.Store.Update(ctx, m)
	}
	return line, err
}

// Heartbeat marks user as still in the room.
func (s *Service) Heartbeat(ctx context.Context, user *model.User, id string) error {
	now := s.now()
	return s.Store.Seen(ctx, id, user.ID, &now)
}

// Audio appends a PCM chunk to user's channel and answers its offset in bytes.
func (s *Service) Audio(ctx context.Context, m *model.Meeting, user *model.User, pcm []byte) (int64, error) {
	if m.State != model.MeetingOpen {
		return 0, ErrEnded
	}
	return s.Channels.Append(m.ID, user.ID, pcm)
}

// Line logs what user said, with the span of their channel it covers, and answers the line.
func (s *Service) Line(ctx context.Context, m *model.Meeting, user *model.User, text string, durationMS int64) (*model.MeetingLine, error) {
	text = strings.TrimSpace(text)
	if text == "" || m.State != model.MeetingOpen {
		return nil, nil
	}
	size, err := s.Channels.Size(m.ID, user.ID)
	if err != nil {
		size = 0
	}
	end := size / BytesPerMS
	start := end - durationMS
	if start < 0 {
		start = 0
	}
	line := &model.MeetingLine{
		ID: model.NewID(), MeetingID: m.ID, UserID: user.ID, UserName: user.Name,
		Kind: model.LineSaid, Body: text, StartMs: start, EndMs: end,
	}
	if err := s.Store.AddLine(ctx, line); err != nil {
		return nil, err
	}
	return line, nil
}

// Finish ends a room as user, who is its creator or an admin, and starts its summary.
func (s *Service) Finish(ctx context.Context, user *model.User, id string) (*model.Meeting, error) {
	m, err := s.Store.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if m.State != model.MeetingOpen {
		return m, nil
	}
	if user.ID != m.UserID && !user.CanAdmin() {
		return nil, ErrNotHost
	}
	if err := s.end(ctx, m, "The meeting was ended by "+user.Name); err != nil {
		return nil, err
	}
	return m, nil
}

// RetrySummary starts a summary of an ended room that has none working, as user.
func (s *Service) RetrySummary(ctx context.Context, user *model.User, id string) (*model.Meeting, error) {
	m, err := s.Store.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if m.State != model.MeetingEnded {
		return nil, ErrOpen
	}
	if m.SummaryState == model.SummaryWorking {
		return m, nil
	}
	m.SummaryState, m.SummaryMessage = model.SummaryWorking, ""
	if s.Summarize == nil || s.Env == nil {
		m.SummaryState = model.SummaryFailed
		m.SummaryMessage = "Claude is not set up on this instance"
	}
	if err := s.Store.Update(ctx, m); err != nil {
		return nil, err
	}
	if m.SummaryState == model.SummaryWorking {
		go s.summarize(m)
	}
	return m, nil
}

// Delete removes a room, its log and its audio.
func (s *Service) Delete(ctx context.Context, id string) error {
	if _, err := s.Store.Get(ctx, id); err != nil {
		return err
	}
	if err := s.Store.Delete(ctx, id); err != nil {
		return err
	}
	_ = s.Channels.Remove(id)
	s.Hub.End(id)
	return nil
}

// End ends a room for the sweeper, which has no user to name, and starts its summary.
func (s *Service) End(ctx context.Context, m *model.Meeting, reason string) error {
	if m.State != model.MeetingOpen {
		return nil
	}
	return s.end(ctx, m, reason)
}

// end ends a room once, logs it, tells the pipes and starts the summary.
func (s *Service) end(ctx context.Context, m *model.Meeting, reason string) error {
	now := s.now()
	m.State = model.MeetingEnded
	m.EndedAt = &now
	m.EmptySince = nil
	m.SummaryState = model.SummaryWorking
	m.SummaryMessage = ""
	m.SummaryUserName = m.UserName
	if s.Summarize == nil || s.Env == nil {
		m.SummaryState = model.SummaryFailed
		m.SummaryMessage = "Claude is not set up on this instance"
	}
	if err := s.Store.Update(ctx, m); err != nil {
		return err
	}
	line := &model.MeetingLine{
		ID: model.NewID(), MeetingID: m.ID, Kind: model.LineEnd, Body: reason,
	}
	if err := s.Store.AddLine(ctx, line); err == nil {
		s.Hub.Broadcast(m.ID, nil, TextFrame(Message{Type: "end", Line: line}))
	}
	s.Hub.End(m.ID)
	if m.SummaryState == model.SummaryWorking {
		go s.summarize(m)
	}
	return nil
}

// systemLine logs a join, leave, rejoin or end and tells the pipes.
func (s *Service) systemLine(ctx context.Context, meetingID string, user *model.User, kind, body string) (*model.MeetingLine, error) {
	line := &model.MeetingLine{
		ID: model.NewID(), MeetingID: meetingID, UserID: user.ID, UserName: user.Name, Kind: kind, Body: body,
	}
	if err := s.Store.AddLine(ctx, line); err != nil {
		return nil, err
	}
	s.Hub.Broadcast(meetingID, nil, TextFrame(Message{Type: kind, Line: line}))
	return line, nil
}

// Start sweeps rooms in the background until ctx ends.
func (s *Service) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(SweepEvery)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.Sweep(ctx)
			}
		}
	}()
}

// Sweep marks the participants with no pipe and no hearing gone, and ends the rooms past their hour or left empty.
func (s *Service) Sweep(ctx context.Context) {
	now := s.now()
	before := now.Add(-StaleAfter)
	stale, err := s.Store.Stale(ctx, &before)
	if err != nil {
		s.logger().Warn("meetings: stale participants not read", "err", err)
	}
	touched := map[string]bool{}
	for _, p := range stale {
		if s.Hub != nil && s.Hub.PipesOf(p.MeetingID, p.UserID) > 0 {
			_ = s.Store.Seen(ctx, p.MeetingID, p.UserID, &now)
			continue
		}
		present, err := s.Store.Left(ctx, p.MeetingID, p.UserID, &now)
		if err != nil || !present {
			continue
		}
		touched[p.MeetingID] = true
		_, _ = s.systemLine(ctx, p.MeetingID, &model.User{ID: p.UserID, Name: p.UserName}, model.LineLeave, p.UserName+" left")
	}
	for id := range touched {
		room, err := s.Store.Get(ctx, id)
		if err != nil || room.State != model.MeetingOpen || room.EmptySince != nil {
			continue
		}
		count, err := s.Store.Present(ctx, id)
		if err == nil && count == 0 {
			room.EmptySince = &now
			_ = s.Store.Update(ctx, room)
		}
	}

	open, err := s.Store.Open(ctx)
	if err != nil {
		s.logger().Warn("meetings: open rooms not read", "err", err)
		return
	}
	for _, room := range open {
		m := room
		switch {
		case m.StartedAt != nil && now.Sub(*m.StartedAt) >= time.Duration(m.LimitSeconds)*time.Second:
			_ = s.end(ctx, &m, "The meeting reached its time limit")
		case m.EmptySince != nil && now.Sub(*m.EmptySince) >= EmptyAfter:
			count, err := s.Store.Present(ctx, m.ID)
			if err == nil && count == 0 {
				_ = s.end(ctx, &m, "The meeting ended when everyone left")
			}
		}
	}
}
