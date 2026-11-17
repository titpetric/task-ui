package note

import (
	"context"
	"errors"
	"testing"

	"github.com/titpetric/task-ui/app/model"
)

func TestChatMessageRejects(t *testing.T) {
	b := &Board{}
	if _, err := b.ChatMessage(context.Background(), nil, "bob", "", "hi", "", ""); !errors.Is(err, ErrInvalid) {
		t.Fatalf("nil author: got %v", err)
	}
	user := &model.User{ID: "u1", Name: "alice", Admin: true}
	if _, err := b.ChatMessage(context.Background(), user, "bob", "", "hi", "", ""); !errors.Is(err, ErrChatOff) {
		t.Fatalf("no chat storage: got %v", err)
	}
}

func TestChatParent(t *testing.T) {
	if err := chatParent(nil); err != nil {
		t.Fatalf("nil parent: %v", err)
	}
	if err := chatParent(&model.Note{Channel: model.ChannelNotes}); err != nil {
		t.Fatalf("board note: %v", err)
	}
	if err := chatParent(&model.Note{Channel: model.ChannelChat}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("chat parent: got %v", err)
	}
}

func TestBoardChat(t *testing.T) {
	var (
		_ = (*Board).ChatThread
		_ = (*Board).ChatRead
		_ = (*Board).ChatList
	)
}
