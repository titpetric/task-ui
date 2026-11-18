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
	"github.com/titpetric/task-ui/feature/note"
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

// listThreads returns the user's chats and, with ?threads=1 or ?live=1, the prompt and board feed.
func (h *Handlers) listThreads(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	feed := q.Get("threads") == "1" || q.Get("live") == "1"
	notes, prompts := h.can(r, "note/view"), h.can(r, "prompt/view")
	if !h.board(w) {
		return
	}
	if !feed && !h.require(w, r, "note/view") {
		return
	}
	if feed && !notes && !prompts {
		fail(w, http.StatusForbidden, "you need the note/view or prompt/view permission to read threads")
		return
	}
	out := map[string]any{"unread": int64(0), "chats": []storage.ChatSummary{}}
	if feed {
		conversations, err := h.opts.Board.Notes.Feed(r.Context(), storage.ConversationFilter{
			WorkspaceID: q.Get("workspace"),
			Folder:      strings.Trim(q.Get("folder"), "/"),
			UserName:    anyMatch(q.Get("user")),
			State:       anyMatch(q.Get("state")),
			Query:       strings.TrimSpace(q.Get("q")),
			Live:        q.Get("live") == "1",
			Limit:       queryInt(r, "limit", 50),
		})
		if err != nil {
			failErr(w, err)
			return
		}
		user := access.UserFrom(r.Context())
		kept := conversations[:0]
		for _, row := range conversations {
			if row.IsPrompt() {
				if prompts && seesPlace(h.opts.Catalogue, user, row.Dir) {
					kept = append(kept, row)
				}
				continue
			}
			if notes {
				kept = append(kept, row)
			}
		}
		out["threads"] = kept
	}
	if !feed || notes {
		chats, err := h.opts.Board.ChatList(r.Context(), access.UserFrom(r.Context()), queryInt(r, "limit", 50))
		if err != nil {
			failErr(w, err)
			return
		}
		if chats != nil {
			out["chats"] = chats
		}
		var unread int64
		for _, chat := range chats {
			unread += chat.Unread
		}
		out["unread"] = unread
	}
	writeJSON(w, http.StatusOK, out)
}

// getThread returns a chat with its messages, or a prompt or note with its work.
func (h *Handlers) getThread(w http.ResponseWriter, r *http.Request) {
	if !h.board(w) {
		return
	}
	id := param(r, "id")
	thread, err := h.opts.Board.Notes.Get(r.Context(), id)
	if err != nil {
		failErr(w, err)
		return
	}
	if thread.Channel != model.ChannelChat {
		if !h.threadVisible(r, thread) {
			failErr(w, storage.ErrNotFound)
			return
		}
		conversation, err := h.opts.Board.Conversation(r.Context(), id)
		if err != nil {
			failErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, conversation)
		return
	}
	if !h.require(w, r, "note/view") {
		return
	}
	thread, messages, peer, unread, err := h.opts.Board.ChatThread(r.Context(), access.UserFrom(r.Context()), id)
	if err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"thread": thread, "with": peer, "unread": unread, "messages": messages})
}

// setThreadAgent changes the agent a thread's runs answer as.
func (h *Handlers) setThreadAgent(w http.ResponseWriter, r *http.Request) {
	if !h.board(w) {
		return
	}
	id := param(r, "id")
	thread, err := h.opts.Board.Notes.Get(r.Context(), id)
	if err != nil {
		failErr(w, err)
		return
	}
	if thread.ParentID != "" || thread.Channel == model.ChannelChat || !h.threadVisible(r, thread) {
		failErr(w, storage.ErrNotFound)
		return
	}
	user := access.UserFrom(r.Context())
	if thread.UserName != user.Name && !user.CanAdmin() {
		fail(w, http.StatusForbidden, "only the thread's author or an admin picks its agent")
		return
	}
	if thread.State == model.NoteWorking {
		fail(w, http.StatusConflict, "the thread is working")
		return
	}
	var req struct {
		Agent string `json:"agent"`
	}
	if err := readJSON(r, &req); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	agent := strings.TrimSpace(req.Agent)
	if agent == "" {
		fail(w, http.StatusBadRequest, "the thread needs an agent")
		return
	}
	if h.opts.Board.Users == nil {
		fail(w, http.StatusBadRequest, "no agent "+agent)
		return
	}
	named, err := h.opts.Board.Users.GetByName(r.Context(), agent)
	if err != nil || !named.AgentReply || !named.Mcp {
		fail(w, http.StatusBadRequest, "no agent "+agent)
		return
	}
	if err := h.opts.Board.Notes.SetAgent(r.Context(), id, agent); err != nil {
		failErr(w, err)
		return
	}
	thread.Agent = agent
	if h.opts.Events != nil {
		h.opts.Events.Publish(note.ThreadTopic(id), "note", *thread)
	}
	writeJSON(w, http.StatusOK, map[string]string{"agent": agent})
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
