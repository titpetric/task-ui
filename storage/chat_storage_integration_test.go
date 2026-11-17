//go:build integration

package storage_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/storage"
)

// A chat thread keeps its participants, counts what each of them has not read, and plays its recordings by message.
func TestChatStorage(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	chats := storage.NewChatStorage(newDB(t))
	root := &model.Note{Channel: model.ChannelChat, UserID: "u1", UserName: "alice", Body: "hello", AudioPath: "voice/a.wav"}
	assert.NoError(t, chats.CreateRoot(ctx, root, []storage.Participant{
		{UserID: "u1", UserName: "alice"},
		{UserID: "u2", UserName: "bob"},
	}))
	assert.True(t, root.ID != "")

	got, err := chats.Root(ctx, root.ID)
	assert.NoError(t, err)
	assert.Equal(t, root.ID, got.ID)

	member, err := chats.IsParticipant(ctx, root.ID, "u1")
	assert.NoError(t, err)
	assert.True(t, member)
	member, err = chats.IsParticipant(ctx, root.ID, "u3")
	assert.NoError(t, err)
	assert.False(t, member)

	peer, err := chats.Peer(ctx, root.ID, "u1")
	assert.NoError(t, err)
	assert.Equal(t, "bob", peer)

	unread, err := chats.Unread(ctx, root.ID, "u2")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), unread)

	list, err := chats.List(ctx, "u2", 10)
	assert.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, "alice", list[0].Peer)
	assert.Equal(t, "hello", list[0].LastBody)
	assert.Equal(t, int64(1), list[0].Unread)

	reply := &model.Note{Channel: model.ChannelChat, ParentID: root.ID, UserID: "u2", UserName: "bob", Body: "hi", AudioPath: "voice/b.wav"}
	assert.NoError(t, chats.Reply(ctx, reply, "u2"))
	assert.True(t, reply.ID != "")

	unread, err = chats.Unread(ctx, root.ID, "u1")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), unread)

	messages, err := chats.Messages(ctx, root.ID)
	assert.NoError(t, err)
	assert.Len(t, messages, 2)
	assert.Equal(t, root.ID, messages[0].ID)
	assert.Equal(t, reply.ID, messages[1].ID)

	assert.NoError(t, chats.MarkRead(ctx, root.ID, "u1"))
	unread, err = chats.Unread(ctx, root.ID, "u1")
	assert.NoError(t, err)
	assert.Equal(t, int64(0), unread)

	audio, err := chats.Audio(ctx, reply.ID)
	assert.NoError(t, err)
	assert.Equal(t, "voice/b.wav", audio.AudioPath)
	byPath, err := chats.ByAudio(ctx, "voice/b.wav")
	assert.NoError(t, err)
	assert.Equal(t, reply.ID, byPath.ID)
	_, err = chats.ByAudio(ctx, "voice/none.wav")
	assert.ErrorIs(t, err, storage.ErrNotFound)
}
