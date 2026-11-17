package api

import (
	"net/http"
	"testing"

	"github.com/titpetric/task-ui/app/model"
)

func TestVoicePath(t *testing.T) {
	for path, want := range map[string]bool{
		"":               true,
		"voice/a.wav":    true,
		"voice/x/y.wav":  true,
		"voice/../a.wav": false,
		"voice//a.wav":   false,
		"notes/a.wav":    false,
	} {
		if got := voicePath(path); got != want {
			t.Errorf("voicePath(%q) = %v, want %v", path, got, want)
		}
	}
}

func TestChatArtifactAllowed(t *testing.T) {
	h := &Handlers{}
	if !h.chatArtifactAllowed(nil, "voice/a.wav") {
		t.Fatal("no board allows every artifact")
	}
	if h.chatMember(nil, &model.Note{ID: "n1"}) {
		t.Fatal("no board allows no chat")
	}
}

func TestThreadHandlers(t *testing.T) {
	var (
		_ = (*Handlers).postThread
		_ = (*Handlers).listThreads
		_ = (*Handlers).getThread
		_ = (*Handlers).readThread
		_ = (*Handlers).threadAudio
		_ = (*Handlers).chatArtifactAllowed
	)
	_ = http.StatusOK
}
