package storage

import (
	"context"

	"github.com/titpetric/task-ui/app/model"
)

// ChatStorage keeps the participants of a chat thread and how far each has read.
type ChatStorage struct {
	db DB
}

// NewChatStorage returns the storage over a pool.
func NewChatStorage(db DB) *ChatStorage {
	return &ChatStorage{db: db}
}

// CreateRoot inserts a chat root and both participants atomically.
func (s *ChatStorage) CreateRoot(ctx context.Context, note *model.Note, participants []Participant) error {
	return inTx(ctx, s.db, func(c conn) error {
		if err := insertNote(ctx, c, note); err != nil {
			return err
		}
		for _, participant := range participants {
			row := &model.NoteParticipant{NoteID: note.ID, UserID: participant.UserID, UserName: participant.UserName, LastReadAt: participant.LastReadAt, CreatedAt: stamp()}
			if err := c.Insert(ctx, model.NoteParticipantTable, row); err != nil {
				return err
			}
		}
		return nil
	})
}

// Reply inserts a chat reply and moves its author's read marker atomically.
func (s *ChatStorage) Reply(ctx context.Context, note *model.Note, userID string) error {
	return inTx(ctx, s.db, func(c conn) error {
		if err := insertNote(ctx, c, note); err != nil {
			return err
		}
		return c.Exec(ctx, "UPDATE note_participant SET last_read_at = ? WHERE note_id = ? AND user_id = ?", stamp(), note.ParentID, userID)
	})
}

// Root returns a live chat root by ID, or ErrNotFound.
func (s *ChatStorage) Root(ctx context.Context, id string) (*model.Note, error) {
	note, err := client(s.db).Get[model.Note](ctx, "SELECT * FROM note WHERE id = ? AND channel = ? AND parent_id = '' AND "+Live, id, model.ChannelChat)
	return note, notFound(err)
}

// Messages returns a chat thread's root and replies in the order they were sent.
func (s *ChatStorage) Messages(ctx context.Context, id string) ([]model.Note, error) {
	return client(s.db).Select[model.Note](ctx, "SELECT * FROM note WHERE (id = ? OR parent_id = ?) AND "+Live+" ORDER BY rowid", id, id)
}

// IsParticipant reports whether a user may read and post in a chat thread.
func (s *ChatStorage) IsParticipant(ctx context.Context, noteID, userID string) (bool, error) {
	rows, err := client(s.db).Select[model.NoteParticipant](ctx, "SELECT * FROM note_participant WHERE note_id = ? AND user_id = ?", noteID, userID)
	if err != nil {
		return false, err
	}
	return len(rows) > 0, nil
}

// Peer returns the other participant's name, or the first participant when the reader is not one.
func (s *ChatStorage) Peer(ctx context.Context, noteID, userID string) (string, error) {
	rows, err := client(s.db).Select[model.NoteParticipant](ctx, "SELECT * FROM note_participant WHERE note_id = ? ORDER BY created_at, user_id", noteID)
	if err != nil {
		return "", err
	}
	for _, row := range rows {
		if row.UserID != userID {
			return row.UserName, nil
		}
	}
	return "", nil
}

// Audio returns the live chat message an ID names, or ErrNotFound.
func (s *ChatStorage) Audio(ctx context.Context, id string) (*model.Note, error) {
	note, err := client(s.db).Get[model.Note](ctx, "SELECT * FROM note WHERE id = ? AND channel = ? AND "+Live, id, model.ChannelChat)
	return note, notFound(err)
}

// ByAudio returns the live chat message whose audio path is path, or ErrNotFound.
func (s *ChatStorage) ByAudio(ctx context.Context, path string) (*model.Note, error) {
	note, err := client(s.db).Get[model.Note](ctx, "SELECT * FROM note WHERE audio_path = ? AND channel = ? AND "+Live+" ORDER BY rowid LIMIT 1", path, model.ChannelChat)
	return note, notFound(err)
}

// MarkRead moves a participant's read marker to now.
func (s *ChatStorage) MarkRead(ctx context.Context, noteID, userID string) error {
	return client(s.db).Exec(ctx, "UPDATE note_participant SET last_read_at = ? WHERE note_id = ? AND user_id = ?", stamp(), noteID, userID)
}

// Unread counts a participant's unread messages in a chat thread.
func (s *ChatStorage) Unread(ctx context.Context, noteID, userID string) (int64, error) {
	count, err := client(s.db).Get[int64](ctx, `SELECT COUNT(*) FROM note n
		LEFT JOIN note_participant p ON p.note_id = ? AND p.user_id = ?
		WHERE (n.id = ? OR n.parent_id = ?) AND n.deleted_at IS NULL AND n.user_id <> ?
		AND (p.last_read_at IS NULL OR n.created_at > p.last_read_at)`, noteID, userID, noteID, noteID, userID)
	if err != nil {
		return 0, err
	}
	return *count, nil
}

// List returns the chat threads a user participates in, newest message first.
func (s *ChatStorage) List(ctx context.Context, userID string, limit int) ([]ChatSummary, error) {
	if limit <= 0 {
		limit = 50
	}
	return client(s.db).Select[ChatSummary](ctx, `SELECT p.note_id AS id, p.last_read_at AS last_read_at,
		m.id AS last_id, m.user_id AS last_user_id, m.user_name AS last_user_name,
		m.body AS last_body, m.spoken AS last_spoken, m.audio_path AS last_audio_path, m.created_at AS last_at,
		COALESCE((SELECT q.user_name FROM note_participant q
			WHERE q.note_id = p.note_id AND q.user_id <> p.user_id
			ORDER BY q.created_at, q.user_id LIMIT 1), '') AS peer,
		(SELECT COUNT(*) FROM note u
			WHERE (u.id = p.note_id OR u.parent_id = p.note_id) AND u.deleted_at IS NULL
			AND u.user_id <> p.user_id AND (p.last_read_at IS NULL OR u.created_at > p.last_read_at)) AS unread
		FROM note_participant p
		JOIN note n ON n.id = p.note_id AND n.channel = 'chat' AND n.parent_id = '' AND n.deleted_at IS NULL
		JOIN note m ON m.rowid = (SELECT MAX(rowid) FROM note WHERE (id = p.note_id OR parent_id = p.note_id) AND deleted_at IS NULL)
		WHERE p.user_id = ?
		ORDER BY m.rowid DESC LIMIT ?`, userID, limit)
}
