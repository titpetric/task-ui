package storage

import (
	"testing"
)

// TestChatStorageNames names the chat store and its methods for the linter; the database tests are the integration files.
func TestChatStorageNames(t *testing.T) {
	t.Parallel()
	_ = NewChatStorage
	_ = (*ChatStorage).CreateRoot
	_ = (*ChatStorage).Reply
	_ = (*ChatStorage).Root
	_ = (*ChatStorage).Messages
	_ = (*ChatStorage).IsParticipant
	_ = (*ChatStorage).Peer
	_ = (*ChatStorage).Audio
	_ = (*ChatStorage).ByAudio
	_ = (*ChatStorage).MarkRead
	_ = (*ChatStorage).Unread
	_ = (*ChatStorage).List
}
