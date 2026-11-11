package storage

import (
	"context"
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"

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
	// Threads lists every thread, the board's and the prompts alike, and
	// ignores Prompts and what narrows a prompt.
	Threads bool
}

// Live is what a listing adds to its WHERE: a deleted note and the replies
// under it are read by nothing.
const Live = "deleted_at IS NULL"

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

// Get returns a note by ID, or ErrNotFound once it is deleted.
func (s *NoteStorage) Get(ctx context.Context, id string) (*model.Note, error) {
	note, err := client(s.db).Get[model.Note](ctx, "SELECT * FROM note WHERE id = ? AND "+Live, id)
	return note, notFound(err)
}

// ByRun returns the thread a run of Claude's answers, by the run the
// thread or one of its replies names, or ErrNotFound.
func (s *NoteStorage) ByRun(ctx context.Context, runID string) (*model.Note, error) {
	if runID == "" {
		return nil, ErrNotFound
	}
	note, err := client(s.db).Get[model.Note](ctx, "SELECT * FROM note WHERE run_id = ? AND "+Live+" ORDER BY rowid LIMIT 1", runID)
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
	where := []string{"parent_id = ''", Live}
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
	switch {
	case filter.Threads:
	case !filter.Prompts:
		where = append(where, "workspace_id = ''")
	default:
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
	rows, err := client(s.db).Select[model.Note](ctx, "SELECT * FROM note WHERE parent_id IN ("+marks+") AND "+Live+" ORDER BY rowid", args...)
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

// Answer replaces the agent's answer in its reply, with the run it comes from.
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

// Delete marks a note and its replies deleted, out of every listing.
func (s *NoteStorage) Delete(ctx context.Context, id string) error {
	now := stamp()
	return client(s.db).Exec(ctx, "UPDATE note SET deleted_at = ?, updated_at = ? WHERE (id = ? OR parent_id = ?) AND "+Live, now, now, id, id)
}

// Purge removes a note and its replies, deleted or not, for retention.
func (s *NoteStorage) Purge(ctx context.Context, id string) error {
	return client(s.db).Exec(ctx, "DELETE FROM note WHERE id = ? OR parent_id = ?", id, id)
}

// Threads returns the prompt threads, the notes with a workspace, by
// their last reply, oldest first, for the data retention rules. The
// board's notes are not among them.
func (s *NoteStorage) Threads(ctx context.Context) ([]Thread, error) {
	db := client(s.db)
	roots, err := db.Select[model.Note](ctx, "SELECT * FROM note WHERE parent_id = '' AND workspace_id != '' ORDER BY rowid ASC")
	if err != nil {
		return nil, err
	}
	last, err := db.Select[model.Note](ctx, "SELECT * FROM note WHERE parent_id != '' AND rowid IN (SELECT MAX(rowid) FROM note WHERE parent_id != '' GROUP BY parent_id)")
	if err != nil {
		return nil, err
	}
	open, err := db.Select[noteID](ctx, "SELECT id FROM note WHERE parent_id = '' AND state IN (?, ?) UNION SELECT parent_id AS id FROM note WHERE parent_id != '' AND state IN (?, ?)",
		model.NoteOpen, model.NoteWorking, model.NoteOpen, model.NoteWorking)
	if err != nil {
		return nil, err
	}
	latest, busy := map[string]*time.Time{}, map[string]bool{}
	for i := range last {
		latest[last[i].ParentID] = last[i].CreatedAt
	}
	for _, row := range open {
		busy[row.ID] = true
	}
	threads := make([]Thread, 0, len(roots))
	for _, root := range roots {
		thread := Thread{ID: root.ID, LastAt: root.CreatedAt, Open: busy[root.ID]}
		if at := latest[root.ID]; at != nil {
			thread.LastAt = at
		}
		threads = append(threads, thread)
	}
	sort.SliceStable(threads, func(i, j int) bool {
		return threads[i].LastAt != nil && threads[j].LastAt != nil && threads[i].LastAt.Before(*threads[j].LastAt)
	})
	return threads, nil
}

// noteID is a note's ID, as Threads reads the open ones.
type noteID struct {
	ID string `db:"id"`
}

// DeleteAboutRun removes the threads about a run with their replies.
func (s *NoteStorage) DeleteAboutRun(ctx context.Context, runID string) error {
	return client(s.db).Exec(ctx, "DELETE FROM note WHERE about_run_id = ? OR parent_id IN (SELECT id FROM note WHERE about_run_id = ?)", runID, runID)
}

// DeleteAboutWorkflowRun removes the threads of a workflow run's prompts with their replies, as the workflow run goes.
func (s *NoteStorage) DeleteAboutWorkflowRun(ctx context.Context, id string) error {
	return client(s.db).Exec(ctx, "DELETE FROM note WHERE about_workflow_run_id = ? OR parent_id IN (SELECT id FROM note WHERE about_workflow_run_id = ?)", id, id)
}
