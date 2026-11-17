//go:build integration

package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/titpetric/task-ui/app"
	"github.com/titpetric/task-ui/app/api"
	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/event"
	"github.com/titpetric/task-ui/feature/note"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/internal/permission"
	"github.com/titpetric/task-ui/storage"
)

// voiceNoteRoutes serves the API over a board and its users in memory.
func voiceNoteRoutes(t *testing.T) (http.Handler, *note.Board, *storage.UserStorage, *model.User) {
	t.Helper()
	ctx := context.Background()
	db, err := storage.OpenAndMigrate(ctx, ":memory:", app.Migrations())
	assert.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	users := storage.NewUserStorage(db)
	admin := &model.User{Name: "root", Role: model.RoleAdmin}
	assert.NoError(t, users.Create(ctx, admin))
	board := &note.Board{Notes: storage.NewNoteStorage(db), Users: users, Events: event.NewBus()}
	return api.New(api.Options{Board: board, Users: users}).Routes(), board, users, admin
}

// A spoken note comes back as a board note with its words and clip, and no
// workspace or state.
func TestVoiceNote_Record(t *testing.T) {
	t.Parallel()
	h, board, _, admin := voiceNoteRoutes(t)
	rec := voiceDo(h, admin, http.MethodPost, "/voice/note", `{"spoken":"hello","audio":"voice/x.wav"}`)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var row model.Note
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &row))
	assert.Equal(t, "hello", row.Body)
	assert.Equal(t, "hello", row.Spoken)
	assert.Equal(t, "voice/x.wav", row.AudioPath)
	assert.Equal(t, "", row.WorkspaceID)
	assert.Equal(t, "", row.State)

	held, err := board.Notes.Get(context.Background(), row.ID)
	assert.NoError(t, err)
	assert.Equal(t, row.AudioPath, held.AudioPath)
}

// A parent id makes the next clip a reply in the note thread.
func TestVoiceNote_Reply(t *testing.T) {
	t.Parallel()
	h, board, _, admin := voiceNoteRoutes(t)

	rec := voiceDo(h, admin, http.MethodPost, "/voice/note", `{"spoken":"first","audio":"voice/first.wav"}`)
	assert.Equal(t, http.StatusCreated, rec.Code)
	var root model.Note
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &root))

	rec = voiceDo(h, admin, http.MethodPost, "/voice/note", `{"parent_id":"`+root.ID+`","spoken":"second"}`)
	assert.Equal(t, http.StatusCreated, rec.Code)
	var reply model.Note
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &reply))
	assert.Equal(t, root.ID, reply.ParentID)

	replies, err := board.Notes.Replies(context.Background(), []string{root.ID})
	assert.NoError(t, err)
	assert.Len(t, replies[root.ID], 1)
}

// Nothing said, a prompt's thread and an unknown parent are refused.
func TestVoiceNote_Refuses(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	h, board, _, admin := voiceNoteRoutes(t)

	rec := voiceDo(h, admin, http.MethodPost, "/voice/note", `{"spoken":"","audio":""}`)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	prompt := &model.Note{Channel: model.ChannelNotes, UserID: admin.ID, UserName: admin.Name, Body: "a prompt", WorkspaceID: "infra"}
	assert.NoError(t, board.Notes.Create(ctx, prompt))
	rec = voiceDo(h, admin, http.MethodPost, "/voice/note", `{"parent_id":"`+prompt.ID+`","spoken":"an answer"}`)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	rec = voiceDo(h, admin, http.MethodPost, "/voice/note", `{"parent_id":"no-such-note","spoken":"an answer"}`)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// The note route takes note/edit, and a reader who holds it may dictate.
func TestVoiceNote_NeedsNoteEdit(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	h, _, users, _ := voiceNoteRoutes(t)

	viewer := &model.User{Name: "view", Role: model.RoleViewer}
	assert.NoError(t, users.Create(ctx, viewer))
	assert.NoError(t, users.Policies().Put(ctx, permission.KindUser, viewer.Name, "note/edit", permission.Deny))
	rec := voiceDo(h, viewer, http.MethodPost, "/voice/note", `{"spoken":"hello"}`)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Contains(t, rec.Body.String(), "note/edit")

	assert.NoError(t, users.Policies().Put(ctx, permission.KindUser, viewer.Name, "note/edit", permission.Allow))
	rec = voiceDo(h, viewer, http.MethodPost, "/voice/note", `{"spoken":"hello"}`)
	assert.Equal(t, http.StatusCreated, rec.Code)

	// Transcription follows the same rule: the reader may dictate, and the
	// route answers 200 for ready and 503 only because this instance holds no transcriber.
	rec = voiceDo(h, viewer, http.MethodGet, "/voice/ready", "")
	assert.Equal(t, http.StatusOK, rec.Code)
	rec = voiceDo(h, viewer, http.MethodPost, "/voice/transcribe", "RIFFclip")
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
}
