package storage

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/titpetric/task-ui/app/model"
)

// NoteEventStorage keeps the items of Claude's work on a thread: text, tool
// calls, results, thinking and the status between them.
type NoteEventStorage struct {
	db DB
}

// NewNoteEventStorage returns the storage over a pool.
func NewNoteEventStorage(db DB) *NoteEventStorage {
	return &NoteEventStorage{db: db}
}

// Append adds an event; a run and a sequence that already exist are left as they are.
func (s *NoteEventStorage) Append(ctx context.Context, event *model.NoteEvent) error {
	if event.ID == "" {
		event.ID = model.NewID()
	}
	if event.CreatedAt == nil {
		event.CreatedAt = stamp()
	}
	return client(s.db).Exec(ctx, "INSERT OR IGNORE INTO note_event (id, note_id, run_id, question_id, seq, kind, tool_id, body, data, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		event.ID, event.NoteID, event.RunID, event.QuestionID, event.Seq, event.Kind, event.ToolID, event.Body, event.Data, event.CreatedAt)
}

// List returns a thread's events in the order the conversation reads them.
func (s *NoteEventStorage) List(ctx context.Context, noteID string) ([]model.NoteEvent, error) {
	return client(s.db).Select[model.NoteEvent](ctx, "SELECT * FROM note_event WHERE note_id = ? ORDER BY question_id, run_id, seq", noteID)
}

// ListSince returns a thread's events after one ID, in insertion order, at most count.
func (s *NoteEventStorage) ListSince(ctx context.Context, noteID, afterID string, count int) ([]model.NoteEvent, error) {
	return client(s.db).Select[model.NoteEvent](ctx, "SELECT * FROM note_event WHERE note_id = ? AND rowid > COALESCE((SELECT rowid FROM note_event WHERE id = ?), 0) ORDER BY rowid LIMIT ?", noteID, afterID, count)
}

// ListRun returns one run's events in the order Claude produced them.
func (s *NoteEventStorage) ListRun(ctx context.Context, runID string) ([]model.NoteEvent, error) {
	return client(s.db).Select[model.NoteEvent](ctx, "SELECT * FROM note_event WHERE run_id = ? ORDER BY seq", runID)
}

// Last returns a run's newest event, or ErrNotFound.
func (s *NoteEventStorage) Last(ctx context.Context, runID string) (*model.NoteEvent, error) {
	event, err := client(s.db).Get[model.NoteEvent](ctx, "SELECT * FROM note_event WHERE run_id = ? ORDER BY seq DESC LIMIT 1", runID)
	return event, notFound(err)
}

// DeleteByRun removes a run's events, before a rerun reuses its ID.
func (s *NoteEventStorage) DeleteByRun(ctx context.Context, runID string) error {
	return client(s.db).Exec(ctx, "DELETE FROM note_event WHERE run_id = ?", runID)
}

// DeleteByNote removes a thread's events, for retention.
func (s *NoteEventStorage) DeleteByNote(ctx context.Context, noteID string) error {
	return client(s.db).Exec(ctx, "DELETE FROM note_event WHERE note_id = ?", noteID)
}

// Usage returns what each thread read and wrote, by thread ID.
func (s *NoteEventStorage) Usage(ctx context.Context, noteIDs []string) (map[string]ThreadUsage, error) {
	out := map[string]ThreadUsage{}
	if len(noteIDs) == 0 {
		return out, nil
	}
	marks := strings.TrimSuffix(strings.Repeat("?,", len(noteIDs)), ",")
	args := make([]any, len(noteIDs))
	for i, id := range noteIDs {
		args[i] = id
	}
	rows, err := client(s.db).Select[model.NoteEvent](ctx, "SELECT * FROM note_event WHERE note_id IN ("+marks+") AND kind IN ('tool_use', 'result')", args...)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		usage := out[row.NoteID]
		if row.Kind == "tool_use" {
			usage.Tools++
			out[row.NoteID] = usage
			continue
		}
		var data struct {
			InputTokens int     `json:"input_tokens"`
			Cached      int     `json:"cache_read_input_tokens"`
			Cost        float64 `json:"cost_usd"`
		}
		if json.Unmarshal([]byte(row.Data), &data) == nil {
			usage.Tokens += data.InputTokens + data.Cached
			usage.Cost += data.Cost
		}
		out[row.NoteID] = usage
	}
	return out, nil
}
