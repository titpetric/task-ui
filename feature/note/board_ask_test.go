package note_test

import (
	"testing"

	"github.com/titpetric/task-ui/feature/note"
)

// TestBoard_AskAs asserts the prompt entry points are wired; their rows are an integration test.
func TestBoard_AskAs(tt *testing.T) {
	var _ = (*note.Board).AskAs
	var _ = (*note.Board).AskArea
}
