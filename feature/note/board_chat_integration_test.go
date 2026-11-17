//go:build integration

package note_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/app"
	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/feature/note"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/storage"
)

// A chat is private to its participants, an admin reads it without posting, and the board's own paths refuse one.
func TestChatBoard(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, err := storage.OpenAndMigrate(ctx, ":memory:", app.Migrations())
	assert.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	users := storage.NewUserStorage(db)
	board := &note.Board{Notes: storage.NewNoteStorage(db), Chats: storage.NewChatStorage(db), Users: users}
	alice := &model.User{Name: "alice", Role: model.RoleRunner, Admin: true}
	bob := &model.User{Name: "bob", Role: model.RoleRunner, Admin: true}
	carol := &model.User{Name: "carol", Role: model.RoleViewer}
	dave := &model.User{Name: "dave", Role: model.RoleRunner, Admin: true}
	for _, user := range []*model.User{alice, bob, carol, dave} {
		assert.NoError(t, users.Create(ctx, user))
	}

	root, err := board.ChatMessage(ctx, alice, "bob", "", "hello", "", "")
	assert.NoError(t, err)
	assert.Equal(t, model.ChannelChat, root.Channel)
	assert.Equal(t, alice.ID, root.UserID)

	reply, err := board.ChatMessage(ctx, bob, "", root.ID, "", "spoken", "voice/b.wav")
	assert.NoError(t, err)
	assert.Equal(t, bob.ID, reply.UserID)
	assert.Equal(t, "spoken", reply.Body)

	thread, messages, peer, unread, err := board.ChatThread(ctx, alice, root.ID)
	assert.NoError(t, err)
	assert.Equal(t, root.ID, thread.ID)
	assert.Len(t, messages, 2)
	assert.Equal(t, "bob", peer)
	assert.Equal(t, int64(1), unread)

	assert.NoError(t, board.ChatRead(ctx, alice, root.ID))
	_, _, _, unread, err = board.ChatThread(ctx, alice, root.ID)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), unread)

	_, _, _, _, err = board.ChatThread(ctx, carol, root.ID)
	assert.ErrorIs(t, err, storage.ErrNotFound)
	_, err = board.ChatMessage(ctx, carol, "", root.ID, "hi", "", "")
	assert.ErrorIs(t, err, storage.ErrNotFound)

	_, messages, _, unread, err = board.ChatThread(ctx, dave, root.ID)
	assert.NoError(t, err)
	assert.Len(t, messages, 2)
	assert.Equal(t, int64(0), unread)
	assert.ErrorIs(t, board.ChatRead(ctx, dave, root.ID), storage.ErrNotFound)

	_, err = board.ChatMessage(ctx, alice, "alice", "", "hi", "", "")
	assert.ErrorIs(t, err, note.ErrInvalid)
	_, err = board.ChatMessage(ctx, alice, "nobody", "", "hi", "", "")
	assert.ErrorIs(t, err, note.ErrInvalid)

	_, err = board.Post(ctx, alice, root.ID, "a note")
	assert.ErrorIs(t, err, note.ErrInvalid)
	_, err = board.Record(ctx, alice, root.ID, "a note", "")
	assert.ErrorIs(t, err, note.ErrInvalid)
	_, err = board.Comment(ctx, alice, root.ID, "a note", "")
	assert.ErrorIs(t, err, note.ErrInvalid)
}
