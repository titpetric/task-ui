package note_test

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/note"
)

// A message with a recording or a transcription was spoken; a typed one
// carries neither.
func TestSpoken_Said(t *testing.T) {
	t.Parallel()
	assert.True(t, note.Spoken{Audio: "voice/2026-11-11.wav"}.Said())
	assert.True(t, note.Spoken{Text: "what runs here"}.Said())
	assert.False(t, note.Spoken{}.Said())
}
