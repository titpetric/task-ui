package api

import (
	"errors"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/feature/artifacts"
	"github.com/titpetric/task-ui/internal/access"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/storage"
)

// postThread posts a private message, a new thread to a user or a reply in one.
func (h *Handlers) postThread(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "note/edit") || !h.require(w, r, "note/view") || !h.board(w) {
		return
	}
	var req ThreadRequest
	if err := readJSON(r, &req); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	if !voicePath(req.Audio) {
		fail(w, http.StatusBadRequest, "a recording goes under voice/")
		return
	}
	note, err := h.opts.Board.ChatMessage(r.Context(), access.UserFrom(r.Context()), req.To, req.ParentID, req.Body, req.Spoken, req.Audio)
	if err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, note)
}

// listThreads returns the chat threads of the user, newest message first, and their unread count.
func (h *Handlers) listThreads(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "note/view") || !h.board(w) {
		return
	}
	chats, err := h.opts.Board.ChatList(r.Context(), access.UserFrom(r.Context()), queryInt(r, "limit", 50))
	if err != nil {
		failErr(w, err)
		return
	}
	var unread int64
	for _, chat := range chats {
		unread += chat.Unread
	}
	writeJSON(w, http.StatusOK, map[string]any{"unread": unread, "chats": chats})
}

// getThread returns a chat thread with its messages.
func (h *Handlers) getThread(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "note/view") || !h.board(w) {
		return
	}
	thread, messages, peer, unread, err := h.opts.Board.ChatThread(r.Context(), access.UserFrom(r.Context()), param(r, "id"))
	if err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"thread": thread, "with": peer, "unread": unread, "messages": messages})
}

// readThread moves the user's read marker in a chat thread to now.
func (h *Handlers) readThread(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "note/view") || !h.board(w) {
		return
	}
	if err := h.opts.Board.ChatRead(r.Context(), access.UserFrom(r.Context()), param(r, "id")); err != nil {
		failErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// threadAudio streams a chat message's recording to a participant.
func (h *Handlers) threadAudio(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "note/view") || !h.board(w) {
		return
	}
	if h.opts.ArtifactStore == nil || h.opts.Board.Chats == nil {
		fail(w, http.StatusNotFound, artifactstore.ErrNotFound.Error())
		return
	}
	note, err := h.opts.Board.Chats.Audio(r.Context(), param(r, "id"))
	if err != nil {
		failErr(w, err)
		return
	}
	if !h.chatMember(r, note) || !voicePath(note.AudioPath) || note.AudioPath == "" {
		fail(w, http.StatusNotFound, artifactstore.ErrNotFound.Error())
		return
	}
	file, err := h.opts.ArtifactStore.Path(note.AudioPath)
	if err != nil {
		fail(w, http.StatusNotFound, artifactstore.ErrNotFound.Error())
		return
	}
	clip, err := os.Open(file)
	if err != nil {
		fail(w, http.StatusNotFound, artifactstore.ErrNotFound.Error())
		return
	}
	defer clip.Close()
	streamFile(w, r, note.AudioPath, runner.ContentType(note.AudioPath), time.Time{}, clip)
}

// voicePath reports whether a recording path lies under voice/ and needs no cleaning.
func voicePath(p string) bool {
	return p == "" || strings.HasPrefix(p, "voice/") && path.Clean(p) == p
}

// chatMember reports whether the request's user may read a chat message.
func (h *Handlers) chatMember(r *http.Request, note *model.Note) bool {
	if h.opts.Board == nil || h.opts.Board.Chats == nil {
		return false
	}
	root := note.ID
	if note.ParentID != "" {
		root = note.ParentID
	}
	user := access.UserFrom(r.Context())
	member, err := h.opts.Board.Chats.IsParticipant(r.Context(), root, user.ID)
	if err != nil {
		return false
	}
	return member || user.CanAdmin()
}

// chatArtifactAllowed reports whether a recording at path may be read: a chat's clip is its participants' alone.
func (h *Handlers) chatArtifactAllowed(r *http.Request, path string) bool {
	if h.opts.Board == nil || h.opts.Board.Chats == nil {
		return true
	}
	note, err := h.opts.Board.Chats.ByAudio(r.Context(), path)
	if err != nil {
		return errors.Is(err, storage.ErrNotFound)
	}
	return h.chatMember(r, note)
}
