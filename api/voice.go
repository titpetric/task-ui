package api

import (
	"errors"
	"io"
	"net/http"

	"github.com/titpetric/task-ui/stt"
)

// voiceTranscribe answers with the words of a dictated clip. The body is
// the audio, 16 kHz mono WAV as the recorder writes it, and the clip is
// read and dropped: nothing of it is stored here. Takes prompt/edit, the
// rule asking Claude takes, because feeding a prompt is all the words
// are for.
func (h *Handlers) voiceTranscribe(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "prompt/edit") {
		return
	}
	if h.opts.STT == nil {
		fail(w, http.StatusServiceUnavailable, stt.ErrOff.Error())
		return
	}
	clip, err := io.ReadAll(http.MaxBytesReader(w, r.Body, stt.MaxClip+1))
	if err != nil {
		fail(w, http.StatusRequestEntityTooLarge, stt.ErrLarge.Error())
		return
	}
	said, err := h.opts.STT.Text(r.Context(), clip)
	if err != nil {
		fail(w, voiceStatusOf(err), err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"text": said})
}

// voiceReady reports whether the server transcribes, so the orb knows
// whether to post what it records. Takes prompt/edit.
func (h *Handlers) voiceReady(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "prompt/edit") {
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ready": h.opts.STT.Ready(r.Context())})
}

// voiceStatusOf maps what the transcriber says to HTTP statuses.
func voiceStatusOf(err error) int {
	switch {
	case errors.Is(err, stt.ErrOff):
		return http.StatusServiceUnavailable
	case errors.Is(err, stt.ErrEmpty):
		return http.StatusBadRequest
	case errors.Is(err, stt.ErrLarge):
		return http.StatusRequestEntityTooLarge
	}
	return http.StatusInternalServerError
}
