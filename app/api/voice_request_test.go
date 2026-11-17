package api_test

import (
	"encoding/json"
	"testing"

	"github.com/titpetric/task-ui/app/api"
	"github.com/titpetric/task-ui/internal/assert"
)

// The orb's body names the thread, the words and the clip.
func TestVoiceNoteRequest(t *testing.T) {
	t.Parallel()
	var req api.VoiceNoteRequest
	assert.NoError(t, json.Unmarshal([]byte(`{"parent_id":"p1","spoken":"hello","audio":"voice/x.wav"}`), &req))
	assert.Equal(t, "p1", req.ParentID)
	assert.Equal(t, "hello", req.Spoken)
	assert.Equal(t, "voice/x.wav", req.Audio)
}
