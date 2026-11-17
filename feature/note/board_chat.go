package note

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/storage"
)

// ErrChatOff reports chat on a board without chat storage.
var ErrChatOff = errors.New("chat isn't available on this instance")

// ChatMessage posts a private message: a new thread to a user, or a reply when parentID names its root.
func (b *Board) ChatMessage(ctx context.Context, user *model.User, to, parentID, body, spoken, audio string) (*model.Note, error) {
	if user == nil {
		return nil, fmt.Errorf("%w: a chat needs an author", ErrInvalid)
	}
	if b.Chats == nil || b.Users == nil {
		return nil, ErrChatOff
	}
	if !b.permissions(ctx, user).Allowed("note/edit") {
		return nil, ErrNoteForbidden
	}
	body, spoken, audio = strings.TrimSpace(body), strings.TrimSpace(spoken), strings.TrimSpace(audio)
	if body == "" && spoken == "" && audio == "" {
		return nil, fmt.Errorf("%w: a chat needs a message or a recording", ErrInvalid)
	}
	if body == "" {
		body = spoken
	}
	if body == "" {
		body = untitledRecording
	}
	if parentID != "" {
		return b.chatReply(ctx, user, parentID, body, spoken, audio)
	}
	return b.chatRoot(ctx, user, to, body, spoken, audio)
}

// chatRoot starts a chat thread between two users and returns its first message.
func (b *Board) chatRoot(ctx context.Context, user *model.User, to, body, spoken, audio string) (*model.Note, error) {
	to = strings.TrimSpace(to)
	if to == "" {
		return nil, fmt.Errorf("%w: a chat needs someone to talk to", ErrInvalid)
	}
	if to == user.Name {
		return nil, fmt.Errorf("%w: a chat needs someone other than you", ErrInvalid)
	}
	peer, err := b.Users.GetByName(ctx, to)
	if err != nil {
		return nil, fmt.Errorf("%w: no user %q", ErrInvalid, to)
	}
	if !b.Users.Permissions(ctx, peer).Allowed("note/view") {
		return nil, fmt.Errorf("%w: %s can't see notes", ErrInvalid, to)
	}
	note := &model.Note{
		Channel: model.ChannelChat, UserID: user.ID, UserName: user.Name,
		Body: body, Spoken: spoken, AudioPath: audio,
	}
	participants := []storage.Participant{
		{UserID: user.ID, UserName: user.Name},
		{UserID: peer.ID, UserName: peer.Name},
	}
	if err := b.Chats.CreateRoot(ctx, note, participants); err != nil {
		return nil, err
	}
	b.publish(note)
	return note, nil
}

// chatReply posts a message in a chat thread a participant belongs to.
func (b *Board) chatReply(ctx context.Context, user *model.User, parentID, body, spoken, audio string) (*model.Note, error) {
	root, err := b.Chats.Root(ctx, parentID)
	if err != nil {
		return nil, err
	}
	member, err := b.Chats.IsParticipant(ctx, root.ID, user.ID)
	if err != nil {
		return nil, err
	}
	if !member {
		return nil, storage.ErrNotFound
	}
	note := &model.Note{
		Channel: model.ChannelChat, ParentID: root.ID, UserID: user.ID, UserName: user.Name,
		Body: body, Spoken: spoken, AudioPath: audio,
	}
	if err := b.Chats.Reply(ctx, note, user.ID); err != nil {
		return nil, err
	}
	b.publish(note)
	return note, nil
}

// ChatThread returns a thread's root, its messages and its peer, with the unread count a participant sees.
func (b *Board) ChatThread(ctx context.Context, user *model.User, id string) (*model.Note, []model.Note, string, int64, error) {
	if b.Chats == nil {
		return nil, nil, "", 0, ErrChatOff
	}
	root, err := b.Chats.Root(ctx, id)
	if err != nil {
		return nil, nil, "", 0, err
	}
	member, err := b.chatParticipant(ctx, user, root.ID)
	if err != nil {
		return nil, nil, "", 0, err
	}
	messages, err := b.Chats.Messages(ctx, root.ID)
	if err != nil {
		return nil, nil, "", 0, err
	}
	peer, err := b.Chats.Peer(ctx, root.ID, user.ID)
	if err != nil {
		return nil, nil, "", 0, err
	}
	var unread int64
	if member {
		if unread, err = b.Chats.Unread(ctx, root.ID, user.ID); err != nil {
			return nil, nil, "", 0, err
		}
	}
	return root, messages, peer, unread, nil
}

// ChatRead moves a participant's read marker to now.
func (b *Board) ChatRead(ctx context.Context, user *model.User, id string) error {
	if b.Chats == nil {
		return ErrChatOff
	}
	root, err := b.Chats.Root(ctx, id)
	if err != nil {
		return err
	}
	member, err := b.chatParticipant(ctx, user, root.ID)
	if err != nil {
		return err
	}
	if !member {
		return storage.ErrNotFound
	}
	return b.Chats.MarkRead(ctx, root.ID, user.ID)
}

// ChatList returns the chat threads a user participates in, newest message first.
func (b *Board) ChatList(ctx context.Context, user *model.User, limit int) ([]storage.ChatSummary, error) {
	if b.Chats == nil {
		return nil, ErrChatOff
	}
	if user == nil {
		return nil, storage.ErrNotFound
	}
	return b.Chats.List(ctx, user.ID, limit)
}

// chatParticipant reports whether a user participates in a chat thread, an admin reading but not posting.
func (b *Board) chatParticipant(ctx context.Context, user *model.User, id string) (bool, error) {
	if user == nil {
		return false, storage.ErrNotFound
	}
	member, err := b.Chats.IsParticipant(ctx, id, user.ID)
	if err != nil {
		return false, err
	}
	if !member && !user.CanAdmin() {
		return false, storage.ErrNotFound
	}
	return member, nil
}

// chatParent reports a message that belongs to a chat thread, which the board's own paths refuse.
func chatParent(note *model.Note) error {
	if note != nil && note.Channel == model.ChannelChat {
		return fmt.Errorf("%w: a chat message is not a board note", ErrInvalid)
	}
	return nil
}
