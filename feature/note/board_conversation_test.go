package note

import (
	"testing"
)

// TestBoard_Conversation asserts the read is wired; its rows are an integration test.
func TestBoard_Conversation(t *testing.T) {
	var _ = (*Board).Conversation
}
