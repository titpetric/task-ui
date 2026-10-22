package storage

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/titpetric/task-ui/model"
)

// NoteStorage keeps the notes and the chat: the messages, the replies to
// them, and how the agent's work on a note goes.
type NoteStorage struct {
	db DB
}

// NewNoteStorage returns the storage over a pool.
func NewNoteStorage(db DB) *NoteStorage {
	return &NoteStorage{db: db}
}

// NoteFilter narrows a listing. Zero values match everything; Mention and
// State match notes that name the user and are in the state.
type NoteFilter struct {
	Channel string
	Mention string
	State   string
	// Prompts lists the prompt threads, the notes with a workspace;
	// false lists the board's notes, which have none.
	Prompts bool
	// WorkspaceID and Dir narrow the prompts to a workspace and to the
	// place they ran, as Note.Dir holds it.
	WorkspaceID string
	Dir         string
	// Limit keeps the newest; zero keeps them all.
	Limit int
}

// Create inserts a note, with the mentions of its body.
func (s *NoteStorage) Create(ctx context.Context, note *model.Note) error {
	note.Body = strings.TrimSpace(note.Body)
	if note.Body == "" || note.UserName == "" {
		return errors.New("a note needs an author and some text")
	}
	if note.Channel == "" {
		note.Channel = model.ChannelNotes
	}
	if note.ID == "" {
		note.ID = model.NewID()
	}
	note.Mentions = model.EncodeMentions(model.Mentions(note.Body))
	note.CreatedAt = stamp()
	note.UpdatedAt = note.CreatedAt
	return client(s.db).Insert(ctx, model.NoteTable, note)
}

// Get returns a note by ID.
func (s *NoteStorage) Get(ctx context.Context, id string) (*model.Note, error) {
	note, err := client(s.db).Get[model.Note](ctx, "SELECT * FROM note WHERE id = ?", id)
	return note, notFound(err)
}

// List returns the notes that start a thread, newest first.
func (s *NoteStorage) List(ctx context.Context, filter NoteFilter) ([]model.Note, error) {
	where := []string{"parent_id = ''"}
	var args []any
	if filter.Channel != "" {
		where = append(where, "channel = ?")
		args = append(args, filter.Channel)
	}
	if filter.Mention != "" {
		where = append(where, "mentions LIKE ?")
		args = append(args, "% "+filter.Mention+" %")
	}
	if filter.State != "" {
		where = append(where, "state = ?")
		args = append(args, filter.State)
	}
	if !filter.Prompts {
		where = append(where, "workspace_id = ''")
	} else {
		where = append(where, "workspace_id != ''")
		if filter.WorkspaceID != "" {
			where = append(where, "workspace_id = ?")
			args = append(args, filter.WorkspaceID)
		}
		if filter.Dir != "" {
			where = append(where, "dir = ?")
			args = append(args, filter.Dir)
		}
	}
	query := "SELECT * FROM note WHERE " + strings.Join(where, " AND ") + " ORDER BY rowid DESC"
	if filter.Limit > 0 {
		query += " LIMIT " + strconv.Itoa(filter.Limit)
	}
	return client(s.db).Select[model.Note](ctx, query, args...)
}

// Replies returns the replies to the notes, oldest first, by note.
func (s *NoteStorage) Replies(ctx context.Context, ids []string) (map[string][]model.Note, error) {
	replies := map[string][]model.Note{}
	if len(ids) == 0 {
		return replies, nil
	}
	marks := strings.TrimSuffix(strings.Repeat("?,", len(ids)), ",")
	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	rows, err := client(s.db).Select[model.Note](ctx, "SELECT * FROM note WHERE parent_id IN ("+marks+") ORDER BY rowid", args...)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		replies[row.ParentID] = append(replies[row.ParentID], row)
	}
	return replies, nil
}

// SetState records how the agent's work on a note goes, and the run.
func (s *NoteStorage) SetState(ctx context.Context, id, state, runID string) error {
	db := client(s.db)
	if err := db.Exec(ctx, "UPDATE note SET state = ?, run_id = CASE WHEN ? = '' THEN run_id ELSE ? END, updated_at = ? WHERE id = ?", state, runID, runID, stamp(), id); err != nil {
		return err
	}
	if db.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete removes a note and its replies.
func (s *NoteStorage) Delete(ctx context.Context, id string) error {
	return client(s.db).Exec(ctx, "DELETE FROM note WHERE id = ? OR parent_id = ?", id, id)
}
