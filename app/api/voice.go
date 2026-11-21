package api

import (
	"errors"
	"io"
	"net/http"

	"github.com/titpetric/task-ui/feature/voice/stt"
	"github.com/titpetric/task-ui/internal/access"
)

// voiceTranscribe answers with the words of a dictated 16 kHz mono WAV and drops the clip.
func (h *Handlers) voiceTranscribe(w http.ResponseWriter, r *http.Request) {
	if !h.voiceAllowed(w, r) {
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

// voiceReady reports whether the server transcribes, so the orb knows whether to post.
func (h *Handlers) voiceReady(w http.ResponseWriter, r *http.Request) {
	if !h.voiceAllowed(w, r) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ready": h.opts.STT.Ready(r.Context())})
}

// voiceAllowed answers whether the reader may dictate, which feeds a note or a prompt.
func (h *Handlers) voiceAllowed(w http.ResponseWriter, r *http.Request) bool {
	if h.can(r, "note/edit") || h.can(r, "prompt/edit") {
		return true
	}
	fail(w, http.StatusForbidden, "you need the note/edit or prompt/edit permission")
	return false
}

// voiceNote leaves a note on the board with the words and the clip of a dictation, and starts no agent.
func (h *Handlers) voiceNote(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "note/edit") || !h.board(w) {
		return
	}
	var req VoiceNoteRequest
	if err := readJSON(r, &req); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	note, err := h.opts.Board.Record(r.Context(), access.UserFrom(r.Context()), req.ParentID, req.Spoken, req.Audio)
	if err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, note)
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
	case errors.Is(err, stt.ErrBusy):
		return http.StatusTooManyRequests
	}
	return http.StatusInternalServerError
}
