package storage

import (
	"context"
	"time"

	"github.com/titpetric/task-ui/app/model"
)

// MeetingStorage keeps rooms of the orb, who is in one, and the log each leaves.
type MeetingStorage struct {
	db DB
}

// NewMeetingStorage returns the storage over a pool.
func NewMeetingStorage(db DB) *MeetingStorage {
	return &MeetingStorage{db: db}
}

// Create stores a meeting.
func (s *MeetingStorage) Create(ctx context.Context, m *model.Meeting) error {
	if m.CreatedAt == nil {
		m.CreatedAt = stamp()
	}
	m.UpdatedAt = m.CreatedAt
	return client(s.db).Insert(ctx, "meeting", m)
}

// Get returns a meeting by ID, ErrNotFound for one that is not there.
func (s *MeetingStorage) Get(ctx context.Context, id string) (*model.Meeting, error) {
	m, err := client(s.db).Get[model.Meeting](ctx, "SELECT * FROM meeting WHERE id = ?", id)
	return m, notFound(err)
}

// List returns rooms newest first, the open ones above the ended, at most limit.
func (s *MeetingStorage) List(ctx context.Context, limit int) ([]model.Meeting, error) {
	rooms, err := client(s.db).Select[model.Meeting](ctx,
		"SELECT * FROM meeting ORDER BY (state = 'open') DESC, started_at DESC LIMIT ?", limit)
	if rooms == nil {
		rooms = []model.Meeting{}
	}
	return rooms, err
}

// Open returns every room that takes participants, oldest first.
func (s *MeetingStorage) Open(ctx context.Context) ([]model.Meeting, error) {
	rooms, err := client(s.db).Select[model.Meeting](ctx,
		"SELECT * FROM meeting WHERE state = ? ORDER BY started_at", model.MeetingOpen)
	if rooms == nil {
		rooms = []model.Meeting{}
	}
	return rooms, err
}

// Working returns the ended rooms whose summary a stopped process left running.
func (s *MeetingStorage) Working(ctx context.Context) ([]model.Meeting, error) {
	rooms, err := client(s.db).Select[model.Meeting](ctx,
		"SELECT * FROM meeting WHERE state = ? AND summary_state = ?", model.MeetingEnded, model.SummaryWorking)
	if rooms == nil {
		rooms = []model.Meeting{}
	}
	return rooms, err
}

// Update writes a meeting's state, end, emptiness and summary.
func (s *MeetingStorage) Update(ctx context.Context, m *model.Meeting) error {
	m.UpdatedAt = stamp()
	return client(s.db).Update(ctx, "meeting", m, "id")
}

// Delete removes a meeting with its participants and its log.
func (s *MeetingStorage) Delete(ctx context.Context, id string) error {
	return inTx(ctx, s.db, func(c conn) error {
		if err := c.Exec(ctx, "DELETE FROM meeting_line WHERE meeting_id = ?", id); err != nil {
			return err
		}
		if err := c.Exec(ctx, "DELETE FROM meeting_participant WHERE meeting_id = ?", id); err != nil {
			return err
		}
		return c.Exec(ctx, "DELETE FROM meeting WHERE id = ?", id)
	})
}

// Participant returns one user's row in a room, ErrNotFound for one not there.
func (s *MeetingStorage) Participant(ctx context.Context, meetingID, userID string) (*model.MeetingParticipant, error) {
	p, err := client(s.db).Get[model.MeetingParticipant](ctx,
		"SELECT * FROM meeting_participant WHERE meeting_id = ? AND user_id = ?", meetingID, userID)
	return p, notFound(err)
}

// Participants returns everyone who joined a room, oldest first.
func (s *MeetingStorage) Participants(ctx context.Context, meetingID string) ([]model.MeetingParticipant, error) {
	rows, err := client(s.db).Select[model.MeetingParticipant](ctx,
		"SELECT * FROM meeting_participant WHERE meeting_id = ? ORDER BY joined_at", meetingID)
	if rows == nil {
		rows = []model.MeetingParticipant{}
	}
	return rows, err
}

// Join stores a participant, or revives their row: left_at clears and joins counts the return.
func (s *MeetingStorage) Join(ctx context.Context, p *model.MeetingParticipant) error {
	now := stamp()
	if p.JoinedAt == nil {
		p.JoinedAt = now
	}
	p.LastSeenAt = now
	p.LeftAt = nil
	return client(s.db).Exec(ctx, `INSERT INTO meeting_participant
		(meeting_id, user_id, user_name, joined_at, last_seen_at, left_at, joins)
		VALUES (?1, ?2, ?3, ?4, ?5, NULL, ?6)
		ON CONFLICT(meeting_id, user_id) DO UPDATE SET
			user_name = excluded.user_name, last_seen_at = excluded.last_seen_at,
			left_at = NULL, joins = meeting_participant.joins + 1`,
		p.MeetingID, p.UserID, p.UserName, p.JoinedAt, p.LastSeenAt, max(p.Joins, 1))
}

// Seen marks a participant as present now, ErrNotFound for one not in the room.
func (s *MeetingStorage) Seen(ctx context.Context, meetingID, userID string, at *time.Time) error {
	return client(s.db).Exec(ctx,
		"UPDATE meeting_participant SET last_seen_at = ? WHERE meeting_id = ? AND user_id = ?", at, meetingID, userID)
}

// Left marks a participant as gone, and answers whether they are present.
func (s *MeetingStorage) Left(ctx context.Context, meetingID, userID string, at *time.Time) (bool, error) {
	found, err := client(s.db).Get[model.MeetingParticipant](ctx,
		"SELECT * FROM meeting_participant WHERE meeting_id = ? AND user_id = ? AND left_at IS NULL", meetingID, userID)
	if notFound(err) != nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return found != nil, client(s.db).Exec(ctx,
		"UPDATE meeting_participant SET left_at = ? WHERE meeting_id = ? AND user_id = ?", at, meetingID, userID)
}

// Present counts the participants who have not left, "" for whoever reads their names.
func (s *MeetingStorage) Present(ctx context.Context, meetingID string) (int, error) {
	count, err := client(s.db).Get[int](ctx,
		"SELECT COUNT(*) FROM meeting_participant WHERE meeting_id = ? AND left_at IS NULL", meetingID)
	if err != nil {
		return 0, err
	}
	return *count, nil
}

// Stale returns the participants still present whose last sighting came before the time.
func (s *MeetingStorage) Stale(ctx context.Context, before *time.Time) ([]model.MeetingParticipant, error) {
	rows, err := client(s.db).Select[model.MeetingParticipant](ctx,
		"SELECT * FROM meeting_participant WHERE left_at IS NULL AND last_seen_at < ?", before)
	if rows == nil {
		rows = []model.MeetingParticipant{}
	}
	return rows, err
}

// AddLine stores one entry of a room's log.
func (s *MeetingStorage) AddLine(ctx context.Context, l *model.MeetingLine) error {
	if l.CreatedAt == nil {
		l.CreatedAt = stamp()
	}
	return client(s.db).Insert(ctx, "meeting_line", l)
}

// Lines returns a room's log, oldest first, at most limit, all of it for zero.
func (s *MeetingStorage) Lines(ctx context.Context, meetingID string, limit int) ([]model.MeetingLine, error) {
	query := "SELECT * FROM meeting_line WHERE meeting_id = ? ORDER BY created_at, rowid"
	args := []any{meetingID}
	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}
	lines, err := client(s.db).Select[model.MeetingLine](ctx, query, args...)
	if lines == nil {
		lines = []model.MeetingLine{}
	}
	return lines, err
}
