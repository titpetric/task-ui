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

// NoteFilter narrows a listing. Zero values match everything; Mention
// matches the threads whose note or a reply names the user, and State
// the notes in the state.
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
	// AboutRunID narrows the prompts to the threads about a run; without
	// it, a filter by Dir leaves those out, as the run's page lists them.
	AboutRunID string
	// AboutRunIDs narrows the prompts to the threads about any of the
	// runs, as the Files list looks up the summaries of the last runs.
	AboutRunIDs []string
	// AboutWorkflowRunID narrows the prompts to the thread of a workflow
	// run's prompt; a filter by Dir leaves those out, as the workflow
	// run's page lists them.
	AboutWorkflowRunID string
	// Area narrows the prompts to those of an area's page, such as
	// model.AreaWorkflows; a filter by Dir leaves those out.
	Area string
	// UserName narrows to the threads a user started.
	UserName string
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

// ByRun returns the thread a run of Claude's answers, by the run the
// thread or one of its replies names, or ErrNotFound.
func (s *NoteStorage) ByRun(ctx context.Context, runID string) (*model.Note, error) {
	if runID == "" {
		return nil, ErrNotFound
	}
	note, err := client(s.db).Get[model.Note](ctx, "SELECT * FROM note WHERE run_id = ? ORDER BY rowid LIMIT 1", runID)
	if err != nil {
		return nil, notFound(err)
	}
	if note.ParentID != "" {
		return s.Get(ctx, note.ParentID)
	}
	return note, nil
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
		where = append(where, "(mentions LIKE ? OR id IN (SELECT parent_id FROM note WHERE parent_id != '' AND mentions LIKE ?))")
		args = append(args, "% "+filter.Mention+" %", "% "+filter.Mention+" %")
	}
	if filter.State != "" {
		where = append(where, "state = ?")
		args = append(args, filter.State)
	}
	if filter.UserName != "" {
		where = append(where, "user_name = ?")
		args = append(args, filter.UserName)
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
		if filter.AboutRunID != "" || filter.Dir != "" {
			where = append(where, "about_run_id = ?")
			args = append(args, filter.AboutRunID)
		}
		if len(filter.AboutRunIDs) > 0 {
			where = append(where, "about_run_id IN ("+strings.TrimSuffix(strings.Repeat("?,", len(filter.AboutRunIDs)), ",")+")")
			for _, id := range filter.AboutRunIDs {
				args = append(args, id)
			}
		}
		if filter.AboutWorkflowRunID != "" || filter.Dir != "" {
			where = append(where, "about_workflow_run_id = ?")
			args = append(args, filter.AboutWorkflowRunID)
		}
		if filter.Area != "" || filter.Dir != "" {
			where = append(where, "area = ?")
			args = append(args, filter.Area)
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
	return s.change(ctx, "UPDATE note SET state = ?, run_id = CASE WHEN ? = '' THEN run_id ELSE ? END, updated_at = ? WHERE id = ?", state, runID, runID, stamp(), id)
}

// Release sets a note's state and lets go of its run, for someone who works on it over MCP.
func (s *NoteStorage) Release(ctx context.Context, id, state string) error {
	return s.change(ctx, "UPDATE note SET state = ?, run_id = '', updated_at = ? WHERE id = ?", state, stamp(), id)
}

// Edit puts a new body in a note, with its mentions, and marks it
// edited now.
func (s *NoteStorage) Edit(ctx context.Context, id, body string) error {
	body = strings.TrimSpace(body)
	if body == "" {
		return errors.New("a note needs some text")
	}
	now := stamp()
	return s.change(ctx, "UPDATE note SET body = ?, mentions = ?, edited_at = ?, updated_at = ? WHERE id = ?",
		body, model.EncodeMentions(model.Mentions(body)), now, now, id)
}

// Answer puts the agent's answer in its reply in place of the old one, with the run it comes from.
// The state is working while a run answers it again, empty once it has.
func (s *NoteStorage) Answer(ctx context.Context, id, body, state, runID string) error {
	body = strings.TrimSpace(body)
	if body == "" {
		return errors.New("a note needs some text")
	}
	return s.change(ctx, "UPDATE note SET body = ?, mentions = ?, state = ?, run_id = CASE WHEN ? = '' THEN run_id ELSE ? END, updated_at = ? WHERE id = ?",
		body, model.EncodeMentions(model.Mentions(body)), state, runID, runID, stamp(), id)
}

// change runs an update of one note, ErrNotFound when there is none.
func (s *NoteStorage) change(ctx context.Context, query string, args ...any) error {
	db := client(s.db)
	if err := db.Exec(ctx, query, args...); err != nil {
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
